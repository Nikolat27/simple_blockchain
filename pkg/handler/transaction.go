package handler

import (
	"net/http"
	"simple_blockchain/pkg/CryptoGraphy"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/utils"
	"time"
)

func (handler *Handler) SendTransaction(w http.ResponseWriter, r *http.Request) {
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

	txFee := handler.Node.Blockchain.Mempool.CalculateFee(input.Amount)

	newTx := blockchain.Transaction{
		From:       input.From,
		To:         input.To,
		Amount:     input.Amount, // Full amount recipient receives
		Fee:        txFee,        // Fee paid to miner
		Status:     "pending",
		Timestamp:  time.Now().UTC().Unix(),
		IsCoinbase: false,
	}

	// Validate that the 'from' address matches the public key
	derivedAddress, err := CryptoGraphy.DeriveAddressFromPublicKey(input.PublicKey)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid public key format")
		return
	}

	if derivedAddress != input.From {
		utils.WriteJSON(w, http.StatusBadRequest,
			"'From' address does not match the provided public key")
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

	if err := handler.Node.Blockchain.ValidateTransaction(&newTx); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	handler.Node.Blockchain.Mempool.AddTransaction(&newTx)

	if err := handler.Node.BroadcastMempool(handler.Node.Blockchain.Mempool); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, err)
		return
	}

	resp := map[string]any{
		"transaction_hash": newTx.Hash().EncodeToString(),
		"message":          "Transaction added to mempool",
		"amount":           input.Amount,         // Amount recipient receives
		"fee":              txFee,                // Fee paid to miner
		"total_cost":       input.Amount + txFee, // Total cost to sender
		"status":           "pending",
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (handler *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	transactions := handler.Node.Blockchain.Mempool.GetTransactions()

	resp := map[string]any{
		"transactions": transactions,
		"count":        len(transactions),
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (handler *Handler) GetCurrentTxFee(w http.ResponseWriter, r *http.Request) {
	txFee := handler.Node.Blockchain.Mempool.CalculateTxFee()

	txFeePercentage := float64(txFee) / 100

	resp := map[string]any{
		"current_fee_percentage": txFeePercentage,
		"current_fee_basis":      txFee,
		"description":            "Fee is calculated as (amount * fee_basis) / 10000",
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
