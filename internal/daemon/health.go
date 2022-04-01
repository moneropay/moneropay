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
)

type HealthStatus struct {
	Status int `json:"status"`
	Services struct {
		WalletRPC bool `json:"walletrpc"`
		PostgreSQL bool `json:"postgresql"`
	} `json:"services"`
}

func Health() (HealthStatus) {
	d := HealthStatus{Status: http.StatusOK}
	if err := pdbExec(context.Background(), 3 * time.Second,
	    "SELECT value FROM metadata WHERE key = 'last_height'"); err == nil {
		d.Services.PostgreSQL = true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	wMutex.Lock()
	if _, err := wallet.GetHeight(ctx); err == nil {
		d.Services.WalletRPC = true
	}
	defer cancel()
	wMutex.Unlock()
	if !d.Services.PostgreSQL || !d.Services.WalletRPC {
		d.Status = http.StatusServiceUnavailable
	}
	return d
}
