package handler

import (
	"fmt"
	"net/http"

	"github.com/Nikolat27/simple_blockchain/pkg/utils"
)

// MineBlock handles POST /api/mine requests.
// Mines a new block containing pending transactions from the mempool.
// Uses proof-of-work to find a valid block hash and broadcasts it to the network.
//
// Request body (JSON):
//
//	{
//	  "miner_address": "address"  // Address to receive mining reward
//	}
//
// Response: 200 OK with mined block object
// Response: 400 Bad Request if mining fails or validation errors occur
// Response: 409 Conflict if mining was cancelled or block already mined
// Response: 500 Internal Server Error if broadcast fails
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
	}

	utils.WriteJSON(w, http.StatusOK, minedBlock)
}
