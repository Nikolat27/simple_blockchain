package handler

import (
	"net/http"
	"simple_blockchain/pkg/crypto"
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

func (handler *Handler) GenerateKeys(w http.ResponseWriter, r *http.Request) {
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, "Failed to generate key pair: "+err.Error())
		return
	}

	resp := map[string]string{
		"private_key": keyPair.GetPrivateKeyHex(),
		"public_key":  keyPair.GetPublicKeyHex(),
		"address":     keyPair.Address,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
