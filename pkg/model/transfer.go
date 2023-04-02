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
