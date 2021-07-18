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

package config

import (
	"strings"
	"time"
	"log"

	"github.com/namsral/flag"
)

type Data struct {
	BindAddr string
	RpcAddr string
	RpcUser string
	RpcPass string
	TransferMixin uint64
	TransferPriority uint
	TransferUnlockTime uint64
	PostgresHost string
	PostgresPort uint
	PostgresUser string
	PostgresPass string
	PostgresDBName string
	Intervals []time.Duration
}

var Values Data

func Init() {
	flag.StringVar(&Values.BindAddr, "bind", "localhost:5000", "Bind address:port for moneropayd")
	flag.StringVar(&Values.RpcAddr, "rpc-address", "http://localhost:18082/json_rpc", "Wallet RPC server address")
	flag.StringVar(&Values.RpcUser, "rpc-username", "", "Username for monero-wallet-rpc")
	flag.StringVar(&Values.RpcPass, "rpc-password", "", "Password for monero-wallet-rpc")
	flag.Uint64Var(&Values.TransferMixin, "transfer-mixin", 8, "Number of outputs from the blockchain to mix with (0 means no mixing)")
	flag.UintVar(&Values.TransferPriority, "transfer-priority", 0, "Set a priority for transactions")
	flag.Uint64Var(&Values.TransferUnlockTime, "transfer-unlock-time", 10, "Number of blocks before the monero can be spent (0 to not add a lock)")
	flag.StringVar(&Values.PostgresHost, "postgres-host", "localhost", "PostgreSQL database address")
	flag.UintVar(&Values.PostgresPort, "postgres-port", 5432, "PostgreSQL database port")
	flag.StringVar(&Values.PostgresUser, "postgres-username", "moneropay", "Username for PostgreSQL database")
	flag.StringVar(&Values.PostgresPass, "postgres-password", "", "Password for PostgreSQL database")
	flag.StringVar(&Values.PostgresDBName, "postgres-database", "moneropay", "Name for PostgreSQL database")
	var i string
	flag.StringVar(&i, "intervals", "1m,5m,15m,30m,1h", "Comma seperated list of callback intervals")
	flag.Parse()
	for _, m := range strings.Split(i, ",") {
		v, err := time.ParseDuration(m)
		if err != nil {
			log.Fatal(err)
		}
		// TODO: Make sure no values less than 30sec or however long is the thread interval time.
		Values.Intervals = append(Values.Intervals, v)
	}
}
