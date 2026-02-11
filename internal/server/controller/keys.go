/*
 * MoneroPay is a Monero payment processor.
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

package controller

import (
	"encoding/json"
	"net/http"
)

// KeysGetHandler returns the one-time wallet keys (view-only init mode only).
func (c *Controller) KeysGetHandler(w http.ResponseWriter, r *http.Request) {
	if !c.Daemon.IsViewOnlyInit() {
		writeError(w, http.StatusNotFound, nil, "No keys available. Server was not started with --init-view-only flag")
		return
	}
	if c.Daemon.IsKeysConsumed() {
		writeError(w, http.StatusGone, nil, "Keys have already been retrieved and deleted")
		return
	}

	keys, err := c.Daemon.GetOneTimeKeys()
	if err != nil {
		writeError(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	json.NewEncoder(w).Encode(keys)
}
