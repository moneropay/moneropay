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
	"context"
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"

	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

// Receive creates a new payment request with a unique subaddress.
func (d *Daemon) Receive(ctx context.Context, xmr uint64, desc, callbackUrl string) (string, time.Time, error) {
	resp, err := d.createAddress(ctx, &walletrpc.CreateAddressRequest{})
	if err != nil {
		return "", time.Time{}, err
	}

	t := time.Now()
	var tx *sql.Tx
	tx, err = d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", time.Time{}, err
	}

	if _, err = tx.ExecContext(ctx, "INSERT INTO subaddresses(address_index,address)VALUES($1,$2)",
		resp.AddressIndex, resp.Address); err != nil {
		tx.Rollback()
		return "", time.Time{}, err
	}

	h, err := d.getHeight(ctx)
	if err != nil {
		return "", time.Time{}, err
	}

	if _, err = tx.ExecContext(ctx, "INSERT INTO receivers(subaddress_index,expected_amount,description,"+
		"callback_url,created_at,received_amount,creation_height)VALUES($1,$2,$3,$4,$5,0,$6)",
		resp.AddressIndex, xmr, desc, callbackUrl, t, h.Height); err != nil {
		tx.Rollback()
		return "", time.Time{}, err
	}

	if err = tx.Commit(); err != nil {
		return "", time.Time{}, err
	}

	log.Info().Uint64("amount", xmr).Str("description", desc).Str("callback_url", callbackUrl).
		Msg("Created new payment request")
	return resp.Address, t, err
}

// Receiver holds payment request data from the database.
type Receiver struct {
	Index, Expected uint64
	Description     string
	CreatedAt       time.Time
}

// getReceiver retrieves receiver data for a given address.
func (d *Daemon) getReceiver(ctx context.Context, address string) (Receiver, error) {
	var r Receiver
	row := d.db.QueryRowContext(ctx,
		"SELECT address_index,expected_amount,description,created_at "+
			"FROM subaddresses,receivers WHERE address_index=subaddress_index AND address=$1",
		address)
	err := row.Scan(&r.Index, &r.Expected, &r.Description, &r.CreatedAt)
	return r, err
}

// getReceivedTransfers retrieves transfers for a specific subaddress.
func (d *Daemon) getReceivedTransfers(ctx context.Context, index, min, max uint64) ([]walletrpc.Transfer, error) {
	resp, err := d.GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		SubaddrIndices: []uint64{index},
		In:             true,
		Pool:           d.config.ZeroConf,
		FilterByHeight: (min > 0 || max > 0),
		MinHeight:      min,
		MaxHeight:      max,
	})
	if err != nil {
		return nil, err
	}

	comb := resp.In
	if d.config.ZeroConf {
		comb = append(resp.In, resp.Pool...)
	}
	return comb, nil
}

// GetPaymentRequest retrieves the status of a payment request.
func (d *Daemon) GetPaymentRequest(ctx context.Context, address string, min, max uint64) (model.ReceiveGetResponse, error) {
	var data model.ReceiveGetResponse

	// Get data for address from DB.
	recv, err := d.getReceiver(ctx, address)
	if err != nil {
		return data, err
	}

	// TODO: This call to wallet RPC can be avoided by caching the
	// get_transfers response in the callback runner
	tx, err := d.getReceivedTransfers(ctx, recv.Index, min, max)
	if err != nil {
		return data, err
	}

	var total, unlocked uint64
	for _, r1 := range tx {
		isLocked, _ := getTransferLockStatus(r1)
		if !isLocked {
			unlocked += r1.Amount
		}
		total += r1.Amount
		r2 := model.TransactionData{
			Amount:          r1.Amount,
			Confirmations:   r1.Confirmations,
			DoubleSpendSeen: r1.DoubleSpendSeen,
			Fee:             r1.Fee,
			Height:          r1.Height,
			Timestamp:       time.Unix(int64(r1.Timestamp), 0),
			TxHash:          r1.Txid,
			UnlockTime:      r1.UnlockTime,
			Locked:          isLocked,
		}
		data.Transactions = append(data.Transactions, r2)
	}

	data.Amount.Expected = recv.Expected
	data.Description = recv.Description
	data.CreatedAt = recv.CreatedAt
	data.Amount.Covered.Total = total
	data.Amount.Covered.Unlocked = unlocked
	data.Complete = data.Amount.Covered.Unlocked >= data.Amount.Expected
	return data, nil
}

// DeletePaymentRequest removes a payment request by address.
func (d *Daemon) DeletePaymentRequest(ctx context.Context, address string) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM subaddresses WHERE address=$1", address)
	return err
}
