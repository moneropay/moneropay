package model

type BalanceResponse struct {
	Total uint64 `json:"total"`
	Unlocked uint64 `json:"unlocked"`
}
