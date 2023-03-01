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
	"strconv"

	"github.com/go-chi/chi/v5"

	"gitlab.com/moneropay/moneropay/v2/internal/daemon"
	"gitlab.com/moneropay/moneropay/v2/model"
)

func ReceivePostHandler(w http.ResponseWriter, r *http.Request) {
	var j model.ReceivePostRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		writeError(w, http.StatusBadRequest, nil, err.Error())
		return
	}
	a, t, err := daemon.Receive(r.Context(), j.Amount, j.Description, j.CallbackUrl)
	if err != nil {
		writeComplexError(w, err)
		return
	}
	d := model.ReceivePostResponse{
		Address: a,
		Amount: j.Amount,
		Description: j.Description,
		CreatedAt: t,
	}
	json.NewEncoder(w).Encode(d)
}

func ReceiveGetHandler(w http.ResponseWriter, r *http.Request) {
	a := chi.URLParam(r, "address")
	// Parse optional GET parameters.
	var min, max uint64
	q := r.URL.Query()
	qmin, _ := q["min"]
	qmax, _ := q["max"]
	if len(qmin) > 0 && qmin[0] != "" {
		n, err := strconv.ParseUint(qmin[0], 10, 64)
		if err == nil {
			min = n
		}
	}
	if len(qmax) > 0 && qmax[0] != "" {
		n, err := strconv.ParseUint(qmax[0], 10, 64)
		if err == nil {
			max = n
			if max < min {
				writeError(w, http.StatusBadRequest, nil,
				    "Maximum block height cannot be lower than minimum")
				return
			}
		}
	}
	d, err := daemon.GetPaymentRequest(r.Context(), a, min, max)
	if err != nil {
		writeComplexError(w, err)
		return
	}
	json.NewEncoder(w).Encode(d)
}
