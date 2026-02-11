/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2022 İrem Kuyucu <siren@kernal.eu>
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
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

	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

// TransferPostHandler sends funds to one or more destinations.
func (c *Controller) TransferPostHandler(w http.ResponseWriter, r *http.Request) {
	if c.Daemon.IsViewOnly() {
		writeError(w, http.StatusForbidden, nil, "This is a view-only wallet")
		return
	}
	var j model.TransferPostRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		writeError(w, http.StatusBadRequest, nil, err.Error())
		return
	}
	resp, err := c.Daemon.TransferSplit(r.Context(), &walletrpc.TransferSplitRequest{
		Destinations: j.Destinations,
		Priority:     walletrpc.Priority(c.Daemon.Config().TransferPriority),
	})
	if err != nil {
		writeComplexError(w, err)
		return
	}
	var amount uint64 = 0
	var fee uint64 = 0
	for i, amt := range resp.AmountList {
		amount += uint64(amt)
		fee += uint64(resp.FeeList[i])
	}
	d := model.TransferPostResponse{
		Amount:       amount,
		Fee:          fee,
		TxHash:       resp.TxHashList[0],
		TxHashList:   resp.TxHashList,
		Destinations: j.Destinations,
	}
	json.NewEncoder(w).Encode(d)
}

// TransferGetHandler returns the status of an outgoing transfer.
func (c *Controller) TransferGetHandler(w http.ResponseWriter, r *http.Request) {
	txHash := chi.URLParam(r, "tx_hash")
	resp, err := c.Daemon.GetTransferByTxid(r.Context(), &walletrpc.GetTransferByTxidRequest{
		Txid: txHash,
	})
	if err != nil {
		writeComplexError(w, err)
		return
	}
	if resp.Transfer.Type == "in" {
		writeError(w, http.StatusBadRequest, nil, "Not an outgoing transaction")
		return
	}
	if resp.Transfer.Type == "out" {
		resp.Transfer.Type = "completed"
	}

	d := model.TransferGetResponse{
		Amount:          resp.Transfer.Amount,
		Fee:             resp.Transfer.Fee,
		State:           resp.Transfer.Type,
		Destinations:    resp.Transfer.Destinations,
		Confirmations:   resp.Transfer.Confirmations,
		DoubleSpendSeen: resp.Transfer.DoubleSpendSeen,
		Height:          resp.Transfer.Height,
		Timestamp:       time.Unix(int64(resp.Transfer.Timestamp), 0),
		UnlockTime:      resp.Transfer.UnlockTime,
		TxHash:          resp.Transfer.Txid,
	}
	json.NewEncoder(w).Encode(d)
}
