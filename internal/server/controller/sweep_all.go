package controller

import (
	"encoding/json"
	"net/http"

	"gitlab.com/moneropay/go-monero/walletrpc"

	"gitlab.com/moneropay/moneropay/v2/internal/daemon"
)

type sweepAllPostRequest struct {
	Address string `json:"address"`
}

func SweepAllPostHandler(w http.ResponseWriter, r *http.Request) {
	var j sweepAllPostRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		writeError(w, http.StatusBadRequest, nil, err.Error())
		return
	}
	resp, err := daemon.SweepAll(r.Context(), &walletrpc.SweepAllRequest{Address: j.Address})
	if err != nil {
		writeComplexError(w, err)
		return
	}
	json.NewEncoder(w).Encode(resp)
}
