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

package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gitlab.com/moneropay/moneropay/internal/moneropayd/config"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/wallet"
	"gitlab.com/moneropay/moneropay/pkg/models"
        "gitlab.com/moneropay/moneropay/internal/moneropayd/database"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "MoneroPay/" + config.Version)

	d := models.HealthGetResponse{
		Status: http.StatusOK,
	}
	if err := database.ExecWithTimeout(r.Context(), 2 * time.Second,
	    "SELECT value FROM metadata WHERE key = 'last_height'"); err == nil {
		d.Services.PostgreSQL = true
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2 * time.Second)
	wallet.Lock()
	if _, err := wallet.Wallet.GetHeight(ctx); err == nil {
		d.Services.WalletRPC = true
	}
	defer cancel()
	wallet.Unlock()
	if !d.Services.WalletRPC || !d.Services.PostgreSQL {
		d.Status = http.StatusServiceUnavailable
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(d.Status)
		json.NewEncoder(w).Encode(d)
	default:
		w.WriteHeader(d.Status)
	}
}
