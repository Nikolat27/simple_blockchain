package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"simple_blockchain/pkg/CryptoGraphy"
	"simple_blockchain/pkg/utils"
)

const CoinbaseTxFee = 0

type Transaction struct {
	From       string `json:"from,omitempty"`
	To         string `json:"to"`
	Amount     uint64 `json:"amount"`
	Timestamp  int64  `json:"timestamp"`
	PublicKey  string // sender's public key (hex string)
	Signature  []byte // Ed25519 signature
	Fee        uint64 `json:"fee"`
	Status     string `json:"status"`
	IsCoinbase bool   `json:"is_coinbase"`
}

type TxHash []byte

func (hash TxHash) EncodeToString() string {
	return hex.EncodeToString(hash)
}

func (tx *Transaction) Hash() TxHash {
	txCopy := *tx

	txCopy.Signature = nil
	// Keep PublicKey for hashing as it's part of the transaction data

	data, _ := json.Marshal(txCopy)
	hash := sha256.Sum256(data)
	return hash[:]
}

func (tx *Transaction) Sign(keyPair *CryptoGraphy.KeyPair) error {
	tx.PublicKey = keyPair.GetPublicKeyHex() // Set PublicKey BEFORE hashing
	hash := tx.Hash()

	tx.Signature = keyPair.Sign(hash)
	return nil
}

func (tx *Transaction) SignWithHexKeys(privateKeyHex, publicKeyHex string) error {
	keyPair, err := CryptoGraphy.LoadKeyPairFromHex(privateKeyHex, publicKeyHex)
	if err != nil {
		return err
	}

	tx.PublicKey = keyPair.GetPublicKeyHex()
	hash := tx.Hash()
	tx.Signature = keyPair.Sign(hash)
	return nil
}

func (tx *Transaction) Verify() bool {
	if tx.Signature == nil || tx.PublicKey == "" {
		return false
	}

	hash := tx.Hash()
	return CryptoGraphy.VerifySignature(tx.PublicKey, hash, tx.Signature)
}

func (tx *Transaction) Size() int {
	size := 0

	// Strings: number of bytes in UTF-8 encoding
	size += len(tx.From)
	size += len(tx.To)
	size += len(tx.PublicKey)

	// Fixed-size fields
	size += 8 // Amount (uint64)
	size += 8 // Timestamp (int64)
	size += 4 // Fee (uint32)

	size += len(tx.Signature) // signature bytes

	// Bool and Status string
	size += 1 // IsCoinbase bool
	size += len(tx.Status)

	return size
}

func createCoinbaseTx(minerAddress string, miningReward uint64) *Transaction {
	return &Transaction{
		To:         minerAddress,
		Amount:     miningReward,
		Fee:        CoinbaseTxFee,
		Timestamp:  utils.GetTimestamp(),
		Status:     "confirmed",
		IsCoinbase: true,
	}
}
