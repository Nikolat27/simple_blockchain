package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"simple_blockchain/pkg/database"
	"strings"
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

	// Create a deterministic string representation of transactions
	txData, err := json.Marshal(block.Transactions)
	if err != nil {
		return err
	}

	record := fmt.Sprintf("%d-%s-%d-%d-%s", block.Id, prevHashStr,
		block.Timestamp, block.Nonce, string(txData))

	hash := sha256.Sum256([]byte(record))

	block.Hash = hash[:]

	return nil
}

func (block *Block) IsValidHash() bool {
	hashStr := hex.EncodeToString(block.Hash)

	return strings.HasPrefix(hashStr, strings.Repeat("0", Difficulty))
}

// parseDBTransactions -> Convert DB transactions to blockchain transactions
func (block *Block) parseDBTransactions(dbTxs []database.DBTransactionSchema) {
	block.Transactions = make([]Transaction, len(dbTxs))

	for idx, dbTx := range dbTxs {
		block.Transactions[idx] = Transaction{
			From:       dbTx.From,
			To:         dbTx.To,
			Amount:     dbTx.Amount,
			Fee:        dbTx.Fee,
			Timestamp:  dbTx.Timestamp,
			PublicKey:  dbTx.PublicKey,
			Signature:  dbTx.Signature,
			Status:     dbTx.Status,
			IsCoinbase: dbTx.IsCoinbase,
		}
	}
}

func createGenesisBlock() (*Block, error) {
	block := &Block{
		Id:           0,
		PrevHash:     make([]byte, 32),
		Timestamp:    1609459200, // Fixed timestamp: 2021-01-01 00:00:00 UTC
		Transactions: []Transaction{},
		Nonce:        0,
	}

	if err := block.HashBlock(); err != nil {
		return nil, err
	}

	return block, nil
}
