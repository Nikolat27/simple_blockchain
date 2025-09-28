package handler

import (
	"net/http"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/crypto"
	"simple_blockchain/pkg/utils"
	"time"
)

func (handler *Handler) AddTransaction(w http.ResponseWriter, r *http.Request) {
	var input struct {
		From       string `json:"from"`
		To         string `json:"to"`
		Amount     uint64 `json:"amount"`
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
	}

	if err := utils.ParseJSON(r, 10_000, &input); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if input.Amount == 0 {
		utils.WriteJSON(w, http.StatusBadRequest, "Your transaction amount must be more than 0")
		return
	}

	newTx := blockchain.Transaction{
		From:       input.From,
		To:         input.To,
		Amount:     input.Amount,
		Status:     "pending",
		Timestamp:  time.Now().UTC().Unix(),
		IsCoinbase: false,
	}

	// Validate that the from address matches the public key
	derivedAddress, err := crypto.DeriveAddressFromPublicKey(input.PublicKey)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid public key format")
		return
	}

	if derivedAddress != input.From {
		utils.WriteJSON(w, http.StatusBadRequest, "'From' address does not match the provided public key")
		return
	}

	// Sign the transaction with the provided keys
	if err := newTx.SignWithHexKeys(input.PrivateKey, input.PublicKey); err != nil {
	 	utils.WriteJSON(w, http.StatusBadRequest, "Failed to sign transaction: "+err.Error())
	 	return
	}

	if !newTx.Verify() {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	if err := handler.Blockchain.ValidateTransaction(&newTx); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	handler.Blockchain.Mempool.AddTransaction(&newTx)

	resp := map[string]string{
		"message": "Transaction added to mempool",
		"status":  "pending",
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (handler *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	transactions := handler.Blockchain.Mempool.GetTransactions()

	resp := map[string]any{
		"transactions": transactions,
		"count":        len(transactions),
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
