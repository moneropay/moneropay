/*
 * Copyright (C) 2021 Laurynas Četyrkinas <stnby@kernal.eu>
 * Copyright (C) 2021 İrem Kuyucu <siren@kernal.eu>
 *
 * This file is part of MoneroPay.
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

package callback

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gitlab.com/moneropay/go-monero/walletrpc"

	"gitlab.com/moneropay/moneropay/pkg/models"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/config"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/database"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/wallet"
)

type callbackDest struct {
	Url string
	Description string
}

func doCallback(url, payload string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		// Callback url can't be parsed.
		// This can't be resolved in the future, don't retry.
		log.Println(err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoneroPay/1.0.0")
	c := &http.Client{Timeout: time.Second * 3}
	if _, err := c.Do(req); err != nil {
		return err
	}
	return nil
}

func retryFailedCallbacks() {
	// Get scheduled failed callbacks from DB.
	rows, err := database.QueryWithTimeout(context.Background(), 5 * time.Second,
	    "SELECT uid, callback_url,request_body, attempts " +
	    "FROM receivers, failed_callbacks " +
	    "WHERE $1 > next_retry " +
	    "AND receivers.subaddress_index = failed_callbacks.subaddress_index", time.Now())
	if err != nil {
		log.Println(err)
		return
	}
	l := len(config.Values.Intervals)
	var url, body string
	var attempts, uid int
	for rows.Next() {
		rows.Scan(&uid, &url, &body, &attempts)
		// Send the transfer data to the callback url and remove from DB if successful,
		// increment 'attempts' and schedule it again in case of a failure.
		if err = doCallback(url, body); err == nil || attempts == l {
			// Remove from DB if failed too many times as well.
			log.Printf("Removing callback, uid = %v, attempts = %v.", uid, attempts)
			if err = database.ExecWithTimeout(context.Background(), 2 * time.Second,
			    "DELETE FROM failed_callbacks WHERE uid = $1", uid); err != nil {
				log.Println(err)
			}
			continue
		}
		// Next scheduled timestamp depends on failed 'attempts'.
		log.Printf("Rescheduling failed callback, uid = %v, attempts = %v.", uid, attempts)
		if err = database.ExecWithTimeout(context.Background(), 2 * time.Second,
		    "UPDATE failed_callbacks SET attempts = attempts + 1, next_retry = $1 " +
		    "WHERE uid = $2",
		    time.Now().Add(config.Values.Intervals[attempts]), uid); err != nil {
			log.Println(err)
		}
	}
}

// Fetch new transfers from wallet-rpc, find their 'callback_url' in DB.
// Attempt a callback, on failure add the callback to 'failed_callbacks' table.
func fetchTransfers(h *uint64) {
	// Get last checked block height from DB.
	// Get new transfers from the wallet-rpc.
	wallet.Lock()
	resp, err := wallet.Wallet.GetTransfers(&walletrpc.GetTransfersRequest{
		In: true,
		FilterByHeight: true,
		MinHeight: *h,
	})
	wallet.Unlock()
	if err != nil {
		log.Println(err)
		return
	}
	if resp.In == nil {
		return
	}

	// Create a list of subaddress indices.
	a := []uint64{}
	for _, t := range resp.In {
		a = append(a, t.SubaddrIndex.Minor)
	}

	// Get additional data for callback.
	rows, err := database.QueryWithTimeout(context.Background(), 3 * time.Second,
	    "SELECT subaddress_index, description, callback_url " +
	    "FROM receivers WHERE subaddress_index = ANY ($1)", a)
	if err != nil {
		log.Println(err)
		return
	}

	// Map subaddress_index to callbackDest{url, description}.
	m := make(map[uint64]callbackDest)
	var i uint64
	var u, d string
	for rows.Next() {
		if err = rows.Scan(&i, &d, &u); err != nil {
			continue
		}
		m[i] = callbackDest{u, d}
	}

	// Check if any of the transfers were made to
	for _, t := range resp.In {
		if t.Height > *h {
			*h = t.Height
		}

		d, ok := m[t.SubaddrIndex.Minor]
		if !ok {
			continue
		}

		// Prepare a callback json payload.
		j, _ := json.Marshal(
		    models.CallbackData{
			Amount: t.Amount,
			Description: d.Description,
			Fee: t.Fee,
			Timestamp: time.Unix(int64(t.Timestamp), 0),
			Address: t.Address,
			Height: t.Height,
			TxHash: t.Txid,
			Confirmations: t.Confirmations,
			DoubleSpendSeen: t.DoubleSpendSeen,
		})
		b := string(j)
		if err = doCallback(d.Url, b); err != nil {
			// FIXME: If database is inaccessible within 5 seconds,
			// the failed callback will be lost.
			// For HA setups we recommend using a connection pooler.
			err = database.ExecWithTimeout(context.Background(), 5 * time.Second,
			    "INSERT INTO failed_callbacks (subaddress_index, request_body, next_retry) " +
			    "VALUES ($1, $2, $3)", t.SubaddrIndex.Minor, b,
			    time.Now().Add(config.Values.Intervals[0]))
			if err != nil {
				log.Println(err)
			}
		}
	}
	err = database.ExecWithTimeout(context.Background(), 2 * time.Second,
	    "UPDATE metadata SET value = $1 WHERE key = 'last_height'", *h)
	if err != nil {
		log.Println(err)
	}
}

func Run() {
	// Get the previously checked block height,
	// so we don't reprocess old transfers.
	row, err := database.QueryRowWithTimeout(context.Background(), 2 * time.Second,
	    "SELECT value FROM metadata WHERE key = 'last_height'")
	if err != nil {
		log.Println(err)
		return
	}
	var h uint64
	if err = row.Scan(&h); err != nil {
		log.Fatal(err)
	}

	// Check for new incoming transfers and send out a callback payload.
	go func() {
		for {
			fetchTransfers(&h)
			time.Sleep(30 * time.Second)
		}
	}()

	// Retry previously failed callbacks.
	for {
		retryFailedCallbacks()
		time.Sleep(30 * time.Second)
	}
}
