package blockchain

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/Nikolat27/simple_blockchain/pkg/utils"
)

func (bc *Blockchain) MineBlock(ctx context.Context, mempool *Mempool, minerAddress string) (*Block, error) {
	select {
	case <-ctx.Done():
		fmt.Println("Mining cancelled")
		return nil, nil
	default:
		transactions := mempool.GetTransactionsCopy()

		// Priority based
		sortedTxs := sortTxsByFee(transactions)

		coinBaseTx := CreateCoinbaseTx(minerAddress, MiningReward)

		allTransactions := append([]Transaction{*coinBaseTx}, sortedTxs...)

		bc.Mutex.RLock()
		prevHash := getPreviousBlockHash(bc.Blocks)
		blockIndex := len(bc.Blocks)
		bc.Mutex.RUnlock()

		newBlock := &Block{
			Id:           int64(blockIndex),
			PrevHash:     prevHash,
			Hash:         nil,
			Timestamp:    utils.GetTimestamp(),
			Transactions: allTransactions,
			Nonce:        0,
		}

		// mining started...
		mined, err := bc.proofOfWork(ctx, newBlock)
		if err != nil {
			return nil, fmt.Errorf("ERROR proofOfWork: %v", err)
		}

		if !mined {
			fmt.Println("POW operation cancelled")
			return nil, nil
		}

		bc.Mutex.RLock()
		latestHash := getPreviousBlockHash(bc.Blocks)
		bc.Mutex.RUnlock()

		if !bytes.Equal(latestHash, newBlock.PrevHash) {
			fmt.Println("Block was already mined by someone else")
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

		mempool.DeleteMinedTransactions(newBlock.Transactions)

		log.Println("Mined a block")

		return newBlock, nil
	}
}

func (bc *Blockchain) proofOfWork(ctx context.Context, block *Block) (bool, error) {
	for {
		select {
		case <-ctx.Done():
			// cancelled
			return false, nil
		case <-bc.CancelMiningCh:
			// mined or cancelled
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

// sortTxsByFee sorts transactions in descending order by their fee
func sortTxsByFee(txs map[string]Transaction) []Transaction {
	sortedTxs := make([]Transaction, 0, len(txs))
	for _, tx := range txs {
		sortedTxs = append(sortedTxs, tx)
	}

	sort.Slice(sortedTxs, func(i, j int) bool {
		return sortedTxs[i].Fee > sortedTxs[j].Fee
	})

	return sortedTxs
}
