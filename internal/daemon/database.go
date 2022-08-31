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
	"github.com/rs/zerolog/log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var pdb *pgxpool.Pool

func pdbConnect() {
	var err error
	if pdb, err = pgxpool.Connect(context.Background(), Config.postgresCS); err != nil {
		log.Fatal().Err(err).Msg("Startup failure")
	}
}

func pdbMigrate() {
	m, err := migrate.New("file://db/postgres", Config.postgresCS)
	if err != nil {
		log.Fatal().Err(err).Msg("Startup failure")
	}
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return
		}
		log.Fatal().Err(err).Msg("Startup failure")
	}
}

func pdbQueryRow(ctx context.Context, query string, args ...interface{}) (pgx.Row, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
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

func pdbQuery(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	type queryRet struct { rows pgx.Rows; err error }
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

func pdbExec(ctx context.Context, query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
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
