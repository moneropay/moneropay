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

	"github.com/jackc/pgx/v4/pgxpool"
)

var pdb *pgxpool.Pool

func pdbConnect() error {
	var err error
	pdb, err = pgxpool.Connect(context.Background(), Config.postgresCS);
	return err
}

func pdbMigrate(ctx context.Context) error {
	c := make(chan error)
	go func() {
		_, err := pdb.Exec(ctx, `
		    CREATE TABLE IF NOT EXISTS last_block_height (
		        height	bigint NOT NULL DEFAULT 0
		    );
		    INSERT INTO last_block_height (height) VALUES (0) ON CONFLICT DO NOTHING;
		    CREATE TABLE IF NOT EXISTS subaddresses (
		        address_index	bigint PRIMARY KEY,
		        address		character(95) UNIQUE NOT NULL,
		        used_until	bigint
		    );
		    CREATE TABLE IF NOT EXISTS receivers (
		        subaddress_index	bigint PRIMARY KEY REFERENCES subaddresses(address_index) ON DELETE CASCADE,
		        expected_amount		bigint NOT NULL CHECK (expected_amount >= 0),
		        description		character varying(1024),
		        callback_url		character varying(2048) NOT NULL,
		        created_at		timestamp
		    );`)
		c <- err
	}()
	select {
		case <-ctx.Done(): return ctx.Err()
		case err := <-c: return err
	}
}
