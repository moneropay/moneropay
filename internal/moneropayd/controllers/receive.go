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
	"strconv"
	"time"
	"log"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"gitlab.com/moneropay/go-monero/walletrpc"

	"gitlab.com/moneropay/moneropay/internal/moneropayd/wallet"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/helpers"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/database"
	"gitlab.com/moneropay/moneropay/pkg/models"
)

func ReceivePostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	amount, err := strconv.ParseUint(r.FormValue("amount"), 10, 64)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, nil, err.Error())
		return
	}
	description := r.FormValue("description")
	if len(description) > 1024 {
		helpers.WriteError(w, http.StatusBadRequest, nil, "Description too long")
		return
	}
	callbackUrl := r.FormValue("callback_url")
	if len(callbackUrl) > 1024 {
		helpers.WriteError(w, http.StatusBadRequest, nil, "Callback_url too long")
		return
	}

	// Create a subaddress (blocking operation)
	ctx, cancel := context.WithTimeout(r.Context(), 3 * time.Second)
	wallet.Lock()
	resp, err := wallet.Wallet.CreateAddress(ctx, &walletrpc.CreateAddressRequest{})
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

	// Insert subaddress association to its index into the DB.
	var tx pgx.Tx
	ctx, cancel = context.WithTimeout(r.Context(), 4 * time.Second)
	go func() {
		defer cancel()
		tx, err = database.DB.Begin(ctx)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		if _, err = tx.Exec(ctx, "INSERT INTO subaddresses (index, address) VALUES ($1, $2)",
		    resp.AddressIndex, resp.Address); err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, nil, err.Error())
			tx.Rollback(ctx)
			return
		}
		if _, err = tx.Exec(ctx, "INSERT INTO receivers (subaddress_index, expected_amount, " +
		   "description, callback_url, created_at) VALUES ($1, $2, $3, $4, $5)", resp.AddressIndex, amount,
		   description, callbackUrl, time.Now()); err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, nil, err.Error())
			tx.Rollback(ctx)
			return
		}
		if err = tx.Commit(ctx); err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
	}()
	<-ctx.Done()
	if err != nil {
		return
	}
	if ret := ctx.Err(); ret != context.Canceled {
		helpers.WriteError(w, http.StatusGatewayTimeout, nil, "Context timeout exceeded")
		return
	}

	d := models.ReceivePostResponse{
		Address: resp.Address,
		Amount: amount,
		Description: description,
		CreatedAt: time.Now(),
	}
	json.NewEncoder(w).Encode(d)
}

func ReceiveGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	address := mux.Vars(r)["address"]
	var addressIndex uint64
	var d models.ReceiveGetResponse

	// Find 'subaddress_index' in the DB by subaddress.
	row, err := database.QueryRowWithTimeout(r.Context(), 3 * time.Second,
	    "SELECT index, expected_amount, description, created_at " +
	    "FROM subaddresses, receivers " +
	    "WHERE index=subaddress_index AND address=$1", address)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, nil, err.Error())
		return
	}
	err = row.Scan(&addressIndex, &d.Amount.Expected, &d.Description, &d.CreatedAt)
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, nil, err.Error())
		return
	}
	// Get all transfer done on the subaddress (blocking operation)
	ctx, cancel := context.WithTimeout(r.Context(), 3 * time.Second)
	wallet.Lock()
	resp, err := wallet.Wallet.GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		SubaddrIndices: []uint64{addressIndex},
		In: true,
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
	for _, r1 := range resp.In {
		r2 := models.ReceiveTransaction{
			Amount: r1.Amount,
			Confirmations: r1.Confirmations,
			DoubleSpendSeen: r1.DoubleSpendSeen,
			Fee: r1.Fee,
			Height: r1.Height,
			Timestamp: time.Unix(int64(r1.Timestamp), 0),
			TxHash: r1.Txid,
			UnlockTime: r1.UnlockTime,
		}
		if r1.DoubleSpendSeen == true {
			log.Println("Double spend attempt detected:", r1.Txid, r1.Address, r1.Timestamp,)
			continue
		}
		if r1.UnlockTime == 0 {
			d.Amount.Covered.Unlocked += r1.Amount
		} else {
			d.Amount.Covered.Locked += r1.Amount
		}
		d.Transactions = append(d.Transactions, r2)
	}
	d.Amount.Covered.Total = d.Amount.Covered.Unlocked + d.Amount.Covered.Locked
	d.Complete = d.Amount.Covered.Unlocked >= d.Amount.Expected
	json.NewEncoder(w).Encode(d)
}
