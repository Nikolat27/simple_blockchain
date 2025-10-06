package handler

import (
	"net/http"

	"github.com/Nikolat27/simple_blockchain/pkg/CryptoGraphy"
	"github.com/Nikolat27/simple_blockchain/pkg/utils"
)

// GetBalance handles GET /api/balance requests.
// Returns the confirmed balance for a given wallet address, accounting for pending transactions.
//
// Query parameters:
//   - address: The wallet address to check (required)
//
// Response: 200 OK with JSON body:
//
//	{
//	  "balance": 10000  // Available balance in base units
//	}
//
// Response: 400 Bad Request if address parameter is missing
// Response: 500 Internal Server Error if balance lookup fails
func (handler *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		utils.WriteJSON(w, http.StatusBadRequest, "Address parameter required")
		return
	}

	balance, err := handler.Node.Blockchain.GetBalance(address)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err)
		return
	}

	resp := map[string]any{
		"balance": balance,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GenerateKeys handles POST /api/keys requests.
// Generates a new ECDSA key pair and derives a wallet address.
//
// Response: 200 OK with JSON body:
//
//	{
//	  "private_key": "hex",  // Private key (hex encoded) - keep secure!
//	  "public_key": "hex",   // Public key (hex encoded)
//	  "address": "address"   // Wallet address derived from public key
//	}
//
// Response: 500 Internal Server Error if key generation fails
func (handler *Handler) GenerateKeys(w http.ResponseWriter, r *http.Request) {
	keyPair, err := CryptoGraphy.GenerateKeyPair()
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
