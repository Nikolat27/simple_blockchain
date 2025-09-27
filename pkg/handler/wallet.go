package handler

import (
	"net/http"
	"simple_blockchain/pkg/crypto"
	"simple_blockchain/pkg/utils"
)

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
