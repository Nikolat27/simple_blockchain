package blockchain

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
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
	Mutex  sync.RWMutex

	Database *database.Database
	Mempool  *Mempool `json:"mempool"`
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

	sqlTx, err := db.BeginTx()
	if err != nil {
		return nil, err
	}
	defer sqlTx.Rollback()

	if err := bc.AddBlock(sqlTx, genesisBlock); err != nil {
		return nil, err
	}

	if err := sqlTx.Commit(); err != nil {
		return nil, err
	}

	bc.AddBlockToMemory(genesisBlock)

	return bc, nil
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

	allBlocksValid, err := bc.VerifyBlocks(blocks)
	if err != nil {
		return nil, err
	}

	// If any block failed verification, clear database and start fresh
	if !allBlocksValid {
		return startFresh(db)
	}

	return bc, nil
}

func (bc *Blockchain) VerifyBlocks(blocks []Block) (bool, error) {
	// database is empty
	if len(blocks) == 0 {
		return true, nil
	}

	for _, block := range blocks {
		isVerified, err := bc.VerifyBlock(&block)
		if err != nil {
			return false, err
		}

		if !isVerified {
			return false, nil
		}

		bc.AddBlockToMemory(&block)
	}

	return true, nil
}

func startFresh(db *database.Database) (*Blockchain, error) {
	log.Println("Found corrupted blockchain data, clearing database")

	sqlTx, err := db.BeginTx()
	if err != nil {
		return nil, err
	}
	defer sqlTx.Rollback()

	if err := db.ClearAllData(sqlTx); err != nil {
		return nil, fmt.Errorf("ERROR clearing corrupted data: %w", err)
	}

	if err := sqlTx.Commit(); err != nil {
		return nil, err
	}

	return &Blockchain{
		Blocks:   make([]Block, 0),
		Database: db,
	}, nil
}

func (bc *Blockchain) AddBlock(sqlTx *sql.Tx, newBlock *Block) error {
	prevHashStr := hex.EncodeToString(newBlock.PrevHash)
	hashStr := hex.EncodeToString(newBlock.Hash)
	merkleRootStr := hex.EncodeToString(newBlock.MerkleRoot)

	blockId, err := bc.Database.AddBlock(sqlTx, prevHashStr, hashStr, merkleRootStr,
		newBlock.Nonce, newBlock.Timestamp, newBlock.Id)

	if err != nil {
		return err
	}

	if err := bc.AddTransactionToDB(sqlTx, int(blockId), newBlock.Transactions); err != nil {
		return err
	}

	if err := bc.UpdateUserBalances(sqlTx, newBlock.Transactions); err != nil {
		return err
	}

	return nil
}

