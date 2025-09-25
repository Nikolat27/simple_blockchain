package blockchain

import (
	"bytes"
	"encoding/hex"
	"sync"
	"time"
)

const MiningReward = 50

type Blockchain struct {
	Blocks []Block      `json:"blocks"`
	mu     sync.RWMutex // Protects concurrent access to blocks
}

func NewBlockchain(genesisAddress string) *Blockchain {
	bc := &Blockchain{
		Blocks: make([]Block, 0),
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

func (bc *Blockchain) MineBlock(mempool *Mempool, minerAddress string) *Block {
	for {
		transactions := mempool.GetTransactions()

		coinBaseTx := Transaction{
			To:         minerAddress,
			Amount:     MiningReward,
			Status:     "confirmed",
			IsCoinbase: true,
		}

		allTransactions := append([]Transaction{coinBaseTx}, transactions...)

		bc.mu.RLock()
		prevHash := getPreviousBlockHash(bc.Blocks)
		blockIndex := len(bc.Blocks)
		bc.mu.RUnlock()

		newBlock := &Block{
			Index:        blockIndex,
			PrevHash:     prevHash,
			Timestamp:    time.Now(),
			Transactions: allTransactions,
			Nonce:        0,
		}

		proofOfWork(newBlock)

		bc.mu.RLock()
		currentTip := getPreviousBlockHash(bc.Blocks)
		isBlockMined := !bytes.Equal(currentTip, newBlock.PrevHash)
		bc.mu.RUnlock()

		if isBlockMined {
			// Somebody else mined the block first → try mining again
			continue
		}

		bc.AddBlock(newBlock)
		mempool.Clear()

		return newBlock
	}
}

func getPreviousBlockHash(blocks []Block) []byte {
	var prevHash []byte
	if len(blocks) > 0 {
		prevHash = blocks[len(blocks)-1].Hash
	} else {
		prevHash = make([]byte, 32)
	}

	return prevHash
}

func proofOfWork(newBlock *Block) {
	for {
		newBlock.HashBlock()
		if newBlock.IsValidHash() {
			break
		}

		newBlock.Nonce++
	}
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

	hashMatches := hex.EncodeToString(tempBlock.Hash) == hex.EncodeToString(block.Hash)

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

func (bc *Blockchain) ValidateTransaction(tx *Transaction) bool {
	if tx.IsCoinbase {
		return true
	}

	balance := bc.GetBalance(tx.From)
	return balance >= tx.Amount
}

func createGenesisBlock(genesisTx *Transaction) *Block {
	block := &Block{
		Index:        0,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now(),
		Transactions: []Transaction{*genesisTx},
		Nonce:        0,
	}

	block.HashBlock()
	return block
}
