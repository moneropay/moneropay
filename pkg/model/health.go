package model

type HealthResponse struct {
	Status int `json:"status"`
	Services struct {
		WalletRPC bool `json:"walletrpc"`
		Database bool `json:"database"`
	} `json:"services"`
}
