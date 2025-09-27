package handler

import (
	"net/http"
	"simple_blockchain/pkg/utils"
)

func (handler *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		utils.WriteJSON(w, http.StatusBadRequest, "Address parameter required")
		return
	}

	balance, err := handler.Blockchain.GetBalance(address)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err)
		return
	}

	resp := map[string]any{
		"balance": balance,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
