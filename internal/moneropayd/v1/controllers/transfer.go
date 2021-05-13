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
	"time"

	"github.com/gorilla/mux"
	"gitlab.com/kernal/go-monero/walletrpc"

	"gitlab.com/kernal/moneropay/internal/moneropayd/config"
	"gitlab.com/kernal/moneropay/internal/moneropayd/v1/helpers"
	"gitlab.com/kernal/moneropay/internal/moneropayd/wallet"
	"gitlab.com/kernal/moneropay/pkg/v1/models"
)

func TransferPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	rpc := wallet.Wallet
	var j models.TransferPostRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, nil, err.Error())
		return
	}
	wallet.Lock()
	resp, err := rpc.Transfer(&walletrpc.TransferRequest{
		Destinations: j.Destinations,
		Priority: walletrpc.Priority(config.Values.TransferPriority),
		Mixin: config.Values.TransferMixin,
		UnlockTime: config.Values.TransferUnlockTime,
	})
	wallet.Unlock()
	if err != nil {
		_, werr := walletrpc.GetWalletError(err)
		helpers.WriteError(w, http.StatusInternalServerError, (*int)(&werr.Code), werr.Message)
		return
	}
	// TODO: Save the callback to the database and handle it.
	d := models.TransferPostResponse{
		Amount: resp.Amount,
		Fee: resp.Fee,
		TxHash: resp.TxHash,
		Destinations: j.Destinations,
	}
	json.NewEncoder(w).Encode(d)
}

func TransferGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	rpc := wallet.Wallet
	wallet.Lock()
	resp, err := rpc.GetTransferByTxid(&walletrpc.GetTransferByTxidRequest{
		Txid: mux.Vars(r)["tx_hash"],
	})
	wallet.Unlock()
	if err != nil {
		_, werr := walletrpc.GetWalletError(err)
		helpers.WriteError(w, http.StatusInternalServerError, (*int)(&werr.Code), werr.Message)
		return
	}
	if (resp.Transfer.Type == "in") {
		helpers.WriteError(w, http.StatusBadRequest, nil, "Not an outgoing transaction")
		return
	}
	if (resp.Transfer.Type == "out") {
		resp.Transfer.Type = "completed"
	}
	d := models.TransferGetResponse{
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
