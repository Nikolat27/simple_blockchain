package handler

import (
	"net/http"
	"simple_blockchain/pkg/utils"
)

func (handler *Handler) GetAllBlocks(w http.ResponseWriter, r *http.Request) {
	blocks, err := handler.Blockchain.GetAllBlocks()
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]any{
		"blocks": blocks,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
