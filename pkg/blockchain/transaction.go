package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"simple_blockchain/pkg/crypto"
)

type Transaction struct {
	From       string `json:"from,omitempty"`
	To         string `json:"to"`
	Amount     uint64 `json:"amount"`
	Timestamp  int64
	PublicKey  string // sender's public key (hex string)
	Signature  []byte // Ed25519 signature
	Status     string `json:"status"`
	IsCoinbase bool   `json:"is_coinbase"`
}

func (tx *Transaction) hash() []byte {
	txCopy := *tx

	txCopy.Signature = nil
	// Keep PublicKey for hashing as it's part of the transaction data

	data, _ := json.Marshal(txCopy)
	hash := sha256.Sum256(data)
	return hash[:]
}

func (tx *Transaction) Sign(keyPair *crypto.KeyPair) error {
	hash := tx.hash()
	tx.Signature = keyPair.Sign(hash)
	tx.PublicKey = keyPair.GetPublicKeyHex()
	return nil
}

func (tx *Transaction) SignWithHexKeys(privateKeyHex, publicKeyHex string) error {
	keyPair, err := crypto.LoadKeyPairFromHex(privateKeyHex, publicKeyHex)
	if err != nil {
		return err
	}

	hash := tx.hash()
	tx.Signature = keyPair.Sign(hash)
	tx.PublicKey = keyPair.GetPublicKeyHex()
	return nil
}

func (tx *Transaction) Verify() bool {
	if tx.Signature == nil || tx.PublicKey == "" {
		return false
	}

	hash := tx.hash()
	return crypto.VerifySignature(tx.PublicKey, hash, tx.Signature)
}
