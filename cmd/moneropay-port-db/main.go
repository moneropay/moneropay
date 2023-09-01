/*
 * moneropay-db-port is a helper program to switch between PostgreSQL
 * and SQLite3.
 * Copyright (C) 2023 İrem Kuyucu <siren@kernal.eu>
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

package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/moneropay/moneropay/v2/internal/daemon"
)

type dbPorter struct {
	source, target *sql.DB
	tx             *sql.Tx
	timeout        time.Duration
}

// This program should not be run while MoneroPay is running. The source
// database's scheme must be up-to-date with the version tag this tool is from.
func main() {
	sourceCs, targetCs, timeout := parseOptions()
	porter := newPorter(sourceCs, targetCs, timeout)

	porter.migrateLastBlockHeightRow()
	porter.migrateSubaddressesRows()
	porter.migrateReceiversRows()

	if err := porter.tx.Commit(); err != nil {
		log.Fatal("Failed to commit changes")
	}
}

func newPorter(sourceCs, targetCs string, timeout time.Duration) *dbPorter {
	var (
		d   dbPorter
		err error
	)

	// Connect and run database migration files
	d.source = connect(sourceCs)

	d.target = connect(targetCs)

	// Start a transaction. This will revert all changes if something fails.
	d.tx, err = d.target.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Fatal(err)
	}

	d.timeout = timeout
	return &d
}

func parseOptions() (sourceCs, targetCs string, timeout time.Duration) {
	flag.StringVar(&sourceCs, "from", "", "URI to existing database to migrate from")
	flag.StringVar(&targetCs, "to", "", "URI to new database to migrate to")
	flag.DurationVar(&timeout, "timeout", 5*time.Minute, "Timeout for database queries")
	flag.Parse()

	if sourceCs == "" || targetCs == "" {
		log.Println("-from and -to cannot be empty.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	return
}

func connect(conn string) (db *sql.DB) {
	u, err := url.Parse(conn)
	if err != nil {
		log.Fatal("Failed to parse database connection string: ", err)
	}
	log.Println("scheme: ", u.Scheme)
	if u.Scheme == "postgres" || u.Scheme == "postgresql" {
		daemon.DbMigrate("file://db/postgres", conn)
		if db, err = sql.Open("pgx", conn); err != nil {
			log.Fatal("Failed to open PostgreSQL database: ", err)
		}
		log.Println("Connected to postgresql")
		return
	}

	// go-migrate's sqlite3 library doesn't use standard DSN connection strings
	daemon.DbMigrate("file://db/sqlite3", daemon.SqliteMigrateParseDSN(conn))

	if db, err = sql.Open("sqlite3", conn); err != nil {
		log.Fatal("Failed to open SQLite3 database: ", err)
	}
	log.Println("Connected to sqlite3")
	return
}

func (d *dbPorter) migrateLastBlockHeightRow() {
	ctx1, c1 := context.WithTimeout(context.Background(), d.timeout)
	defer c1()
	row := d.source.QueryRowContext(ctx1, "SELECT height FROM last_block_height")
	var height uint64
	if err := row.Scan(&height); err != nil {
		if rollbackErr := d.tx.Rollback(); rollbackErr != nil {
			log.Fatalf("last_block_height: select failed: %v, unable to back: %v",
				err, rollbackErr)
		}
		log.Fatal("last_block_height select: ", err)
	}

	ctx2, c2 := context.WithTimeout(context.Background(), d.timeout)
	defer c2()
	_, err := d.tx.ExecContext(ctx2, "UPDATE last_block_height SET height=$1", height)
	if err != nil {
		if rollbackErr := d.tx.Rollback(); rollbackErr != nil {
			log.Fatalf("last_block_height: update failed: %v, unable to back: %v",
				err, rollbackErr)
		}
		log.Fatal("last_block_height update: ", err)

	}
}

func (d *dbPorter) migrateSubaddressesRows() {
	ctx1, c1 := context.WithTimeout(context.Background(), d.timeout)
	defer c1()
	rows, err := d.source.QueryContext(ctx1,
		"SELECT address_index, address FROM subaddresses")
	if err != nil {
		log.Fatal("subaddresses select: ", err)
	}
	defer rows.Close()

	type subaddress struct {
		addressIndex uint64
		address      string
	}
	for rows.Next() {
		var s subaddress
		if err := rows.Scan(&s.addressIndex, &s.address); err != nil {
			if rollbackErr := d.tx.Rollback(); rollbackErr != nil {
				log.Fatalf("subaddresses: scan failed: %v, unable to back: %v",
					err, rollbackErr)
			}
			log.Fatal("subaddresses scan: ", err)
		}
		ctx2, c2 := context.WithTimeout(context.Background(), d.timeout)
		defer c2()
		_, err := d.tx.ExecContext(ctx2,
			"INSERT INTO subaddresses(address_index,address)VALUES($1,$2)",
			s.addressIndex, s.address)
		if err != nil {
			if rollbackErr := d.tx.Rollback(); rollbackErr != nil {
				log.Fatalf("subaddresses: insert failed: %v, unable to back: %v",
					err, rollbackErr)
			}
			log.Fatal("subaddresses insert: ", err)
		}
	}
}

func (d *dbPorter) migrateReceiversRows() {
	ctx1, c1 := context.WithTimeout(context.Background(), d.timeout)
	defer c1()
	rows, err := d.source.QueryContext(ctx1,
		"SELECT subaddress_index,expected_amount,description,"+
			"callback_url,created_at,received_amount,creation_height"+
			" FROM receivers")
	if err != nil {
		log.Fatal("receivers select: ", err)
	}

	defer rows.Close()

	type receiver struct {
		index, expected, received, creationHeight uint64
		description, callbackUrl                  string
		createdAt                                 time.Time
	}

	for rows.Next() {
		var r receiver
		if err := rows.Scan(&r.index, &r.expected, &r.description,
			&r.callbackUrl, &r.createdAt, &r.received,
			&r.creationHeight); err != nil {
			if rollbackErr := d.tx.Rollback(); rollbackErr != nil {
				log.Fatalf("receivers: scan failed: %v, unable to back: %v",
					err, rollbackErr)
			}
			log.Fatal("receivers scan: ", err)
		}

		ctx2, c2 := context.WithTimeout(context.Background(), d.timeout)
		defer c2()
		_, err := d.tx.ExecContext(ctx2,
			"INSERT INTO receivers(subaddress_index,expected_amount,"+
				"description,callback_url,created_at,received_amount,"+
				"creation_height)VALUES($1,$2,$3,$4,$5,$6,$7)",
			r.index, r.expected, r.description, r.callbackUrl,
			r.createdAt, r.received, r.creationHeight)
		if err != nil {
			if rollbackErr := d.tx.Rollback(); rollbackErr != nil {
				log.Fatalf("receivers: insert failed: %v, unable to back: %v",
					err, rollbackErr)
			}
			log.Fatal("receivers insert: ", err)
		}
	}
}
