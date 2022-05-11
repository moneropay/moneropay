/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2022 Laurynas Četyrkinas <stnby@kernal.eu>
 * Copyright (C) 2022 İrem Kuyucu <siren@kernal.eu>
 *
 * MoneroPay is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * MoneroPay is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with MoneroPay.  If not, see <https://www.gnu.org/licenses/>.
 */

package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gitlab.com/moneropay/go-monero/walletrpc"
)

type additionalCallbackData struct {
	ExpectedAmount uint64
	Description string
	CallbackUrl string
	CreatedAt time.Time
	Balance uint64
	UnlockedBalance uint64
}

type ReceiveTransaction struct {
	Amount uint64 `json:"amount"`
	Confirmations uint64 `json:"confirmations"`
	DoubleSpendSeen bool `json:"double_spend_seen"`
	Fee uint64 `json:"fee"`
	Height uint64 `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	TxHash string `json:"tx_hash"`
	UnlockTime uint64 `json:"unlock_time"`
}

type callbackRequest struct {
	Amount struct {
		Expected uint64 `json:"expected"`
		Covered struct {
			Total uint64 `json:"total"`
			Unlocked uint64 `json:"unlocked"`
		} `json:"covered"`
	} `json:"amount"`
	Complete bool `json:"complete"`
	Description string `json:"description,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Transaction ReceiveTransaction `json:"transaction"`
}

var callbackLastHeight uint64
var lastHeightUpdated bool

func readCallbackLastHeight(ctx context.Context) {
	c := make(chan error)
	go func() {
		row := pdb.QueryRow(ctx, "SELECT height FROM last_block_height")
		c <- row.Scan(&callbackLastHeight)
	}()
	select {
		case <-ctx.Done(): log.Fatal(ctx.Err())
		case err := <-c: log.Fatal(err)
	}
}

func saveCallbackLastHeight(ctx context.Context) error {
	c := make(chan error)
	go func() {
		_, err := pdb.Exec(ctx, "UPDATE last_block_height SET height = $1", callbackLastHeight)
		c <- err
	}()
	select {
		case <-ctx.Done(): return ctx.Err()
		case err := <-c: return err
	}
}

func mapAditionalCallbackData(indices []uint64) (map[uint64]additionalCallbackData, error) {
	ctx, c1 := context.WithTimeout(context.Background(), 30 * time.Second)
	defer c1()
	resp, err := Balance(ctx, indices)
	if err != nil {
		return nil, err
	}
	ctx, c2 := context.WithTimeout(context.Background(), 30 * time.Second)
	defer c2()
	type ret struct {resp map[uint64]additionalCallbackData; err error}
	c := make(chan ret)
	go func() {
		rows, err := pdb.Query(ctx, "SELECT subaddress_index, expected_amount, description," +
		    " callback_url, created_at FROM receivers WHERE subaddress_index = ANY ($1) AND" +
		    " callback_url != ''", indices)
		if err != nil {
			c <- ret{nil, err}
			return
		}
		// Map subaddress_index table to additionalCallbackData.
		m := make(map[uint64]additionalCallbackData)
		var si, ea uint64
		var d, cu string
		var ca time.Time
		for rows.Next() {
			if err = rows.Scan(&si, &ea, &d, &cu, &ca); err != nil {
				c <- ret{nil, err}
				return
			}
			m[si] = additionalCallbackData{
				ExpectedAmount: ea,
				Description: d,
				CallbackUrl: cu,
				CreatedAt: ca,
			}
		}
		// Map subaddress balance to additionalCallbackData.
		for _, b := range resp.PerSubaddress {
			if e, ok := m[b.AddressIndex]; ok {
				e.Balance = b.Balance
				e.UnlockedBalance = b.UnlockedBalance
				m[b.AddressIndex] = e
			}
		}
		c <- ret{m, nil}
	}()
	select {
		case <-ctx.Done(): return nil, ctx.Err()
		case r := <-c: return r.resp, r.err
	}
}

func sendCallback(url string, data callbackRequest) error {
	j, _ := json.Marshal(data)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoneroPay/" + Version)
	c := &http.Client{Timeout: time.Second * 3}
	if _, err := c.Do(req); err != nil {
		return err
	}
	return nil
}

// Fetch new transfers from wallet-rpc, find their 'callback_url' in DB
// and send a callback.
func fetchTransfers() {
	// TODO: what should we do with the below duration?
	ctx, c1 := context.WithTimeout(context.Background(), 30 * time.Second)
	defer c1()
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In: true,
		FilterByHeight: true,
		MinHeight: callbackLastHeight,
	})
	if err != nil {
		log.Println(err)
		return
	}
	if resp.In == nil {
		return
	}
	// Create a list of subaddress indices.
	i := []uint64{}
	for _, t := range resp.In {
		i = append(i, t.SubaddrIndex.Minor)
	}
	m, err := mapAditionalCallbackData(i)
	if err != nil {
		log.Println(err)
		return
	}
	lastHeightUpdated = false
	// Check if any of the transfers were made to receivers
	for _, t := range resp.In {
		if t.Height > callbackLastHeight {
			callbackLastHeight = t.Height
			lastHeightUpdated = true
		}
		if e, ok := m[t.SubaddrIndex.Minor]; ok {
			// Prepare a callback json payload.
			var d callbackRequest
			d.Amount.Expected = e.ExpectedAmount
			d.Amount.Covered.Total = e.Balance
			d.Amount.Covered.Unlocked = e.UnlockedBalance
			d.Complete = d.Amount.Covered.Unlocked >= d.Amount.Expected
			d.Description = e.Description
			d.CreatedAt = e.CreatedAt
			d.Transaction = ReceiveTransaction{
				Amount: t.Amount,
				Confirmations: t.Confirmations,
				DoubleSpendSeen: t.DoubleSpendSeen,
				Fee: t.Fee,
				Height: t.Height,
				Timestamp: time.Unix(int64(t.Timestamp), 0),
				TxHash: t.Txid,
				UnlockTime: t.UnlockTime,
			}
			if err = sendCallback(e.CallbackUrl, d); err != nil {
				log.Println(err)
			}
		}
	}
	if !lastHeightUpdated {
		return
	}
	ctx, c2 := context.WithTimeout(context.Background(), 30 * time.Second)
	defer c2()
	if saveCallbackLastHeight(ctx) != nil {
		log.Println(err)
	}
}

func callbackRunner() {
	// Check for new incoming transfers and send out a callback payload.
	for {
		fetchTransfers()
		time.Sleep(30 * time.Second)
	}
}
