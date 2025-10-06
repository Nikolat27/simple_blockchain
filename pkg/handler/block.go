package handler

import (
	"net/http"

	"github.com/Nikolat27/simple_blockchain/pkg/utils"
)

// GetAllBlocks handles GET /api/blocks requests.
// Returns all blocks in the blockchain with their transactions.
//
// Response: 200 OK with JSON body:
//
//	{
//	  "blocks": [...]  // Array of block objects
//	}
//
// Response: 500 Internal Server Error if database query fails
func (handler *Handler) GetAllBlocks(w http.ResponseWriter, r *http.Request) {
	blocks, err := handler.Node.Blockchain.GetAllBlocks()
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]any{
		"blocks": blocks,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
