package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"simple_blockchain/pkg/crypto"
	"time"
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

	tx.Timestamp = time.Now().Unix()

	data, _ := json.Marshal(txCopy)
	hash := sha256.Sum256(data)
	return hash[:]
}

func (tx *Transaction) Sign(keyPair *crypto.KeyPair) error {
	tx.PublicKey = keyPair.GetPublicKeyHex() // Set PublicKey BEFORE hashing
	hash := tx.hash()
	tx.Signature = keyPair.Sign(hash)
	return nil
}

func (tx *Transaction) SignWithHexKeys(privateKeyHex, publicKeyHex string) error {
	keyPair, err := crypto.LoadKeyPairFromHex(privateKeyHex, publicKeyHex)
	if err != nil {
		return err
	}

	tx.PublicKey = keyPair.GetPublicKeyHex() // Set PublicKey BEFORE hashing
	hash := tx.hash()
	tx.Signature = keyPair.Sign(hash)
	return nil
}

func (tx *Transaction) Verify() bool {
	if tx.Signature == nil || tx.PublicKey == "" {
		return false
	}

	hash := tx.hash()
	return crypto.VerifySignature(tx.PublicKey, hash, tx.Signature)
}

func (bc *Blockchain) ValidateTransaction(tx *Transaction) error {
	if tx.IsCoinbase {
		return nil
	}

	balance, err := bc.GetBalance(tx.From)
	if err != nil {
		return err
	}

	if balance >= tx.Amount {
		return nil
	}

	return errors.New("balance is insufficient")
}
