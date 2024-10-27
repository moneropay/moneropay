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
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"

	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

type recv struct {
	index, expected, received, creationHeight uint64
	description, callbackUrl                  string
	createdAt                                 time.Time
	updated                                   bool
}

var (
	// Height of the transaction that last resulted in a callback
	lastCallbackHeight uint64

	// Last reported height by wallet-rpc
	lastSeenHeight uint64
)

func readLastCallbackHeight(ctx context.Context) {
	row := db.QueryRowContext(ctx, "SELECT height FROM last_block_height")
	if err := row.Scan(&lastCallbackHeight); err != nil {
		log.Fatal().Err(err).Msg("Failed to read last callback height")
	}
}

func saveLastCallbackHeight(ctx context.Context) error {
	_, err := db.ExecContext(ctx, "UPDATE last_block_height SET height=$1",
		lastCallbackHeight)
	return err
}

func sendCallbackRequest(d model.CallbackResponse, u string) error {
	j, _ := json.Marshal(d)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoneroPay/"+Version)
	c := &http.Client{Timeout: 30 * time.Second}
	_, err = c.Do(req)
	return err
}

func callback(ctx context.Context, r *recv, t *walletrpc.Transfer, locked bool) error {
	resp, err := Balance(ctx, []uint64{r.index})
	if err != nil {
		return err
	}
	// Prepare a callback json payload.
	var d model.CallbackResponse
	d.Amount.Expected = r.expected
	d.Amount.Covered.Total = r.received + (resp.PerSubaddress[0].Balance -
		resp.PerSubaddress[0].UnlockedBalance)
	d.Amount.Covered.Unlocked = r.received
	d.Complete = d.Amount.Covered.Unlocked >= d.Amount.Expected
	d.Description = r.description
	d.CreatedAt = r.createdAt
	d.Transaction = model.TransactionData{
		Amount:          t.Amount,
		Confirmations:   t.Confirmations,
		DoubleSpendSeen: t.DoubleSpendSeen,
		Fee:             t.Fee,
		Height:          t.Height,
		Timestamp:       time.Unix(int64(t.Timestamp), 0),
		TxHash:          t.Txid,
		UnlockTime:      t.UnlockTime,
		Locked:          locked,
	}
	return sendCallbackRequest(d, r.callbackUrl)
}

func findMinCreationHeight(rs map[uint64]*recv) uint64 {
	var h uint64
	for _, r := range rs {
		h = r.creationHeight
		break
	}
	for _, r := range rs {
		if r.creationHeight < h {
			h = r.creationHeight
		}
	}
	return h
}

func updateReceivers(ctx context.Context, rs map[uint64]*recv) {
	for _, r := range rs {
		if !r.updated {
			continue
		}
		if _, err := db.ExecContext(ctx,
			"UPDATE receivers SET received_amount=$1 WHERE subaddress_index=$2",
			r.received, r.index); err != nil {
			log.Error().Err(err).Uint64("address_index", r.index).
				Msg("Failed to update payment request")
		}
	}
}

func makeRecvMap(rows *sql.Rows) map[uint64]*recv {
	rs := make(map[uint64]*recv)
	for rows.Next() {
		var t recv
		if err := rows.Scan(&t.index, &t.expected, &t.received, &t.description, &t.callbackUrl,
			&t.createdAt, &t.creationHeight); err != nil {
			log.Error().Err(err).Msg("Failed to get payment requests from database")
		}
		rs[t.index] = &t
	}
	return rs
}

func checkTransfers() {
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, "SELECT subaddress_index,expected_amount,received_amount,description,"+
		"callback_url,created_at,creation_height FROM receivers")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get payment requests from database")
		return
	}
	defer rows.Close()

	rs := makeRecvMap(rows)
	if len(rs) == 0 {
		return
	}
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In:             true,
		FilterByHeight: true,
		// If there are very old rows and they aren't removed, there can be
		// performance issues
		MinHeight: findMinCreationHeight(rs),
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get incoming transfers")
		return
	}
	if resp.In == nil {
		return
	}
	maxHeight := lastCallbackHeight
	for _, t := range resp.In {
		locked, eventHeight := getTransferLockStatus(t)
		if eventHeight <= lastCallbackHeight {
			continue
		}
		if r, ok := rs[t.SubaddrIndex.Minor]; ok {
			if !locked {
				r.received += t.Amount
				r.updated = true
			}
			if r.callbackUrl != "" {
				if err = callback(ctx, r, &t, locked); err != nil {
					log.Error().Err(err).Uint64("address_index", t.SubaddrIndex.Minor).
						Uint64("amount", t.Amount).Str("tx_id", t.Txid).
						Uint64("event_height", eventHeight).Bool("locked", locked).
						Bool("pool", false).Uint64("confirmations", t.Confirmations).
						Msg("Failed callback")
				} else {
					log.Info().Uint64("address_index", t.SubaddrIndex.Minor).
						Uint64("amount", t.Amount).Str("tx_id", t.Txid).
						Uint64("event_height", eventHeight).Bool("locked", locked).
						Bool("pool", false).Uint64("confirmations", t.Confirmations).
						Msg("Sent callback")
				}
			}
			// Don't depend on wallet-rpc's ordering of transfers
			if eventHeight > maxHeight {
				maxHeight = eventHeight
			}
		}
	}
	if maxHeight == lastCallbackHeight {
		return
	}
	lastCallbackHeight = maxHeight
	if err := saveLastCallbackHeight(ctx); err != nil {
		log.Error().Err(err).Uint64("height", lastCallbackHeight).Msg("Failed to save last callback height")
	} else {
		log.Info().Uint64("height", lastCallbackHeight).Msg("Saved last callback height")
	}
	updateReceivers(ctx, rs)
}

