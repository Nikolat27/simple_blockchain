package blockchain

import (
	"bytes"
	"fmt"
	"simple_blockchain/pkg/LevelDB"
	"strconv"
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

	// Apply genesis transaction balance
	if err := bc.LevelDB.IncreaseUserBalance([]byte(genesisAddress), int(genesisTx.Amount)); err != nil {
		panic(fmt.Sprintf("failed to apply genesis balance: %v", err))
	}

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
	value, err := bc.LevelDB.Get([]byte(address), []byte("0"))
	if err != nil {
		return 0, err
	}

	// convert []byte to uint64
	return strconv.ParseUint(string(value), 10, 64)
}
