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
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

func Receive(ctx context.Context, xmr uint64, desc, callbackUrl string) (string, time.Time, error) {
	resp, err := createAddress(ctx, &walletrpc.CreateAddressRequest{})
	if err != nil {
		return "", time.Time{}, err
	}
	t := time.Now()
	var tx pgx.Tx
	tx, err = pdb.Begin(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO subaddresses(address_index,address)VALUES($1,$2)",
	    resp.AddressIndex, resp.Address); err != nil {
		tx.Rollback(ctx)
		return "", time.Time{}, err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO receivers(subaddress_index,expected_amount,description," +
	    "callback_url,created_at,received_amount,last_height)VALUES($1,$2,$3,$4,$5,0,$6)",
	    resp.AddressIndex, xmr, desc, callbackUrl, t, callbackLastHeight); err != nil {
		tx.Rollback(ctx)
		return "", time.Time{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", time.Time{}, err
	}
	log.Info().Uint64("amount", xmr).Str("description", desc).Str("callback_url",callbackUrl).
	    Msg("Created new payment request")
	return resp.Address, t, err
}

type Receiver struct {
	Index, Expected, Received uint64
	Description string
	CreatedAt time.Time
}

func GetReceiver(ctx context.Context, address string) (Receiver, error) {
	type ret struct {resp Receiver; err error}
	c := make(chan ret)
	go func() {
		var r ret
		row := pdb.QueryRow(ctx,
		    "SELECT address_index, expected_amount, received_amount, description, created_at " +
		    "FROM subaddresses, receivers WHERE address_index=subaddress_index AND address=$1",
		    address)
		r.err = row.Scan(&r.resp.Index, &r.resp.Expected, &r.resp.Received, &r.resp.Description,
		    &r.resp.CreatedAt)
		c <- r
	}()
	select {
		case <-ctx.Done(): return Receiver{}, ctx.Err()
		case r := <-c: return r.resp, r.err
	}
}

func GetReceivedTransfers(ctx context.Context, index, min, max uint64) ([]walletrpc.Transfer, error) {
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		SubaddrIndices: []uint64{index},
		In: true,
		FilterByHeight: (min > 0 || max > 0),
		MinHeight: min,
		MaxHeight: max,
	})
	if err != nil {
		return nil, err
	}
	return resp.In, nil
}
