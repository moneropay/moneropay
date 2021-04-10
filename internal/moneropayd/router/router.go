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

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"

	"gitlab.com/kernal/moneropay/internal/moneropayd/walletrpc"
	"gitlab.com/kernal/moneropay/internal/moneropayd/config"
)

var (
	r = gin.Default()
	w wallet.Client
)

func getRoutes() {
	v1 := r.Group("/v1")
	addPingRoutes(v1)
	addBalanceRoutes(v1)
	addAddressRoutes(v1)
	addTransferRoutes(v1)
}

// Start the API server.
func Run() {
	// Initialize Monero wallet RPC client
	w = walletrpc.Init(config.Values.RpcAddr, config.Values.RpcUser, config.Values.RpcPass)

	getRoutes()
	r.Run(config.Values.BindAddr)
}
