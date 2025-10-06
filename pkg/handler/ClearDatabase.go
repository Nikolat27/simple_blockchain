package handler

import (
	"net/http"

	"github.com/Nikolat27/simple_blockchain/pkg/utils"
)

// ClearDatabase handles DELETE /api/clear requests.
// Clears all blockchain data from the database including blocks, transactions, and balances.
// WARNING: This is a destructive operation that cannot be undone.
//
// Response: 200 OK with success message
// Response: 400 Bad Request if database clear fails
// Response: 500 Internal Server Error if transaction fails
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
