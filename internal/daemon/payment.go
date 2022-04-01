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
	"gitlab.com/moneropay/go-monero/walletrpc"
)

func Receive(xmr uint64, desc, callbackUrl string) (string, time.Time, error) {
	resp, err := createAddress(&walletrpc.CreateAddressRequest{})
	if err != nil {
		return "", time.Time{}, err
	}
	t := time.Now()
	var tx pgx.Tx
	ctx, cancel := context.WithTimeout(context.Background(), 4 * time.Second)
	go func() {
		defer cancel()
		tx, err = pdb.Begin(ctx)
		if err != nil {
			return
		}
		if _, err = tx.Exec(ctx, "INSERT INTO subaddresses (index, address) VALUES ($1, $2)",
		    resp.AddressIndex, resp.Address); err != nil {
			tx.Rollback(ctx)
			return
		}
		if _, err = tx.Exec(ctx, "INSERT INTO receivers (subaddress_index, expected_amount, " +
		    "description, callback_url, created_at) VALUES ($1, $2, $3, $4, $5)", resp.AddressIndex, xmr,
		    desc, callbackUrl, t); err != nil {
			tx.Rollback(ctx)
			return
		}
		if err = tx.Commit(ctx); err != nil {
			return
		}
	}()
	<-ctx.Done()
	if err != nil {
		return "", time.Time{}, err
	}
	return resp.Address, t, nil
}

func GetReceiver(address string) (pgx.Row, error) {
	row, err := pdbQueryRow(context.Background(), 3 * time.Second,
	    "SELECT index, expected_amount, description, created_at " +
	    "FROM subaddresses, receivers WHERE index=subaddress_index " +
	    "AND address=$1", address)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func GetReceived(index, min, max uint64) ([]walletrpc.Transfer, error) {
	resp, err := GetTransfers(&walletrpc.GetTransfersRequest{
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
