package handler

import (
	"encoding/json"
	"net/http"
	"simple_blockchain/pkg/crypto"
)

func (handler *Handler) GenerateKeys(w http.ResponseWriter, r *http.Request) {
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to generate key pair: " + err.Error()))
		return
	}

	resp := map[string]string{
		"private_key": keyPair.GetPrivateKeyHex(),
		"public_key":  keyPair.GetPublicKeyHex(),
		"address":     keyPair.Address,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
