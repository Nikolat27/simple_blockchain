package blockchain

import (
	"log"
	"time"
)

func (bc *Blockchain) StartMining(mempool *Mempool, minerAddress string) {
	go bc.MineBlock(mempool, minerAddress)
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

		if proofOfWork(newBlock) {
			bc.AddBlock(newBlock)

			if err := bc.LevelDB.IncreaseUserBalance([]byte(minerAddress), MiningReward); err != nil {
				log.Fatalf("ERROR increasing user balance: %v\n", err)
			}

			mempool.Clear()

			// restart mining again on the new block
			bc.StartMining(mempool, minerAddress)

			return newBlock
		}
	}
}

func proofOfWork(block *Block) bool {
	for {
		block.HashBlock()
		if block.IsValidHash() {
			return true
		}
		block.Nonce++
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
