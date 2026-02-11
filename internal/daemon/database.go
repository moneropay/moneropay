/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
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
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// connectDB establishes connection to the database and runs migrations.
func (d *Daemon) connectDB() {
	var err error

	if d.config.SQLiteCS != "" {
		DbMigrate("file://db/sqlite3", SqliteMigrateParseDSN(d.config.SQLiteCS))
		if d.db, err = sql.Open("sqlite3", d.config.SQLiteCS); err != nil {
			log.Fatal().Err(err).Msg("Failed to open SQLite3 database")
		}
		return
	}

	DbMigrate("file://db/postgres", d.config.PostgresCS)
	if d.db, err = sql.Open("pgx", d.config.PostgresCS); err != nil {
		log.Fatal().Err(err).Msg("Failed to open PostgreSQL database")
	}
}

// DbMigrate runs database migrations from the given URL to the connection.
// This is exported for use by moneropay-port-db tool.
func DbMigrate(migrateURL, conn string) {
	m, err := migrate.New(migrateURL, conn)
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

// SqliteMigrateParseDSN converts a SQLite DSN to the format expected by golang-migrate.
// go-migrate's sqlite3 library doesn't use standard DSN connection strings.
// This is exported for use by moneropay-port-db tool.
func SqliteMigrateParseDSN(conn string) string {
	u, err := url.Parse(conn)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse sqlite3 connection string")
	}
	u.Scheme = ""
	return "sqlite3://" + u.String()
}
