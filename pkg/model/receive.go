package model

import "time"

type ReceivePostRequest struct {
	Amount uint64 `json:"amount"`
	CallbackUrl string `json:"callback_url"`
	Description string `json:"description,omitempty"`
}

type ReceivePostResponse struct {
	Address string `json:"address"`
	Amount uint64 `json:"amount"`
	Description string `json:"description,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type TransactionData struct {
	Amount uint64 `json:"amount"`
	Confirmations uint64 `json:"confirmations"`
	DoubleSpendSeen bool `json:"double_spend_seen"`
	Fee uint64 `json:"fee"`
	Height uint64 `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	TxHash string `json:"tx_hash"`
	UnlockTime uint64 `json:"unlock_time"`
	Locked bool `json:"locked"`
}

type ReceiveGetResponse struct {
	Amount struct {
		Expected uint64 `json:"expected"`
		Covered struct {
			Total uint64 `json:"total"`
			Unlocked uint64 `json:"unlocked"`
		} `json:"covered"`
	} `json:"amount"`
	Complete bool `json:"complete"`
	Description string `json:"description,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Transactions []TransactionData `json:"transactions"`
}
