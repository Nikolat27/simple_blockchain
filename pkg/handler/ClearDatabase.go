package handler

import (
	"net/http"
	"simple_blockchain/pkg/utils"
)

func (handler *Handler) ClearDatabase(w http.ResponseWriter, r *http.Request) {
	if err := handler.Node.Blockchain.Database.ClearAllData(); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "database got cleared successfully")
}
