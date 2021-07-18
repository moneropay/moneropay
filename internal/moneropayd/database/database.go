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
	"time"

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
	err := ExecWithTimeout(context.Background(), 3 * time.Second, `
	    CREATE TABLE IF NOT EXISTS metadata (
	        key	text UNIQUE NOT NULL,
	        value	bigint NOT NULL
	    );
	    INSERT INTO metadata (key, value) VALUES ('last_height', 0) ON CONFLICT DO NOTHING;
	    CREATE TABLE IF NOT EXISTS subaddresses (
	        index	bigint PRIMARY KEY,
	        address	character(95) UNIQUE NOT NULL CHECK (LENGTH (address) = 95)
	    );
	    CREATE TABLE IF NOT EXISTS receivers (
	        subaddress_index	bigint PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
	        expected_amount		bigint NOT NULL CHECK (expected_amount >= 0),
	        description		character varying(1024),
	        callback_url		character varying(2048) NOT NULL,
	        created_at		timestamp with time zone
	    );
	    CREATE TABLE IF NOT EXISTS failed_callbacks (
	        uid			serial PRIMARY KEY,
	        subaddress_index	bigint REFERENCES subaddresses ON DELETE CASCADE,
	        request_body		text NOT NULL,
	        attempts		smallint DEFAULT 1,
	        next_retry		timestamp with time zone NOT NULL
	    )`)
	if err != nil {
		log.Fatal(err)
	}
}
