/*
 * Copyright (C) 2023 Laurynas Četyrkinas <stnby@kernal.eu>
 * Copyright (C) 2023 İrem Kuyucu <siren@kernal.eu>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

import (
	"time"

	"gitlab.com/moneropay/go-monero/walletrpc"
)

type TransferPostRequest struct {
	Destinations []walletrpc.Destination `json:"destinations"`
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
