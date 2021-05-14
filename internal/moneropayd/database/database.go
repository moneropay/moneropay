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
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var DB *pgxpool.Pool

func Connect(host string, port uint, user, pass, dbname string) {
	u := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, pass, host, port, dbname)
	var err error
	DB, err = pgxpool.Connect(context.Background(), u)
	if err != nil {
		log.Fatalln("Unable to connect to database:", err)
	}
}

func Close() {
	DB.Close()
}

func Migrate() {
	b := &pgx.Batch{}
	b.Queue(`
	    CREATE TABLE IF NOT EXISTS subaddresses (
	        index	bigint PRIMARY KEY,
	        address	character(95) UNIQUE NOT NULL CHECK (LENGTH (address) = 95)
	    )`)
	b.Queue(`
	    CREATE TABLE IF NOT EXISTS receivers (
	        subaddress_index	bigint PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
	        expected_amount		bigint NOT NULL CHECK (expected_amount >= 0),
	        description		character varying(1024),
	        callback_url		character varying(1024),
		callback_failed		integer,
	        created_at		timestamp without time zone NOT NULL
	    )`)

	br := DB.SendBatch(context.Background(), b)
	for i := b.Len(); i > 0; i-- {
		if _, err := br.Exec(); err != nil {
			log.Fatal(err.Error())
		}
	}
}
