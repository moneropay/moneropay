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
	"net/http"
	"time"

	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

func Health(ctx context.Context) (model.HealthResponse) {
	d := model.HealthResponse{Status: http.StatusOK}
	ctx, c1 := context.WithTimeout(context.Background(), 10 * time.Second)
	defer c1()
	if err := db.PingContext(ctx); err == nil {
		if Config.sqliteCS != "" {
			d.Services.SQLite = true
		} else {
			d.Services.PostgreSQL = true
		}
	}
	ctx, c2 := context.WithTimeout(context.Background(), 10 * time.Second)
	defer c2()
	wMutex.Lock()
	if _, err := wallet.GetHeight(ctx); err == nil {
		d.Services.WalletRPC = true
	}
	wMutex.Unlock()
	if !(d.Services.PostgreSQL || d.Services.SQLite) || !d.Services.WalletRPC {
		d.Status = http.StatusServiceUnavailable
	}
	return d
}
