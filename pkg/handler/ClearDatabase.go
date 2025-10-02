package handler

import (
	"net/http"
	"simple_blockchain/pkg/utils"
)

func (handler *Handler) ClearDatabase(w http.ResponseWriter, r *http.Request) {
	sqlTx, err := handler.Node.Blockchain.Database.BeginTx()
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err)
		return
	}
	defer sqlTx.Rollback()

	if err := handler.Node.Blockchain.Database.ClearAllData(sqlTx); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	if err := sqlTx.Commit(); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, "database got cleared successfully")
}
