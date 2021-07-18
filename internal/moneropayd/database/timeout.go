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

package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
)

func QueryRowWithTimeout(ctx context.Context, t time.Duration, query string,
    args ...interface{}) (pgx.Row, error) {
	ctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()
	c := make(chan pgx.Row, 1)
	go func() { c <- DB.QueryRow(ctx, query, args...) }()
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

func QueryWithTimeout(ctx context.Context, t time.Duration, query string,
    args ...interface{}) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()
	c := make(chan queryRet, 1)
	go func() {
		rows, err := DB.Query(ctx, query, args...)
		c <- queryRet{rows, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case ret := <-c:
		return ret.rows, ret.err
	}
}

func ExecWithTimeout(ctx context.Context, t time.Duration, query string,
    args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()
	c := make(chan error, 1)
	go func() {
		_, err := DB.Exec(ctx, query, args...)
		c <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
}
