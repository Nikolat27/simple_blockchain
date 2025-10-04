package blockchain

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (bc *Blockchain) MineBlock(ctx context.Context, mempool *Mempool, minerAddress string) (*Block, error) {
	// Continuously mines new blocks until the operation is cancelled
	select {
	case <-ctx.Done():
		fmt.Println("Mining cancelled")
		return nil, nil
	default:
		transactions := mempool.GetTransactions()

		// Priority based
		sortedTxs := bc.Mempool.SortTxsByFee(transactions)

		coinBaseTx := createCoinbaseTx(minerAddress, MiningReward)

		allTransactions := append([]Transaction{*coinBaseTx}, sortedTxs...)

		bc.Mutex.RLock()
		prevHash := getPreviousBlockHash(bc.Blocks)
		blockIndex := len(bc.Blocks)
		bc.Mutex.RUnlock()

		newBlock := &Block{
			Id:           int64(blockIndex),
			PrevHash:     prevHash,
			Hash:         nil,
			Timestamp:    time.Now().Unix(),
			Transactions: allTransactions,
			Nonce:        0,
		}

		// mining started...
		mined, err := proofOfWork(ctx, newBlock)
		if err != nil {
			return nil, fmt.Errorf("ERROR proofOfWork: %v", err)
		}

		if !mined {
			fmt.Println("POW operation cancelled")
			return nil, nil
		}

		sqlTx, err := bc.Database.BeginTx()
		if err != nil {
			return nil, err
		}
		defer sqlTx.Rollback()

		// block found
		if err := bc.AddBlock(sqlTx, newBlock); err != nil {
			return nil, err
		}

		if err := sqlTx.Commit(); err != nil {
			return nil, err
		}

		bc.AddBlockToMemory(newBlock)

		mempool.Clear()

		log.Println("Mined a block")

		return newBlock, nil
	}
}

func proofOfWork(ctx context.Context, block *Block) (bool, error) {
	for {
		select {
		case <-ctx.Done():
			// cancelled
			return false, nil
		default:
			if err := block.HashBlock(); err != nil {
				return false, err
			}

			if block.IsValidHash() {
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
