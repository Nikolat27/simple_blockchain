package handler

import (
	"encoding/json"
	"net/http"
)

func (handler *Handler) MineBlock(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MinerAddress string `json:"miner_address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	ctx := r.Context()
	minedBlock := handler.Blockchain.MineBlock(ctx, handler.Mempool, input.MinerAddress)

	if minedBlock == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No transactions to mine"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(minedBlock)
}
