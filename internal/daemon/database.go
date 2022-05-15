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
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var pdb *pgxpool.Pool

func pdbConnect() {
	var err error
	if pdb, err = pgxpool.Connect(context.Background(), Config.postgresCS); err != nil {
		log.Fatal(err)
	}
}

func pdbMigrate() {
	m, err := migrate.New("file://db/postgres", Config.postgresCS)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return
		}
		log.Fatal(err)
	}
}
