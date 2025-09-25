package blockchain

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const MiningReward = 50

type Blockchain struct {
	Blocks       []Block           `json:"blocks"`
	orphanBlocks map[string]*Block // blocks that don't connect to current tip (keyed by prevHash)
	mu           sync.RWMutex

	mineCancel   context.CancelFunc
	onBlockMined func(*Block) // callback for when a block is mined locally
}

func NewBlockchain(genesisAddress string) *Blockchain {
	bc := &Blockchain{
		Blocks:       make([]Block, 0),
		orphanBlocks: make(map[string]*Block),
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

// SetOnBlockMined sets the callback function that will be called when a block is mined locally
func (bc *Blockchain) SetOnBlockMined(callback func(*Block)) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.onBlockMined = callback
}

// StopMining cancels any ongoing mining operation
func (bc *Blockchain) StopMining() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if bc.mineCancel != nil {
		bc.mineCancel()
		bc.mineCancel = nil
	}
}

func (bc *Blockchain) AddBlock(newBlock *Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.Blocks = append(bc.Blocks, *newBlock)
}

// AcceptBlock implements the longest chain rule for block acceptance
func (bc *Blockchain) AcceptBlock(block *Block) bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Check if block connects to current tip
	if len(bc.Blocks) > 0 {
		tipHash := bc.Blocks[len(bc.Blocks)-1].Hash
		if bytes.Equal(block.PrevHash, tipHash) {
			// Block connects to current tip - add it directly
			bc.Blocks = append(bc.Blocks, *block)

			// Process any orphan blocks that might now connect
			bc.processConnectedOrphans()
			return true
		}
	} else {
		// Empty blockchain - this must be the genesis block
		bc.Blocks = append(bc.Blocks, *block)
		return true
	}

	// Block doesn't connect to current tip - check if it connects to any orphan
	// or if we should treat it as an orphan
	bc.AddOrphanBlock(block)

	// Try to process orphans in case this block enables a chain extension
	bc.processConnectedOrphans()

	return true // Block accepted as orphan or main chain
}

// processConnectedOrphans processes orphan blocks that can now be connected
func (bc *Blockchain) processConnectedOrphans() {
	connectedBlocks := make([]*Block, 0)

	for {
		foundNew := false

		if len(bc.Blocks) == 0 {
			break
		}

		tipHash := bc.Blocks[len(bc.Blocks)-1].Hash
		tipHashStr := hex.EncodeToString(tipHash)

		// Find orphans that connect to current tip
		for prevHashStr, orphanBlock := range bc.orphanBlocks {
			if prevHashStr == tipHashStr {
				connectedBlocks = append(connectedBlocks, orphanBlock)
				delete(bc.orphanBlocks, prevHashStr)
				foundNew = true
				fmt.Printf("🔗 Connected orphan block %d to main chain\n", orphanBlock.Index)
			}
		}

		// Add connected blocks to main chain
		for _, block := range connectedBlocks {
			bc.Blocks = append(bc.Blocks, *block)
			fmt.Printf("✅ Added connected block %d to main chain\n", block.Index)
		}

		connectedBlocks = connectedBlocks[:0] // Clear slice

		if !foundNew {
			break
		}
	}
}

func (bc *Blockchain) GetBlocks() []Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	blocksCopy := make([]Block, len(bc.Blocks))
	copy(blocksCopy, bc.Blocks)

	return blocksCopy
}

func (bc *Blockchain) StartMining(mempool *Mempool, minerAddress string) {
	ctx, cancel := context.WithCancel(context.Background())

	bc.mu.Lock()
	bc.mineCancel = cancel
	bc.mu.Unlock()

	go bc.MineBlock(ctx, mempool, minerAddress)
}

func (bc *Blockchain) MineBlock(ctx context.Context, mempool *Mempool, minerAddress string) *Block {
	for {
		select {
		case <-ctx.Done():
			// stop mining
			return nil
		default:
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

			if proofOfWork(ctx, newBlock) {
				bc.mu.Lock()

				if bc.mineCancel != nil {
					bc.mineCancel()
					bc.mineCancel = nil
				}
				bc.AddBlock(newBlock)

				// Call the callback if set (for broadcasting the block)
				callback := bc.onBlockMined
				bc.mu.Unlock()

				mempool.Clear()

				// Broadcast the block if callback is set
				if callback != nil {
					callback(newBlock)
				}

				// restart mining again on the new block
				bc.StartMining(mempool, minerAddress)

				return newBlock
			}
		}
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

func proofOfWork(ctx context.Context, block *Block) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
			block.HashBlock()
			if block.IsValidHash() {
				return true
			}
			block.Nonce++
		}
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

// AddOrphanBlock adds a block to the orphan pool
// Note: This function assumes the caller already holds bc.mu
func (bc *Blockchain) AddOrphanBlock(block *Block) {
	bc.orphanBlocks[hex.EncodeToString(block.PrevHash)] = block
	fmt.Printf("🧒 Added block %d to orphan pool (waiting for parent %x)\n",
		block.Index, block.PrevHash[:8])
}

// ProcessOrphanBlocks checks if any orphan blocks can now be connected
func (bc *Blockchain) ProcessOrphanBlocks() []*Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	connectedBlocks := make([]*Block, 0)

	if len(bc.Blocks) == 0 {
		return connectedBlocks
	}

	tipHash := bc.Blocks[len(bc.Blocks)-1].Hash
	tipHashStr := hex.EncodeToString(tipHash)

	// Check if any orphan blocks connect to the current tip
	for prevHashStr, orphanBlock := range bc.orphanBlocks {
		if prevHashStr == tipHashStr {
			// This orphan block can now be connected
			connectedBlocks = append(connectedBlocks, orphanBlock)
			delete(bc.orphanBlocks, prevHashStr)
		}
	}

	return connectedBlocks
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
