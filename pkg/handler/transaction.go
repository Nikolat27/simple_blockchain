package handler

import (
	"net/http"

	"github.com/Nikolat27/simple_blockchain/pkg/CryptoGraphy"
	"github.com/Nikolat27/simple_blockchain/pkg/blockchain"
	"github.com/Nikolat27/simple_blockchain/pkg/utils"
)

// SendTransaction handles POST /api/tx/send requests.
// Creates, signs, validates, and broadcasts a new transaction to the network.
//
// Request body (JSON):
//
//	{
//	  "from": "address",        // Sender's wallet address
//	  "to": "address",          // Recipient's wallet address
//	  "amount": 1000,           // Amount to transfer
//	  "private_key": "hex",     // Sender's private key (hex encoded)
//	  "public_key": "hex"       // Sender's public key (hex encoded)
//	}
//
// Response: 200 OK with JSON body:
//
//	{
//	  "transaction_hash": "hex",  // Hash of the transaction
//	  "message": "...",           // Success message
//	  "amount": 1000,             // Amount recipient receives
//	  "fee": 10,                  // Fee paid to miner
//	  "total_cost": 1010,         // Total deducted from sender
//	  "status": "pending"
//	}
//
// Response: 400 Bad Request if validation fails or insufficient balance
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

	newTx := blockchain.Transaction{
		From:       input.From,
		To:         input.To,
		Amount:     input.Amount, // Full amount recipient receives
		Status:     "pending",
		Timestamp:  utils.GetTimestamp(),
		IsCoinbase: false,
	}

	txFee := handler.Node.Blockchain.Mempool.CalculateFee(&newTx)

	newTx.Fee = txFee

	if handler.Node.Blockchain.Mempool.WillExceedCapacity(&newTx) {
		utils.WriteJSON(w, http.StatusBadRequest, "Mempool capacity exceeded. Try again later")
		return
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

// GetTransactions handles GET /api/txs requests.
// Returns all pending transactions currently in the mempool.
//
// Response: 200 OK with JSON body:
//
//	{
//	  "transactions": {...},  // Map of transaction hash to transaction object
//	  "count": 5              // Number of pending transactions
//	}
func (handler *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	transactions := handler.Node.Blockchain.Mempool.GetTransactionsCopy()

	resp := map[string]any{
		"transactions": transactions,
		"count":        len(transactions),
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetCurrentTxFee handles GET /api/tx/fee requests.
// Returns the current transaction fee rate based on mempool congestion.
//
// Response: 200 OK with JSON body:
//
//	{
//	  "current_fee_percentage": 0.1,  // Fee as percentage (e.g., 0.1 = 0.1%)
//	  "current_fee_basis": 10,        // Fee basis points (e.g., 10 = 0.1%)
//	  "description": "..."            // Explanation of fee calculation
//	}
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
