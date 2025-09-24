package handler

import (
	"encoding/json"
	"net/http"
)

func (handler *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Address parameter required"))
		return
	}

	balance := handler.Blockchain.GetBalance(address)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"address": address,
		"balance": balance,
	})
}
