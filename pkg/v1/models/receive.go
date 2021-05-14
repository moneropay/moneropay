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

import "time"

type ReceivePostRequest struct {
	Amount uint64 `json:"amount"`
	Description string `json:"description"`
	//CallbackUrl *string `json:"callback_url"`
}

type ReceivePostResponse struct {
	Address string `json:"address"`
	Amount uint64 `json:"amount"`
	Description string `json:"description,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ReceiveTransaction struct {
	Amount uint64 `json:"amount"`
	Confirmations uint64 `json:"confirmations"`
	DoubleSpendSeen bool `json:"double_spend_seen"`
	Fee uint64 `json:"fee"`
	Height uint64 `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	TxHash string `json:"tx_hash"`
	UnlockTime uint64 `json:"unlock_time"`
}

type ReceiveGetResponse struct {
	Amount struct {
		Expected uint64 `json:"expected"`
		Covered uint64 `json:"covered"`
	} `json:"amount"`
	Complete bool `json:"complete"`
	Description string `json:"description,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Transactions []ReceiveTransaction `json:"transactions"`
}
