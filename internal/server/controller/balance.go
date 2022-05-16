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

package controller

import (
	"encoding/json"
	"net/http"

	"gitlab.com/moneropay/moneropay/v2/internal/daemon"
)

type balanceResponse struct {
	Total uint64 `json:"total"`
	Unlocked uint64 `json:"unlocked"`
}

func BalanceHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := daemon.Balance(r.Context(), []uint64{0})
	if err != nil {
		writeComplexError(w, err)
		return
	}
	b := balanceResponse{
		Total: resp.Balance,
		Unlocked: resp.UnlockedBalance,
	}
	json.NewEncoder(w).Encode(b)
}
