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
	"database/sql"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func dbConnect() {
	var err error
	if Config.postgresCS != "" {
		dbMigrate("file://db/postgres", Config.postgresCS)
		if db, err = sql.Open("pgx", Config.postgresCS); err != nil {
			log.Fatal().Err(err).Msg("Failed to open PostgreSQL database")
		}
	} else {
		dbMigrate("file://db/sqlite3", sqliteMigrateParseDSN(Config.sqliteCS))
		if db, err = sql.Open("sqlite3", Config.sqliteCS); err != nil {
			log.Fatal().Err(err).Msg("Failed to open SQLite3 database")
		}
	}
}

func dbMigrate(url, conn string) {
	m, err := migrate.New(url, conn)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize new migrate instance")
	}
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return
		}
		log.Fatal().Err(err).Msg("Failed to apply migrations")
	}
}

// go-migrate's sqlite3 library doesn't use standard DSN connection strings
func sqliteMigrateParseDSN(conn string) string {
	u, err := url.Parse(conn)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse sqlite3 connection string")
	}
	u.Scheme = ""
	return "sqlite3://" + u.String()
}
