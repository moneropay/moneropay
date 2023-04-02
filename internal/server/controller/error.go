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
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.com/moneropay/go-monero/walletrpc"
	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

func writeError(w http.ResponseWriter, status int, code *int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.ErrorResponse{
		Status: status,
		Code: code,
		Message: message,
	})
}

func writeComplexError(w http.ResponseWriter, err error) {
	if isWallet, werr := walletrpc.GetWalletError(err); isWallet {
		writeError(w, http.StatusInternalServerError,
			(*int)(&werr.Code), werr.Message)
	} else if errors.Is(err, context.DeadlineExceeded) {
		writeError(w, http.StatusGatewayTimeout, nil, err.Error())
	} else {
		writeError(w, http.StatusBadRequest, nil, err.Error())
	}
}
