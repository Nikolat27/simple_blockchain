package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"simple_blockchain/pkg/database"
	"strings"
	"time"
)

type Block struct {
	Id           int64         `json:"id"`
	PrevHash     []byte        `json:"prev_hash"`
	Hash         []byte        `json:"hash"`
	Timestamp    int64         `json:"timestamp"`
	Nonce        int64         `json:"nonce"`
	Transactions []Transaction `json:"transactions,omitempty"`
}

func (block *Block) HashBlock() error {
	prevHashStr := hex.EncodeToString(block.PrevHash)

	txData := block.SerializeTransactions()

	record := fmt.Sprintf("%d-%s-%d-%d-%s", block.Id, prevHashStr,
		block.Timestamp, block.Nonce, txData)

	hash := sha256.Sum256([]byte(record))

	block.Hash = hash[:]

	return nil
}

func (block *Block) IsValidHash() bool {
	hashStr := hex.EncodeToString(block.Hash)

	return strings.HasPrefix(hashStr, strings.Repeat("0", Difficulty))
}

// parseDBTransactions -> Convert DB transactions to blockchain transactions
func (block *Block) parseDBTransactions(dbTxs []database.DBTransactionSchema) error {
	block.Transactions = make([]Transaction, len(dbTxs), len(dbTxs))

	for idx, dbTx := range dbTxs {
		decodedSignature, err := hex.DecodeString(dbTx.Signature)
		if err != nil {
			return err
		}

		block.Transactions[idx] = Transaction{
			From:       dbTx.From,
			To:         dbTx.To,
			Amount:     dbTx.Amount,
			Fee:        dbTx.Fee,
			Timestamp:  dbTx.Timestamp,
			PublicKey:  dbTx.PublicKey,
			Signature:  decodedSignature,
			Status:     dbTx.Status,
			IsCoinbase: dbTx.IsCoinbase,
		}
	}

	return nil
}

// SerializeTransactions -> Deterministic
func (block *Block) SerializeTransactions() string {
	txs := block.Transactions

	parts := make([]string, len(txs))

	for i, tx := range txs {
		parts[i] = fmt.Sprintf("%s-%s-%d-%x",
			tx.From, tx.To, tx.Amount, tx.Signature)
	}

	return strings.Join(parts, "|")
}

func createGenesisBlock() (*Block, error) {
	block := &Block{
		Id:           0,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		Nonce:        0,
	}

	if err := block.HashBlock(); err != nil {
		return nil, err
	}

	return block, nil
}
