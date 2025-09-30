package handler

import (
	"net/http"
	"simple_blockchain/pkg/utils"
)

func (handler *Handler) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"blockchain": handler.Node.Blockchain,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