func checkMempool() {
	ctx := context.Background()
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		Pool: true,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get mempool transfers")
		return
	}

	rows, err := db.QueryContext(ctx, "SELECT txid FROM mempool_seen")
	if err != nil {
		log.Err(err).Msg("Failed to query mempool cache in database")
		return
	}
	defer rows.Close()

	poolSeen := make(map[string]struct{})
	for rows.Next() {
		var txid string
		if err := rows.Scan(&txid); err != nil {
			log.Err(err).Msg("Failed to scan txid from mempool_seen table")
			continue
		}
		poolSeen[txid] = struct{}{}
	}

	// Some pool transactions are cached in DB but the pool is empty. Purge cached transactions in DB.
	if resp.Pool == nil && len(poolSeen) != 0 {
		if _, err := db.ExecContext(ctx, "DELETE FROM mempool_seen"); err != nil {
			log.Err(err).Msg("Failed to purge mempool cache in database")
		}
		return
	}

	var addressIndices []uint64
	for _, p := range resp.Pool {
		if _, ok := poolSeen[p.Txid]; ok {
			// This pool tx was processed before
			continue
		}
		addressIndices = append(addressIndices, p.SubaddrIndex.Minor)
	}

	if Config.sqliteCS != "" {
		// Less efficient than the PostgreSQL query with ANY but SQLite is not meant for production use
		rows, err = db.QueryContext(ctx, "SELECT subaddress_index,expected_amount,received_amount,description,"+
			"callback_url,created_at,creation_height FROM receivers WHERE subaddress_index")
	} else {
		rows, err = db.QueryContext(ctx, "SELECT subaddress_index,expected_amount,received_amount,description,"+
			"callback_url,created_at,creation_height FROM receivers WHERE subaddress_index = ANY($1)", addressIndices)
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get payment requests from database")
		return
	}
	defer rows.Close()
	rs := makeRecvMap(rows)
	if len(rs) == 0 {
		return
	}

	// Send callbacks for new mempool tx, later on mark them as seen to prevent duplicate callbacks
	for _, t := range resp.Pool {
		if _, ok := poolSeen[t.Txid]; ok {
			// This pool tx was processed before
			continue
		}
		if r, ok := rs[t.SubaddrIndex.Minor]; ok {
			if r.callbackUrl != "" {
				if err = callback(ctx, r, &t, true); err != nil {
					log.Error().Err(err).Uint64("address_index", t.SubaddrIndex.Minor).
						Uint64("amount", t.Amount).Str("tx_id", t.Txid).
						Bool("locked", true).Bool("pool", true).
						Uint64("confirmations", t.Confirmations).
						Msg("Failed callback")
				} else {
					log.Info().Uint64("address_index", t.SubaddrIndex.Minor).
						Uint64("amount", t.Amount).Str("tx_id", t.Txid).
						Bool("locked", false).Bool("pool", true).
						Uint64("confirmations", t.Confirmations).
						Msg("Sent callback")
				}
			}
		}
		if _, err := db.ExecContext(ctx, "INSERT INTO mempool_seen (txid) VALUES ($1)", t.Txid); err != nil {
			log.Err(err).Msg("Failed to save txid into mempool cache in database")
		}
	}

	// Delete database mempool cache tx that no longer exist in mempool.
	var toDelete []string
	for txid := range poolSeen {
		found := false
		for _, tx := range resp.Pool {
			if tx.Txid == txid {
				found = true
			}
		}
		if !found {
			toDelete = append(toDelete, txid)
		}
	}

	if Config.sqliteCS != "" {
		for _, d := range toDelete {
			if _, err := db.ExecContext(ctx, "DELETE FROM mempool_seen WHERE txid=$1", d); err != nil {
				log.Err(err).Str("txid", d).Msg("Failed to delete old mempool_seen cache entry")
			}
		}
	} else {
		if _, err := db.ExecContext(ctx, "DELETE FROM mempool_seen WHERE txid = ANY($1)", toDelete); err != nil {
			log.Err(err).Msg("Failed to delete old mempool_seen cache entries")
		}
	}
}

func callbackRunner() {
	for {
		if Config.zeroConf {
			checkMempool()
		}
		heightResp, err := getHeight(context.Background())
		if err != nil {
			log.Err(err).Msg("Failed to get height from wallet-rpc")
		} else {
			// If there was a new block, see if there is anything to callback
			if heightResp.Height > lastSeenHeight {
				checkTransfers()
				lastSeenHeight = heightResp.Height
			}
		}
		time.Sleep(5 * time.Second)
	}
}
