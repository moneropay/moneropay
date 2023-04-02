package model

type ErrorResponse struct {
	Status int `json:"status"`
	Code *int `json:"code,omitempty"`
	Message string `json:"message"`
}
