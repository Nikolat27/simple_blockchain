package handler

import (
	"net/http"

	"github.com/Nikolat27/simple_blockchain/pkg/utils"
)

// GetBlockchain handles GET /api/chain requests.
// Returns the entire blockchain structure including all blocks in memory.
//
// Response: 200 OK with JSON body:
//
//	{
//	  "blockchain": {...}  // Complete blockchain object
//	}
func (handler *Handler) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"blockchain": handler.Node.Blockchain,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetMempool handles GET /api/mempool requests.
// Returns all pending transactions in the mempool waiting to be mined.
//
// Response: 200 OK with JSON body:
//
//	{
//	  "total": 5,         // Number of pending transactions
//	  "mempool": {...}    // Mempool object with transactions
//	}
func (handler *Handler) GetMempool(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"total":   len(handler.Node.Blockchain.Mempool.Transactions),
		"mempool": handler.Node.Blockchain.Mempool,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
