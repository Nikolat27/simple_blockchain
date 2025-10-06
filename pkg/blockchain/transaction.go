package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	"github.com/Nikolat27/simple_blockchain/pkg/CryptoGraphy"
	"github.com/Nikolat27/simple_blockchain/pkg/utils"
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
	var buf bytes.Buffer

	writeBytes := func(b []byte) {
		_ = binary.Write(&buf, binary.BigEndian, uint32(len(b)))
		buf.Write(b)
	}
	writeString := func(s string) {
		writeBytes([]byte(s))
	}

	// Variable-length strings (length-prefixed)
	writeString(tx.From)
	writeString(tx.To)
	writeString(tx.PublicKey)

	// Fixed-size numeric fields
	_ = binary.Write(&buf, binary.BigEndian, tx.Amount)    // uint64
	_ = binary.Write(&buf, binary.BigEndian, tx.Timestamp) // int64

	// Signature: length-prefixed bytes
	if tx.Signature != nil {
		writeBytes(tx.Signature)
	} else {
		_ = binary.Write(&buf, binary.BigEndian, uint32(0))
	}

	// Fee is uint64, write 8 bytes
	_ = binary.Write(&buf, binary.BigEndian, tx.Fee)

	// Status string
	writeString(tx.Status)

	// Bool as single byte
	if tx.IsCoinbase {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}

	return buf.Len()
}

func CreateCoinbaseTx(minerAddress string, miningReward uint64) *Transaction {
	return &Transaction{
		To:         minerAddress,
		Amount:     miningReward,
		Fee:        CoinbaseTxFee,
		Timestamp:  utils.GetTimestamp(),
		Status:     "confirmed",
		IsCoinbase: true,
	}
}

// NewTransaction -> It`s for unit tests
func NewTransaction(from, to string, amount uint64, timestamp int64) *Transaction {
	return &Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: timestamp,
		Status:    "pending",
	}
}
