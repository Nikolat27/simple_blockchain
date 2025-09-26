package blockchain

import (
	"bytes"
	"simple_blockchain/pkg/LevelDB"
	"sync"
)

const MiningReward = 5
const Difficulty = 4

type Blockchain struct {
	Blocks []Block `json:"blocks"`
	mu     sync.RWMutex

	LevelDB *LevelDB.LevelDB
}

func NewBlockchain(genesisAddress string, levelDB *LevelDB.LevelDB) *Blockchain {
	bc := &Blockchain{
		Blocks:  make([]Block, 0),
		LevelDB: levelDB,
	}

	genesisTx := &Transaction{
		To:         genesisAddress,
		Amount:     1000,
		Status:     "confirmed",
		IsCoinbase: true,
	}

	genesisBlock := createGenesisBlock(genesisTx)

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

func (bc *Blockchain) GetBalance(address string) uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	balance := uint64(0)

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.To == address {
				balance += tx.Amount
			}

			if tx.From == address && !tx.IsCoinbase {
				balance -= tx.Amount
			}
		}
	}

	return balance
}
