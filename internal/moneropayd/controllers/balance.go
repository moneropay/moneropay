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
	"encoding/json"
	"net/http"

	"gitlab.com/moneropay/go-monero/walletrpc"

	"gitlab.com/moneropay/moneropay/internal/moneropayd/helpers"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/wallet"
	"gitlab.com/moneropay/moneropay/pkg/models"
)

func BalanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	wallet.Lock()
	resp, err := wallet.Wallet.GetBalance(&walletrpc.GetBalanceRequest{})
	wallet.Unlock()
	if err != nil {
		_, werr := walletrpc.GetWalletError(err)
		helpers.WriteError(w, http.StatusInternalServerError, (*int)(&werr.Code), werr.Message)
		return
	}
	d := models.BalanceGetResponse{
		Total: resp.Balance,
		Unlocked: resp.UnlockedBalance,
		Locked: resp.Balance - resp.UnlockedBalance,
	}
	json.NewEncoder(w).Encode(d)
}
