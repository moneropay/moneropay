package model

type HealthResponse struct {
	Status int `json:"status"`
	Services struct {
		WalletRPC bool `json:"walletrpc"`
		PostgreSQL bool `json:"postgresql"`
	} `json:"services"`
}
