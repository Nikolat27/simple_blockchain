package handler

import (
	"encoding/json"
	"net/http"
)

func (handler *Handler) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(handler.Blockchain)
}
