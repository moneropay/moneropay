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
	_, err := DB.Exec(context.Background(), `
	    CREATE TABLE IF NOT EXISTS metadata (
	        id			integer UNIQUE,
		version			text,
	        last_polled_block	bigint NOT NULL
	    )`)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = DB.Exec(context.Background(),`
	    INSERT INTO metadata (id, version, last_polled_block) VALUES (1, '0.1.0', 0) ON CONFLICT DO NOTHING`)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = DB.Exec(context.Background(), `
	    CREATE TABLE IF NOT EXISTS subaddresses (
	        index	bigint PRIMARY KEY,
	        address	character(95) UNIQUE NOT NULL CHECK (LENGTH (address) = 95)
	    )`)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = DB.Exec(context.Background(), `
	    CREATE TABLE IF NOT EXISTS receivers (
	        subaddress_index	bigint PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
	        expected_amount		bigint NOT NULL CHECK (expected_amount >= 0),
	        description		character varying(1024),
		callback_url		character varying(2048) NOT NULL,
	        created_at		timestamp with time zone DEFAULT CURRENT_TIMESTAMP
	    )`)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = DB.Exec(context.Background(), `
	    CREATE TABLE IF NOT EXISTS failed_callbacks (
	        id			serial PRIMARY KEY,
		subaddress_index	bigint REFERENCES subaddresses ON DELETE CASCADE,
		request_body		text NOT NULL,
		attempts		smallint DEFAULT 1,
		next_retry		timestamp with time zone NOT NULL
	    )`)
	if err != nil {
		log.Fatal(err.Error())
	}
}
