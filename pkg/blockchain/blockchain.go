package blockchain

import (
	"bytes"
	"fmt"
	"simple_blockchain/pkg/database"
	"sync"
)

const MiningReward = 5
const Difficulty = 5

type Blockchain struct {
	Blocks []Block `json:"blocks"`
	mu     sync.RWMutex

	Database *database.Database
}

func NewBlockchain(db *database.Database) *Blockchain {
	bc := &Blockchain{
		Blocks:   make([]Block, 0),
		Database: db,
	}

	genesisBlock := createGenesisBlock()

	bc.Blocks = append(bc.Blocks, *genesisBlock)

	return bc
}

func (bc *Blockchain) AddBlock(newBlock *Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.Blocks = append(bc.Blocks, *newBlock)
}

func (bc *Blockchain) GetBlocks() []Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	blocksCopy := make([]Block, len(bc.Blocks))
	copy(blocksCopy, bc.Blocks)

	return blocksCopy
}

func (bc *Blockchain) VerifyBlock(block *Block) bool {
	tempBlock := &Block{
		Index:        block.Index,
		PrevHash:     block.PrevHash,
		Timestamp:    block.Timestamp,
		Transactions: block.Transactions,
		Nonce:        block.Nonce,
		Hash:         nil,
	}

	tempBlock.HashBlock()

	hashMatches := bytes.Equal(tempBlock.Hash, block.Hash)

	return hashMatches && tempBlock.IsValidHash()
}

func (bc *Blockchain) GetBalance(address string) (uint64, error) {
	return bc.Database.GetBalance(address)
}

func (bc *Blockchain) updateUserBalances(txs []Transaction) error {
	sqlTx, err := bc.Database.DB.Begin()
	if err != nil {
		return err
	}

	defer sqlTx.Rollback()

	for _, tx := range txs {

		if tx.IsCoinbase {
			if err := bc.Database.IncreaseUserBalance(tx.To, tx.Amount); err != nil {
				return fmt.Errorf("failed to credit miner %s: %w", tx.To, err)
			}

			continue
		}

		if err := bc.Database.DecreaseUserBalance(tx.From, tx.Amount); err != nil {
			return fmt.Errorf("failed to debit sender %s: %w", tx.From, err)
		}

		if err := bc.Database.IncreaseUserBalance(tx.To, tx.Amount); err != nil {
			return fmt.Errorf("failed to credit receiver %s: %w", tx.From, err)
		}
	}

	return sqlTx.Commit()
}
