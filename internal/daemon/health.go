/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2022 İrem Kuyucu <siren@kernal.eu>
 * Copyright (C) 2024 Laurynas Četyrkinas <gpg@gpg.li>
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
	"net/http"
	"sync"
	"time"

	"gitlab.com/moneropay/go-monero/walletrpc"
	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

func Health(ctx context.Context) model.HealthResponse {
	d := model.HealthResponse{Status: http.StatusOK}
	var wg sync.WaitGroup
	ctxt, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := db.PingContext(ctxt); err != nil {
			return
		}
		if Config.sqliteCS != "" {
			d.Services.SQLite = true
		} else {
			d.Services.PostgreSQL = true
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		wMutex.Lock()
		defer wMutex.Unlock()
		if _, err := wallet.Refresh(ctxt, &walletrpc.RefreshRequest{}); err != nil {
			return
		}
		d.Services.WalletRPC = true
	}()
	wg.Wait()
	if !(d.Services.PostgreSQL || d.Services.SQLite) || !d.Services.WalletRPC {
		d.Status = http.StatusServiceUnavailable
	}
	return d
}
