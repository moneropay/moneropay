/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
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

package server

import (
	"net/http"

	"gitlab.com/moneropay/moneropay/v2/internal/daemon"
)

// middlewareServerHeader adds the Server header to all responses.
func middlewareServerHeader(_ *daemon.Daemon) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Server", "MoneroPay/"+daemon.Version)
			next.ServeHTTP(w, r)
		})
	}
}

// middlewareXMoneroPayAddressHeader adds the wallet address header to all responses.
// The header is omitted if the address is not yet known (e.g., before keys are consumed in view-only init mode).
func middlewareXMoneroPayAddressHeader(d *daemon.Daemon) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if d.WalletPrimaryAddress != "" {
				w.Header().Set("X-MoneroPay-Address", d.WalletPrimaryAddress)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// middlewareContentType sets Content-Type to application/json for all responses.
func middlewareContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
