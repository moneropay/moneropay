/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
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

// recv represents a payment receiver/request from the database.
type recv struct {
	index, expected, received, creationHeight uint64
	address, description, callbackUrl         string
	createdAt                                 time.Time
	updated                                   bool
}

// readLastCallbackHeight loads the last callback height from the database.
func (d *Daemon) readLastCallbackHeight(ctx context.Context) {
	row := d.db.QueryRowContext(ctx, "SELECT height FROM last_block_height")
	if err := row.Scan(&d.lastCallbackHeight); err != nil {
		log.Fatal().Err(err).Msg("Failed to read last callback height")
	}
}

// saveLastCallbackHeight persists the last callback height to the database.
func (d *Daemon) saveLastCallbackHeight(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, "UPDATE last_block_height SET height=$1",
		d.lastCallbackHeight)
	return err
}

// sendCallbackRequest sends a POST request with the callback payload to the given URL.
func sendCallbackRequest(data model.CallbackResponse, url, address string) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MoneroPay/"+Version)
	req.Header.Set("X-MoneroPay-Address", address)
	c := &http.Client{Timeout: 30 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// sendPaymentCallback prepares and sends a callback for a payment transfer.
func (d *Daemon) sendPaymentCallback(ctx context.Context, r *recv, t *walletrpc.Transfer, locked bool) error {
	resp, err := d.Balance(ctx, []uint64{r.index})
	if err != nil {
		return err
	}

	// Prepare callback payload
	var data model.CallbackResponse
	data.Amount.Expected = r.expected
	data.Amount.Covered.Total = r.received + (resp.PerSubaddress[0].Balance -
		resp.PerSubaddress[0].UnlockedBalance)
	data.Amount.Covered.Unlocked = r.received
	data.Complete = data.Amount.Covered.Unlocked >= data.Amount.Expected
	data.Description = r.description
	data.CreatedAt = r.createdAt
	data.Transaction = model.TransactionData{
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
	return sendCallbackRequest(data, r.callbackUrl, r.address)
}

// findMinCreationHeight returns the minimum creation height from the receiver map.
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

// updateReceivers updates receivers in the database that have been modified.
func (d *Daemon) updateReceivers(ctx context.Context, rs map[uint64]*recv) {
	for _, r := range rs {
		if !r.updated {
			continue
		}
		if _, err := d.db.ExecContext(ctx,
			"UPDATE receivers SET received_amount=$1 WHERE subaddress_index=$2",
			r.received, r.index); err != nil {
			log.Error().Err(err).Uint64("address_index", r.index).
				Msg("Failed to update payment request")
		}
	}
}

// makeRecvMap creates a map of receivers from database rows.
func makeRecvMap(rows *sql.Rows) map[uint64]*recv {
	rs := make(map[uint64]*recv)
	for rows.Next() {
		var t recv
		if err := rows.Scan(&t.index, &t.expected, &t.received, &t.description, &t.callbackUrl,
			&t.createdAt, &t.creationHeight, &t.address); err != nil {
			log.Error().Err(err).Msg("Failed to get payment requests from database")
		}
		rs[t.index] = &t
	}
	return rs
}

// Base query for selecting receivers joined with their subaddresses.
const receiversQuery = "SELECT r.subaddress_index,r.expected_amount,r.received_amount,r.description," +
	"r.callback_url,r.created_at,r.creation_height,s.address FROM receivers r JOIN subaddresses s ON r.subaddress_index=s.address_index"

// queryReceivers fetches receivers from the database.
// If addressIndices is nil, all receivers are returned.
// If addressIndices is provided, only matching receivers are returned (PostgreSQL uses ANY, SQLite does full scan).
func (d *Daemon) queryReceivers(ctx context.Context, addressIndices []uint64) (*sql.Rows, error) {
	if addressIndices == nil {
		return d.db.QueryContext(ctx, receiversQuery)
	}
	if d.config.SQLiteCS != "" {
		// SQLite doesn't support ANY, do full scan and filter in makeRecvMap caller
		return d.db.QueryContext(ctx, receiversQuery)
	}
	return d.db.QueryContext(ctx, receiversQuery+" WHERE r.subaddress_index = ANY($1)", addressIndices)
}

// checkTransfers checks for new confirmed transfers and sends callbacks.
func (d *Daemon) checkTransfers() {
	ctx := context.Background()
	rows, err := d.queryReceivers(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get payment requests from database")
		return
	}
	defer rows.Close()

	rs := makeRecvMap(rows)
	if len(rs) == 0 {
		return
	}

	resp, err := d.GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In:             true,
		FilterByHeight: true,
		// If there are very old rows and they aren't removed, there can be
		// performance issues
		MinHeight: findMinCreationHeight(rs),
	})
	if err != nil {
		checkFatalWalletError(err)
		log.Error().Err(err).Msg("Failed to get incoming transfers")
		return
	}
	if resp.In == nil {
		return
	}

	maxHeight := d.lastCallbackHeight
	for _, t := range resp.In {
		locked, eventHeight := getTransferLockStatus(t)
		if eventHeight <= d.lastCallbackHeight {
			continue
		}
		if r, ok := rs[t.SubaddrIndex.Minor]; ok {
			if !locked {
				r.received += t.Amount
				r.updated = true
			}
			if r.callbackUrl != "" {
				if err = d.sendPaymentCallback(ctx, r, &t, locked); err != nil {
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

	if maxHeight == d.lastCallbackHeight {
		return
	}
	d.lastCallbackHeight = maxHeight
	if err := d.saveLastCallbackHeight(ctx); err != nil {
		log.Error().Err(err).Uint64("height", d.lastCallbackHeight).Msg("Failed to save last callback height")
	} else {
		log.Info().Uint64("height", d.lastCallbackHeight).Msg("Saved last callback height")
	}
	d.updateReceivers(ctx, rs)
}

// checkMempool checks for new mempool (unconfirmed) transfers and sends 0-conf callbacks.
func (d *Daemon) checkMempool() {
	ctx := context.Background()
	resp, err := d.GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		Pool: true,
	})
	if err != nil {
		checkFatalWalletError(err)
		log.Error().Err(err).Msg("Failed to get mempool transfers")
		return
	}

	rows, err := d.db.QueryContext(ctx, "SELECT txid FROM mempool_seen")
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
		if _, err := d.db.ExecContext(ctx, "DELETE FROM mempool_seen"); err != nil {
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

	rows, err = d.queryReceivers(ctx, addressIndices)
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
				if err = d.sendPaymentCallback(ctx, r, &t, true); err != nil {
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
		if _, err := d.db.ExecContext(ctx, "INSERT INTO mempool_seen (txid) VALUES ($1)", t.Txid); err != nil {
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

	if d.config.SQLiteCS != "" {
		for _, txid := range toDelete {
			if _, err := d.db.ExecContext(ctx, "DELETE FROM mempool_seen WHERE txid=$1", txid); err != nil {
				log.Err(err).Str("txid", txid).Msg("Failed to delete old mempool_seen cache entry")
			}
		}
	} else {
		if _, err := d.db.ExecContext(ctx, "DELETE FROM mempool_seen WHERE txid = ANY($1)", toDelete); err != nil {
			log.Err(err).Msg("Failed to delete old mempool_seen cache entries")
		}
	}
}

// callbackRunner is the main loop that polls for transfers and sends callbacks.
func (d *Daemon) callbackRunner(ctx context.Context) {
	ticker := time.NewTicker(d.config.PollFreq)
	defer ticker.Stop()

	for {
		if d.config.ZeroConf {
			d.checkMempool()
		}

		heightResp, err := d.getHeight(ctx)
		if err != nil {
			checkFatalWalletError(err)
			log.Err(err).Msg("Failed to get height from wallet-rpc")
		} else {
			// If there was a new block, see if there is anything to callback
			if heightResp.Height > d.lastSeenHeight {
				d.checkTransfers()
				d.lastSeenHeight = heightResp.Height
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
