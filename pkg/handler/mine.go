package handler

import (
	"fmt"
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
	minedBlock, err := handler.Node.Blockchain.MineBlock(ctx, handler.Node.Blockchain.Mempool, input.MinerAddress)

	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	// Cancelled or mined before
	if minedBlock == nil {
		utils.WriteJSON(w, http.StatusConflict, "Transaction was either cancelled or already mined")
		return
	}

	if err := handler.Node.BroadcastBlock(minedBlock); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, fmt.Errorf("failed to broadcast block: %v", err))
		return
	}

	if err := handler.Node.CancelMining(); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, fmt.Errorf("failed to broadcast block: %v", err))
		return
	} else {
		fmt.Println("block mined successfully")
	}

	utils.WriteJSON(w, http.StatusOK, minedBlock)
}
