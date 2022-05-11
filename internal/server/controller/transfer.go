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
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/moneropay/go-monero/walletrpc"

	"gitlab.com/moneropay/moneropay/v2/internal/daemon"
)

type transferPostRequest struct {
	Destinations []walletrpc.Destination `json:"destinations"`
}

type transferPostResponse struct {
	Amount uint64 `json:"amount"`
	Fee uint64 `json:"fee"`
	TxHash string `json:"tx_hash"`
	Destinations []walletrpc.Destination `json:"destinations"`
}

type transferGetResponse struct {
	Amount uint64 `json:"amount"`
	Fee uint64 `json:"fee"`
	State string `json:"state"`
	Destinations []walletrpc.Destination `json:"transfer"`
	Confirmations uint64 `json:"confirmations"`
	DoubleSpendSeen bool `json:"double_spend_seen"`
	Height uint64 `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	UnlockTime uint64 `json:"unlock_time"`
	TxHash string `json:"tx_hash"`
}

func TransferPostHandler(w http.ResponseWriter, r *http.Request) {
	var j transferPostRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		writeError(w, http.StatusBadRequest, nil, err.Error())
		return
	}
	resp, err := daemon.Transfer(r.Context(), &walletrpc.TransferRequest{
		Destinations: j.Destinations,
		Priority: walletrpc.Priority(daemon.Config.TransferPriority),
		Mixin: daemon.Config.TransferMixin,
		UnlockTime: daemon.Config.TransferUnlockTime,
	})
	if err != nil {
		writeComplexError(w, err)
		return
	}
	d := transferPostResponse{
		Amount: resp.Amount,
		Fee: resp.Fee,
		TxHash: resp.TxHash,
		Destinations: j.Destinations,
	}
	json.NewEncoder(w).Encode(d)
}

func TransferGetHandler(w http.ResponseWriter, r *http.Request) {
	txHash := chi.URLParam(r, "tx_hash")
	resp, err := daemon.GetTransferByTxid(r.Context(), &walletrpc.GetTransferByTxidRequest{
		Txid: txHash,
	})
	if err != nil {
		writeComplexError(w, err)
		return
	}
	if (resp.Transfer.Type == "in") {
		writeError(w, http.StatusBadRequest, nil, "Not an outgoing transaction")
		return
	}
	if (resp.Transfer.Type == "out") {
		resp.Transfer.Type = "completed"
	}

	d := transferGetResponse{
		Amount: resp.Transfer.Amount,
		Fee: resp.Transfer.Fee,
		State: resp.Transfer.Type,
		Destinations: resp.Transfer.Destinations,
		Confirmations: resp.Transfer.Confirmations,
		DoubleSpendSeen: resp.Transfer.DoubleSpendSeen,
		Height: resp.Transfer.Height,
		Timestamp: time.Unix(int64(resp.Transfer.Timestamp), 0),
		UnlockTime: resp.Transfer.UnlockTime,
		TxHash: resp.Transfer.Txid,
	}
	json.NewEncoder(w).Encode(d)
}
