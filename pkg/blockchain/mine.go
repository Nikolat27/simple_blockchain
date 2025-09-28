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

			SortTxsByFee(&transactions)

			coinBaseTx := createCoinbaseTx(minerAddress, MiningReward)

			allTransactions := append([]Transaction{*coinBaseTx}, transactions...)

			bc.mu.RLock()
			prevHash := getPreviousBlockHash(bc.Blocks)
			blockIndex := len(bc.Blocks)
			bc.mu.RUnlock()

			newBlock := &Block{
				Id:           int64(blockIndex),
				PrevHash:     prevHash,
				Timestamp:    time.Now().Unix(),
				Transactions: allTransactions,
				Nonce:        0,
			}

			isBlockMined, err := proofOfWork(ctx, newBlock)
			if err != nil {
				log.Println(err)
			}

			if !isBlockMined {
				return nil // cancelled
			}

			// block found
			bc.AddBlock(newBlock)

			if err := bc.updateUserBalances(newBlock.Transactions); err != nil {
				log.Printf("failed to update the user balances: %v\n", err)
				return nil
			}

			mempool.Clear()
		}
	}
}

func proofOfWork(ctx context.Context, block *Block) (bool, error) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("POW operation cancelled")
			return false, nil
		default:
			if err := block.HashBlock(); err != nil {
				return false, err
			}

			if block.IsValidHash() {
				fmt.Println("Mined a block")
				return true, nil
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
