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

package main

import (
	"gitlab.com/moneropay/moneropay/internal/moneropayd/callback"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/config"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/database"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/router"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/wallet"
)

func main() {
	// Parse command-line arguments and fill the Values struct.
	config.Init()

	// Initialize Monero wallet RPC.
	wallet.Init(config.Values.RpcAddr, config.Values.RpcUser, config.Values.RpcPass)

	// Initialize the database.
	database.Connect(
		config.Values.PostgresHost, config.Values.PostgresPort,
		config.Values.PostgresUser, config.Values.PostgresPass,
		config.Values.PostgresDBName,
	)
	defer database.Close()
	database.Migrate()

	// Poll and update the database
	go callback.Run()

	// Start the router.
	router.Run(config.Values.BindAddr)
}
