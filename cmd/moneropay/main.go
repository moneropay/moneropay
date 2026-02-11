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

package main

import (
	"gitlab.com/moneropay/moneropay/v2/internal/daemon"
	"gitlab.com/moneropay/moneropay/v2/internal/server"
)

func main() {
	// Load configuration from flags and environment
	cfg := daemon.LoadConfig()

	// Setup logger
	daemon.SetupLogger(cfg.LogFormat)

	// Create and start the daemon
	d := daemon.New(cfg)
	d.Run()

	// Start the HTTP server
	server.Run(d)
}
