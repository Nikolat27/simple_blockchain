package blockchain

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (bc *Blockchain) MineBlock(ctx context.Context, mempool *Mempool, minerAddress string) *Block {
	// Continuously mines new blocks until the operation is cancelled
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Mining cancelled")
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

			// keep looping until POW is true
			if !proofOfWork(ctx, newBlock) {
				return nil // cancelled
			}

			// block found
			bc.AddBlock(newBlock)

			if err := bc.LevelDB.IncreaseUserBalance([]byte(minerAddress), MiningReward); err != nil {
				log.Fatalf("ERROR increasing user balance: %v\n", err)
			}

			mempool.Clear()
		}
	}
}

func proofOfWork(ctx context.Context, block *Block) bool {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("POW operating cancelled")
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

func getPreviousBlockHash(blocks []Block) []byte {
	var prevHash []byte
	if len(blocks) > 0 {
		prevHash = blocks[len(blocks)-1].Hash
	} else {
		prevHash = make([]byte, 32)
	}

	return prevHash
}
