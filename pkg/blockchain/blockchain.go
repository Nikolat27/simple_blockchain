package blockchain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"simple_blockchain/pkg/database"
	"sync"
)

const MiningReward = 10000
const Difficulty = 5

type Blockchain struct {
	Blocks []Block `json:"blocks"`
	mu     sync.RWMutex

	Database *database.Database
	Mempool  *Mempool
}

func LoadBlockchain(db *database.Database, mp *Mempool) (*Blockchain, error) {
	bc := &Blockchain{
		Blocks:   make([]Block, 0),
		Database: db,
		Mempool:  mp,
	}

	blocks, err := bc.GetAllBlocks()
	if err != nil {
		return nil, err
	}

	// database is empty
	if len(blocks) == 0 {
		return bc, nil
	}

	allBlocksValid := true
	for _, block := range blocks {
		isVerified, err := bc.VerifyBlock(&block)
		if err != nil {
			return nil, err
		}

		if !isVerified {
			allBlocksValid = false
			break // If any block is invalid, don't trust the chain
		}

		bc.Blocks = append(bc.Blocks, block)
	}

	// If any block failed verification, clear database and start fresh
	if !allBlocksValid {
		log.Println("Found corrupted blockchain data, clearing database")
		if err := db.ClearAllData(); err != nil {
			return nil,
				fmt.Errorf("ERROR clearing corrupted data: %v", err)
		}

		return &Blockchain{
			Blocks:   make([]Block, 0),
			Database: db,
		}, nil
	}

	return bc, nil
}

func NewBlockchain(db *database.Database, mp *Mempool) (*Blockchain, error) {
	bc := &Blockchain{
		Blocks:   make([]Block, 0),
		Database: db,
		Mempool:  mp,
	}

	genesisBlock, err := createGenesisBlock()
	if err != nil {
		return nil, err
	}

	if err := bc.AddBlock(genesisBlock); err != nil {
		return nil, err
	}

	return bc, nil
}

func (bc *Blockchain) AddBlock(newBlock *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.Blocks = append(bc.Blocks, *newBlock)

	prevHashStr := hex.EncodeToString(newBlock.PrevHash)
	hashStr := hex.EncodeToString(newBlock.Hash)

	// Add block to database with block height
	blockId, err := bc.Database.AddBlock(prevHashStr, hashStr, newBlock.Nonce,
		newBlock.Timestamp, newBlock.Id)

	if err != nil {
		return fmt.Errorf("ERROR adding block: %v\n", err)
	}

	for _, tx := range newBlock.Transactions {
		signatureStr := hex.EncodeToString(tx.Signature)

		dbTx := database.DBTransactionSchema{
			From:       tx.From,
			To:         tx.To,
			Amount:     tx.Amount,
			Fee:        tx.Fee,
			Timestamp:  tx.Timestamp,
			PublicKey:  tx.PublicKey,
			Signature:  signatureStr,
			Status:     "confirmed",
			IsCoinbase: tx.IsCoinbase,
		}

		if err := bc.Database.AddTransaction(dbTx, int(blockId)); err != nil {
			return fmt.Errorf("ERROR adding transaction: %v\n", err)
		}
	}

	return nil
}

func (bc *Blockchain) GetAllBlocks() ([]Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	rows, err := bc.Database.GetAllBlocks()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		var block Block
		var prevHashStr, hashStr string
		var dbId int // database ID (index), not used for block identification

		if err := rows.Scan(&dbId, &prevHashStr,
			&hashStr, &block.Nonce, &block.Timestamp, &block.Id); err != nil {

			return nil, err
		}

		block.PrevHash, err = hex.DecodeString(prevHashStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode 'prevHashStr': %v", err)
		}

		block.Hash, err = hex.DecodeString(hashStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode 'hashStr': %v", err)
		}

		// Load transactions for this block using database ID
		dbTransactions, err := bc.Database.GetTransactionsByBlockId(dbId)
		if err != nil {
			return nil, err
		}

		block.parseDBTransactions(dbTransactions)

		blocks = append(blocks, block)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (bc *Blockchain) VerifyBlock(block *Block) (bool, error) {
	tempBlock := &Block{
		Id:           block.Id,
		PrevHash:     block.PrevHash,
		Timestamp:    block.Timestamp,
		Transactions: block.Transactions,
		Nonce:        block.Nonce,
		Hash:         nil,
	}

	if err := tempBlock.HashBlock(); err != nil {
		return false, err
	}

	hashMatches := bytes.Equal(tempBlock.Hash, block.Hash)

	if hashMatches && isGenesisBlock(tempBlock.Id) {
		return true, nil
	}

	return hashMatches && tempBlock.IsValidHash(), nil
}

func (bc *Blockchain) GetBalance(address string) (uint64, error) {
	mempoolTxs := bc.Mempool.GetTransactions()

	confirmedBalance, err := bc.Database.GetConfirmedBalance(address)
	if err != nil {
		return 0, err
	}

	log.Println("Confirmed Balance: ", confirmedBalance)

	pendingOutgoing := getUserPendingOutgoing(address, mempoolTxs)

	log.Println("Pending outgoing: ", pendingOutgoing)

	if confirmedBalance < pendingOutgoing {
		return 0, nil
	}

	effectiveBalance := confirmedBalance - pendingOutgoing

	log.Println("Effective balance: ", effectiveBalance)

	return effectiveBalance, nil
}

func (bc *Blockchain) updateUserBalances(txs []Transaction) error {
	sqlTx, err := bc.Database.Begin()
	if err != nil {
		return err
	}

	defer sqlTx.Rollback()

	// Calculate total fees from all transactions in this block
	var totalFees uint64

	for _, tx := range txs {
		if !tx.IsCoinbase {
			totalFees += tx.Fee
		}
	}

	for _, tx := range txs {
		if tx.IsCoinbase {
			// Credit miner with mining reward + total fees from this block
			minerReward := tx.Amount + totalFees
			if err := bc.Database.IncreaseUserBalance(sqlTx, tx.To, minerReward); err != nil {
				return fmt.Errorf("failed to credit miner %s: %w", tx.To, err)
			}
			continue
		}

		// For regular transactions: sender pays amount + fee
		totalDebit := tx.Amount + tx.Fee
		if err := bc.Database.DecreaseUserBalance(sqlTx, tx.From, totalDebit); err != nil {
			return fmt.Errorf("failed to debit sender %s: %w", tx.From, err)
		}

		// Credit receiver with the transaction amount
		if tx.To != "" {
			if err := bc.Database.IncreaseUserBalance(sqlTx, tx.To, tx.Amount); err != nil {
				return fmt.Errorf("failed to credit receiver %s: %w", tx.To, err)
			}
		}
	}

	return sqlTx.Commit()
}

func (bc *Blockchain) ValidateTransaction(tx *Transaction) error {
	if tx.IsCoinbase {
		return nil
	}

	balance, err := bc.GetBalance(tx.From)
	if err != nil {
		return err
	}

	totalCost := tx.Amount + tx.Fee
	if balance >= totalCost {
		return nil
	}

	return errors.New("balance is insufficient")
}

func getUserPendingOutgoing(address string, mempoolTxs []Transaction) uint64 {
	var pending uint64

	for _, tx := range mempoolTxs {
		if tx.IsCoinbase {
			continue
		}

		if tx.From == address {
			pending += tx.Amount + tx.Fee // Include both amount and fee
		}
	}

	return pending
}

func isGenesisBlock(id int64) bool {
	return id == 0
}
