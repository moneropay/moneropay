/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2022 İrem Kuyucu <siren@kernal.eu>
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
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

	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

// Health checks the health of all daemon services and returns their status.
func (d *Daemon) Health(ctx context.Context) model.HealthResponse {
	resp := model.HealthResponse{Status: http.StatusOK}

	// In view-only init mode with keys pending, return OK with partial status
	// The server is running and ready to serve /keys
	if d.IsKeysPending() {
		resp.Services.KeysPending = true
		return resp
	}

	done := make(chan struct{})
	ctxt, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	go func() {
		var wg sync.WaitGroup

		wg.Go(func() {
			if d.db == nil {
				return
			}
			if err := d.db.PingContext(ctxt); err != nil {
				return
			}
			if d.config.SQLiteCS != "" {
				resp.Services.SQLite = true
			} else {
				resp.Services.PostgreSQL = true
			}
		})

		wg.Go(func() {
			if _, err := d.refresh(ctxt); err != nil {
				return
			}
			resp.Services.WalletRPC = true
		})

		wg.Wait()
		close(done)
	}()

	select {
	case <-ctxt.Done():
		resp.Status = http.StatusServiceUnavailable
	case <-done:
		if !(resp.Services.PostgreSQL || resp.Services.SQLite) || !resp.Services.WalletRPC {
			resp.Status = http.StatusServiceUnavailable
		}
	}
	return resp
}
