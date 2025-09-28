package handler

import (
	"net/http"
	"simple_blockchain/pkg/utils"
)

func (handler *Handler) MineBlock(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MinerAddress string `json:"miner_address"`
	}

	if err := utils.ParseJSON(r, 1_000, &input); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	ctx := r.Context()
	minedBlock := handler.Blockchain.MineBlock(ctx, handler.Blockchain.Mempool, input.MinerAddress)

	if minedBlock == nil {
		utils.WriteJSON(w, http.StatusBadRequest, "No transactions to mine")
		return
	}

	utils.WriteJSON(w, http.StatusOK, minedBlock)
}
