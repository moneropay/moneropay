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
	"github.com/jackc/pgx/v4/pgxpool"
)

var pdb *pgxpool.Pool

func pdbConnect() error {
	var err error
	pdb, err = pgxpool.Connect(context.Background(), Config.postgresCS);
	if err != nil {
		return err
	}
	return nil
}

func pdbQueryRow(ctx context.Context, t time.Duration, query string,
    args ...interface{}) (pgx.Row, error) {
	ctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()
	c := make(chan pgx.Row, 1)
	go func() { c <- pdb.QueryRow(ctx, query, args...) }()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case row := <-c:
		return row, nil
	}
}

type queryRet struct {
	rows pgx.Rows
	err error
}

func pdbQuery(ctx context.Context, t time.Duration, query string,
    args ...interface{}) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()
	c := make(chan queryRet, 1)
	go func() {
		rows, err := pdb.Query(ctx, query, args...)
		c <- queryRet{rows, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case ret := <-c:
		return ret.rows, ret.err
	}
}

func pdbExec(ctx context.Context, t time.Duration, query string,
    args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()
	c := make(chan error, 1)
	go func() {
		_, err := pdb.Exec(ctx, query, args...)
		c <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func pdbMigrate() error {
	if err := pdbExec(context.Background(), 3 * time.Second, `
	    CREATE TABLE IF NOT EXISTS metadata (
	        key	text UNIQUE NOT NULL,
	        value	bigint NOT NULL
	    );
	    INSERT INTO metadata (key, value) VALUES ('last_height', 0) ON CONFLICT DO NOTHING;
	    CREATE TABLE IF NOT EXISTS subaddresses (
	        index		bigint PRIMARY KEY,
	        address		character(95) UNIQUE NOT NULL,
	        used_until	bigint
	    );
	    CREATE TABLE IF NOT EXISTS receivers (
	        subaddress_index	bigint PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
	        expected_amount		bigint NOT NULL CHECK (expected_amount >= 0),
	        description		character varying(1024),
	        callback_url		character varying(2048) NOT NULL,
	        created_at		timestamp with time zone
	    );`); err != nil {
		return err
	}
	return nil
}
