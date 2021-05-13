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

package models

import (
	"time"

	"gitlab.com/kernal/go-monero/walletrpc"
)

type TransferPostRequest struct {
	Destinations []walletrpc.Destination `json:"destinations"`
	// CallbackUrl *string `json:"callback_url"`
}

type TransferPostResponse struct {
	Amount uint64 `json:"amount"`
	Fee uint64 `json:"fee"`
	TxHash string `json:"tx_hash"`
	Destinations []walletrpc.Destination `json:"destinations"`
}

type TransferGetResponse struct {
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
