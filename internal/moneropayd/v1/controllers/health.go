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

	"gitlab.com/moneropay/moneropay/internal/moneropayd/wallet"
	"gitlab.com/moneropay/moneropay/pkg/v1/models"
        "gitlab.com/moneropay/moneropay/internal/moneropayd/database"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	d := models.HealthGetResponse{
		Status: http.StatusOK,
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2 * time.Second)
	go func() {
		if err := database.DB.Ping(ctx); err == nil {
			d.Services.PostgreSQL = true
		}
		cancel()
	}()
	<-ctx.Done()
	rpc := wallet.Wallet
	wallet.Lock()
	if _, err := rpc.GetHeight(); err == nil {
		d.Services.WalletRPC = true
	}
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
