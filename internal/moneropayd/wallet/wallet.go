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

package wallet

import (
	"sync"

	"github.com/gabstv/httpdigest"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

var (
	Wallet *walletrpc.Client
	mutex sync.Mutex
)

func Lock() {
	mutex.Lock()
}

func Unlock() {
	mutex.Unlock()
}

// Initialize the Monero wallet RPC client.
func Init(RpcAddr string, RpcUser string, RpcPass string) {
        t := httpdigest.New(RpcUser, RpcPass)
        Wallet = walletrpc.New(walletrpc.Config{
                Address: RpcAddr,
                Transport: t,
        })
}
