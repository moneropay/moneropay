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

	"github.com/gorilla/mux"
	"gitlab.com/moneropay/go-monero/walletrpc"

	"gitlab.com/moneropay/moneropay/internal/moneropayd/config"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/helpers"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/wallet"
	"gitlab.com/moneropay/moneropay/pkg/models"
)

func TransferPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Decode json input.
	var j models.TransferPostRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	// Do a transfer (blocking operation)
	wallet.Lock()
	ctx, cancel := context.WithTimeout(r.Context(), 3 * time.Second)
	resp, err := wallet.Wallet.Transfer(ctx, &walletrpc.TransferRequest{
		Destinations: j.Destinations,
		Priority: walletrpc.Priority(config.Values.TransferPriority),
		Mixin: config.Values.TransferMixin,
		UnlockTime: config.Values.TransferUnlockTime,
	})
	defer cancel()
	wallet.Unlock()
	if err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			helpers.WriteError(w, http.StatusInternalServerError,
				(*int)(&werr.Code), werr.Message)
		} else if cerr := ctx.Err(); cerr != nil {
			helpers.WriteError(w, http.StatusGatewayTimeout, nil, cerr.Error())
		} else {
			helpers.WriteError(w, http.StatusBadRequest, nil, err.Error())
		}
		return
	}

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

	// Get information about transfer (blocking operation)
	ctx, cancel := context.WithTimeout(r.Context(), 3 * time.Second)
	wallet.Lock()
	resp, err := wallet.Wallet.GetTransferByTxid(ctx, &walletrpc.GetTransferByTxidRequest{
		Txid: mux.Vars(r)["tx_hash"],
	})
	defer cancel()
	wallet.Unlock()
	if err != nil {
		if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
			helpers.WriteError(w, http.StatusInternalServerError,
				(*int)(&werr.Code), werr.Message)
		} else if cerr := ctx.Err(); cerr != nil {
			helpers.WriteError(w, http.StatusGatewayTimeout, nil, cerr.Error())
		} else {
			helpers.WriteError(w, http.StatusBadRequest, nil, err.Error())
		}
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
