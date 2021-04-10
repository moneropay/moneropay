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

import "flag"

type Data struct {
	BindAddr string
	RpcAddr string
	RpcUser string
	RpcPass string
	TransferMixin uint64
	TransferPriority uint
	TransferUnlockTime uint64
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
	flag.Parse()
}