func (bc *Blockchain) AddTransactionToDB(dbTx *sql.Tx, blockId int, txs []Transaction) error {
	for _, tx := range txs {
		signatureStr := hex.EncodeToString(tx.Signature)

		txInstance := database.DBTransactionSchema{
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

		if err := bc.Database.AddTransaction(dbTx, txInstance, blockId); err != nil {
			return fmt.Errorf("ERROR adding transaction: %w", err)
		}
	}

	return nil
}

func (bc *Blockchain) GetAllBlocks() ([]Block, error) {
	bc.Mutex.RLock()
	defer bc.Mutex.RUnlock()

	rows, err := bc.Database.GetAllBlocks()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blocksCount, err := bc.Database.GetBlocksCount()
	if err != nil {
		return nil, err
	}

	blocks := make([]Block, 0, blocksCount)
	for rows.Next() {
		block, err := bc.parseBlock(rows)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, *block)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (bc *Blockchain) parseBlock(rows *sql.Rows) (*Block, error) {
	var block Block

	var prevHashStr, hashStr, merkleRootStr string
	var dbId int // database ID (index), not used for block identification

	if err := rows.Scan(&dbId, &prevHashStr, &hashStr, &merkleRootStr, &block.Nonce,
		&block.Timestamp, &block.Id); err != nil {

		return nil, err
	}

	var err error

	block.PrevHash, err = hex.DecodeString(prevHashStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode 'prevHashStr': %v", err)
	}

	block.Hash, err = hex.DecodeString(hashStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode 'hashStr': %v", err)
	}

	block.MerkleRoot, err = hex.DecodeString(merkleRootStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode 'merkleRootStr': %v", err)
	}

	// Load transactions for this block using database ID
	dbTransactions, err := bc.Database.GetTransactionsByBlockId(dbId)
	if err != nil {
		return nil, err
	}

	if err := block.parseDBTransactions(dbTransactions); err != nil {
		return nil, err
	}

	return &block, nil
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

	// No more validation for genesis block
	if block.Id == 0 {
		return hashMatches, nil
	}

	prevBlock, err := bc.GetBlockById(block.Id - 1)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(prevBlock.Hash, tempBlock.PrevHash) {
		return false, err
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

func (bc *Blockchain) UpdateUserBalances(sqlTx *sql.Tx, txs []Transaction) error {
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

	return nil
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

func (bc *Blockchain) VerifyHeaders(headers []BlockHeader) (bool, error) {
	if len(headers) == 0 {
		return true, nil
	}

	genesis := headers[0]
	if genesis.Id != 0 {
		return false, fmt.Errorf("first block must have ID 0")
	}

	var emptyHash [sha256.Size]byte
	if !bytes.Equal(genesis.PrevHash, emptyHash[:]) {
		return false, fmt.Errorf("genesis block must have empty prev Hash")
	}

	for idx, header := range headers {
		var prevHash []byte
		if idx > 0 {
			prevHash = headers[idx-1].Hash
		} else {
			// Genesis block has no previous Hash
			prevHash = make([]byte, sha256.Size)
		}

		verified, err := header.Verify(int64(idx), prevHash)
		if err != nil {
			return false, err
		}

		if !verified {
			return false, nil
		}
	}

	return true, nil
}

func (bc *Blockchain) GetBlockById(blockId int64) (*Block, error) {
	row, err := bc.Database.GetBlockById(blockId)
	if err != nil {
		return nil, err
	}

	var block Block
	var prevHashStr, hashStr, merkleRootStr string
	var dbId int // database ID (index)

	if err := row.Scan(&dbId, &prevHashStr, &hashStr, &merkleRootStr,
		&block.Nonce, &block.Timestamp, &block.Id); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, err
	}

	block.PrevHash, err = hex.DecodeString(prevHashStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode prevHash: %v", err)
	}

	block.Hash, err = hex.DecodeString(hashStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Hash: %v", err)
	}

	block.MerkleRoot, err = hex.DecodeString(merkleRootStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode merkleRoot: %v", err)
	}

	// Load transactions for this block
	dbTransactions, err := bc.Database.GetTransactionsByBlockId(dbId)
	if err != nil {
		return nil, err
	}

	if err := block.parseDBTransactions(dbTransactions); err != nil {
		return nil, err
	}

	return &block, nil
}

func (bc *Blockchain) AddBlockToMemory(block *Block) bool {
	bc.Mutex.Lock()
	defer bc.Mutex.Unlock()

	// Check for duplicate block ID
	for _, existing := range bc.Blocks {
		if existing.Id == block.Id {
			return false // Block already exists
		}
	}

	bc.Blocks = append(bc.Blocks, *block)
	return true
}

func (bc *Blockchain) SyncMempool(mempool *Mempool) {
	bc.Mempool.Mutex.Lock()
	defer bc.Mempool.Mutex.Unlock()

	for idx, tx := range mempool.Transactions {
		for _, existingTx := range bc.Mempool.Transactions {
			if existingTx.
		}

		if _, exists := bc.Mempool.Transactions[idx]; !exists {
			bc.Mempool.Transactions[idx] = tx
		}
	}
}
