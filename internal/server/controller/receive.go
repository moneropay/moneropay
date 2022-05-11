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
	"time"

	"github.com/go-chi/chi/v5"

	"gitlab.com/moneropay/moneropay/v2/internal/daemon"
)

type receivePostRequest struct {
	Amount uint64 `json:"amount"`
	CallbackUrl string `json:"callback_url"`
	Description string `json:"description,omitempty"`
}

type receivePostResponse struct {
	Address string `json:"address"`
	Amount uint64 `json:"amount"`
	Description string `json:"description,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type receiveGetResponse struct {
	Amount struct {
		Expected uint64 `json:"expected"`
		Covered struct {
			Total uint64 `json:"total"`
			Unlocked uint64 `json:"unlocked"`
		} `json:"covered"`
	} `json:"amount"`
	Complete bool `json:"complete"`
	Description string `json:"description,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Transactions []daemon.ReceiveTransaction `json:"transactions"`
}

func ReceivePostHandler(w http.ResponseWriter, r *http.Request) {
	var j receivePostRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		writeError(w, http.StatusBadRequest, nil, err.Error())
		return
	}
	a, t, err := daemon.Receive(r.Context(), j.Amount, j.Description, j.CallbackUrl)
	if err != nil {
		writeComplexError(w, err)
		return
	}
	d := receivePostResponse{
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
				writeError(w, http.StatusBadRequest, nil, "Maximum block height cannot be lower than minimum")
				return
			}
		}
	}
	ctx := r.Context()
	// Get data for address from DB.
	recv, err := daemon.GetReceiver(ctx, a)
	if err != nil {
		writeError(w, http.StatusInternalServerError, nil, err.Error())
		return
	}
	var d receiveGetResponse
	d.Amount.Expected = recv.Expected; d.Description = recv.Description; d.CreatedAt = recv.CreatedAt
	// Get balance for address from WalletRPC.
	resp, err := daemon.Balance(ctx, []uint64{recv.Index})
	if err != nil {
		writeComplexError(w, err)
		return
	}
	d.Amount.Covered.Total = resp.PerSubaddress[0].Balance
	d.Amount.Covered.Unlocked = resp.PerSubaddress[0].UnlockedBalance
	d.Complete = d.Amount.Covered.Unlocked >= d.Amount.Expected
	tx, err := daemon.GetReceived(ctx, recv.Index, min, max)
	if err != nil {
		writeError(w, http.StatusInternalServerError, nil, err.Error())
		return
	}
	for _, r1 := range tx {
		r2 := daemon.ReceiveTransaction{
			Amount: r1.Amount,
			Confirmations: r1.Confirmations,
			DoubleSpendSeen: r1.DoubleSpendSeen,
			Fee: r1.Fee,
			Height: r1.Height,
			Timestamp: time.Unix(int64(r1.Timestamp), 0),
			TxHash: r1.Txid,
			UnlockTime: r1.UnlockTime,
		}
		d.Transactions = append(d.Transactions, r2)
	}
	json.NewEncoder(w).Encode(d)
}
