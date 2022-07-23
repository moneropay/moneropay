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
	"time"
)

const Version = "2.2.0"

func init() {
	loadConfig()
	logger()
	walletConnect()
	pdbMigrate()
	pdbConnect()
	daemonMigrate()
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	readCallbackLastHeight(ctx)
}

func Run() {
	go callbackRunner()
	go transferAccountingRunner()
}
