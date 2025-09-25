package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/crypto"
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

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if input.Amount == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Your transaction amount must be more than 0"))
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid public key format"))
		return
	}

	if derivedAddress != input.From {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("From address does not match the provided public key"))
		return
	}

	// Sign the transaction with the provided keys
	err = newTx.SignWithHexKeys(input.PrivateKey, input.PublicKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to sign transaction: " + err.Error()))
		return
	}

	if !newTx.Verify() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid signature"))
		return
	}

	if !handler.Blockchain.ValidateTransaction(&newTx) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Insufficient balance"))
		return
	}

	handler.Mempool.AddTransaction(&newTx)

	if handler.P2PNode != nil {
		handler.P2PNode.BroadcastTransaction(&newTx)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transaction added to mempool",
		"status":  "pending",
	})
}

func (handler *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	transactions := handler.Mempool.GetTransactions()

	fmt.Println(transactions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"transactions": transactions,
		"count":        len(transactions),
	})
}
