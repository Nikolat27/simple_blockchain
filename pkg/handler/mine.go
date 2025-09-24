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

	minedBlock := handler.Blockchain.MineBlock(handler.Mempool, input.MinerAddress)

	if minedBlock == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No transactions to mine"))
		return
	}

	// Broadcast the mined block to peers
	if handler.P2PNode != nil {
		handler.P2PNode.BroadcastBlock(minedBlock)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(minedBlock)
}
