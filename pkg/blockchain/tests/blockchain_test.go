package tests

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nikolat27/simple_blockchain/pkg/CryptoGraphy"
	"github.com/Nikolat27/simple_blockchain/pkg/blockchain"
	"github.com/Nikolat27/simple_blockchain/pkg/database"
	"github.com/Nikolat27/simple_blockchain/pkg/utils"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) (*database.Database, string, func()) {
	// Create a temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_blockchain.db")

	db, err := database.New("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	err = runMigrations(db.DB)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, dbPath, cleanup
}

// runMigrations creates the necessary tables for testing
func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS balances (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			address TEXT NOT NULL UNIQUE,
			balance INTEGER NOT NULL DEFAULT(0)
		)`,
		`CREATE TABLE IF NOT EXISTS blocks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			prev_hash TEXT UNIQUE NOT NULL,
			hash TEXT UNIQUE NOT NULL,
			merkle_root TEXT UNIQUE NOT NULL,
			nonce INTEGER DEFAULT (0),
			timestamp INTEGER DEFAULT (strftime('%s', 'now')),
			block_height INTEGER DEFAULT (0)
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			block_id INTEGER NULL,
			sender TEXT NULL,
			recipient TEXT NOT NULL,
			amount INTEGER DEFAULT (0),
			fee INTEGER DEFAULT (0),
			timestamp INTEGER DEFAULT (strftime('%s', 'now')),
			public_key TEXT NULL,
			signature TEXT NULL,
			status TEXT NOT NULL DEFAULT ('pending') CHECK (status IN ('pending', 'confirmed')),
			is_coin_base INTEGER NOT NULL DEFAULT (0) CHECK (is_coin_base IN (0, 1)),
			FOREIGN KEY (block_id) REFERENCES blocks (id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_block_id ON transactions (block_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}

// mineBlock is a helper function to mine a block for testing
// Note: With difficulty 6, this can take several seconds
func mineBlock(block *blockchain.Block) error {
	block.ComputeMerkleRoot()

	// Simple mining loop - find a nonce that produces a valid hash
	// Increased limit for difficulty 6 (requires 6 leading zeros)
	for nonce := int64(0); nonce < 100000000; nonce++ {
		block.Nonce = nonce
		if err := block.HashBlock(); err != nil {
			return err
		}

		if block.IsValidHash() {
			return nil
		}

		// Log progress every million attempts
		if nonce%1000000 == 0 && nonce > 0 {
			// Optional: can add logging here if needed
		}
	}

	return fmt.Errorf("failed to mine block within nonce limit")
}

// TestNewBlockchain tests blockchain initialization with genesis block
func TestNewBlockchain(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)

	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	if bc == nil {
		t.Fatal("Blockchain should not be nil")
	}

	if len(bc.Blocks) != 1 {
		t.Errorf("Expected 1 block (genesis), got %d", len(bc.Blocks))
	}

	// Verify genesis block properties
	genesisBlock := bc.Blocks[0]
	if genesisBlock.Id != 0 {
		t.Errorf("Genesis block ID should be 0, got %d", genesisBlock.Id)
	}

	if len(genesisBlock.Transactions) != 0 {
		t.Errorf("Genesis block should have 0 transactions, got %d", len(genesisBlock.Transactions))
	}
}

// TestNewBlockchain_NilMempool tests blockchain creation with nil mempool
func TestNewBlockchain_NilMempool(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	bc, err := blockchain.NewBlockchain(db, nil)

	if err == nil {
		t.Error("Expected error when creating blockchain with nil mempool")
	}

	if bc != nil {
		t.Error("Blockchain should be nil when mempool is nil")
	}
}

// TestLoadBlockchain tests loading blockchain from database
func TestLoadBlockchain(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)

	// Create initial blockchain
	bc1, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Load blockchain from database
	bc2, err := blockchain.LoadBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to load blockchain: %v", err)
	}

	if len(bc2.Blocks) != len(bc1.Blocks) {
		t.Errorf("Loaded blockchain should have %d blocks, got %d", len(bc1.Blocks), len(bc2.Blocks))
	}
}

// TestLoadBlockchain_EmptyDatabase tests loading from empty database
func TestLoadBlockchain_EmptyDatabase(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)

	// Load from empty database
	bc, err := blockchain.LoadBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to load blockchain from empty database: %v", err)
	}

	if len(bc.Blocks) != 0 {
		t.Errorf("Expected 0 blocks from empty database, got %d", len(bc.Blocks))
	}
}

// TestAddBlock tests adding a block to blockchain
func TestAddBlock(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Create a new block with a unique transaction to avoid merkle root collision
	prevBlock := bc.GetLatestBlock()
	coinbaseTx := blockchain.CreateCoinbaseTx("miner1", blockchain.MiningReward)

	newBlock := &blockchain.Block{
		Id:           prevBlock.Id + 1,
		PrevHash:     prevBlock.Hash,
		Timestamp:    utils.GetTimestamp(),
		Transactions: []blockchain.Transaction{*coinbaseTx},
		Nonce:        0,
	}

	newBlock.ComputeMerkleRoot()
	err = newBlock.HashBlock()
	if err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	// Add block to database
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = bc.AddBlock(sqlTx, newBlock)
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify block was added to database
	blocks, err := bc.GetAllBlocks()
	if err != nil {
		t.Fatalf("Failed to get all blocks: %v", err)
	}

	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(blocks))
	}
}

// TestAddBlockToMemory tests adding blocks to in-memory storage
func TestAddBlockToMemory(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	initialCount := len(bc.Blocks)

	// Create a new block
	newBlock := &blockchain.Block{
		Id:           1,
		PrevHash:     bc.GetLatestBlock().Hash,
		Timestamp:    utils.GetTimestamp(),
		Transactions: []blockchain.Transaction{},
		Nonce:        0,
	}

	newBlock.ComputeMerkleRoot()
	err = newBlock.HashBlock()
	if err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	// Add block to memory
	added := bc.AddBlockToMemory(newBlock)
	if !added {
		t.Error("Block should be added to memory")
	}

	if len(bc.Blocks) != initialCount+1 {
		t.Errorf("Expected %d blocks in memory, got %d", initialCount+1, len(bc.Blocks))
	}

	// Try to add duplicate block
	added = bc.AddBlockToMemory(newBlock)
	if added {
		t.Error("Duplicate block should not be added")
	}

	if len(bc.Blocks) != initialCount+1 {
		t.Errorf("Block count should remain %d after duplicate add, got %d", initialCount+1, len(bc.Blocks))
	}
}

// TestGetLatestBlock tests retrieving the latest block
func TestGetLatestBlock(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	latestBlock := bc.GetLatestBlock()
	if latestBlock == nil {
		t.Fatal("Latest block should not be nil")
	}

	if latestBlock.Id != 0 {
		t.Errorf("Latest block ID should be 0 (genesis), got %d", latestBlock.Id)
	}

	// Add another block with unique transaction
	coinbaseTx := blockchain.CreateCoinbaseTx("miner2", blockchain.MiningReward)
	newBlock := &blockchain.Block{
		Id:           1,
		PrevHash:     latestBlock.Hash,
		Timestamp:    utils.GetTimestamp(),
		Transactions: []blockchain.Transaction{*coinbaseTx},
		Nonce:        0,
	}

	newBlock.ComputeMerkleRoot()
	err = newBlock.HashBlock()
	if err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	bc.AddBlockToMemory(newBlock)

	latestBlock = bc.GetLatestBlock()
	if latestBlock.Id != 1 {
		t.Errorf("Latest block ID should be 1, got %d", latestBlock.Id)
	}
}

// TestVerifyBlock tests block verification
func TestVerifyBlock(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Verify genesis block
	genesisBlock := bc.GetLatestBlock()
	isValid, err := bc.VerifyBlock(genesisBlock)
	if err != nil {
		t.Fatalf("Error verifying genesis block: %v", err)
	}

	if !isValid {
		t.Error("Genesis block should be valid")
	}

	// Create and verify a new block with unique transaction
	coinbaseTx := blockchain.CreateCoinbaseTx("miner3", blockchain.MiningReward)
	newBlock := &blockchain.Block{
		Id:           1,
		PrevHash:     genesisBlock.Hash,
		Timestamp:    utils.GetTimestamp(),
		Transactions: []blockchain.Transaction{*coinbaseTx},
		Nonce:        0,
	}

	// Mine the block to find valid nonce
	err = mineBlock(newBlock)
	if err != nil {
		t.Fatalf("Failed to mine block: %v", err)
	}

	// Add to database first
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = bc.AddBlock(sqlTx, newBlock)
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	bc.AddBlockToMemory(newBlock)

	// Verify the new block
	isValid, err = bc.VerifyBlock(newBlock)
	if err != nil {
		t.Fatalf("Error verifying new block: %v", err)
	}

	if !isValid {
		t.Error("New block should be valid")
	}
}

// TestVerifyBlock_InvalidHash tests verification with invalid hash
func TestVerifyBlock_InvalidHash(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	genesisBlock := bc.GetLatestBlock()

	// Create block with invalid hash and unique transaction
	coinbaseTx := blockchain.CreateCoinbaseTx("miner4", blockchain.MiningReward)
	invalidBlock := &blockchain.Block{
		Id:           1,
		PrevHash:     genesisBlock.Hash,
		Timestamp:    utils.GetTimestamp(),
		Transactions: []blockchain.Transaction{*coinbaseTx},
		Nonce:        0,
		Hash:         []byte("invalid_hash"),
	}

	invalidBlock.ComputeMerkleRoot()

	// Add to database
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = bc.AddBlock(sqlTx, invalidBlock)
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	bc.AddBlockToMemory(invalidBlock)

	// Verify should fail
	isValid, err := bc.VerifyBlock(invalidBlock)
	if err != nil {
		t.Fatalf("Error verifying invalid block: %v", err)
	}

	if isValid {
		t.Error("Block with invalid hash should not be valid")
	}
}

// TestGetBalance tests balance retrieval
func TestGetBalance(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Test balance for non-existent address
	balance, err := bc.GetBalance("nonexistent")
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	if balance != 0 {
		t.Errorf("Expected balance 0 for non-existent address, got %d", balance)
	}

	// Add balance to an address
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = db.IncreaseUserBalance(sqlTx, "alice", 1000)
	if err != nil {
		t.Fatalf("Failed to increase balance: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Get balance
	balance, err = bc.GetBalance("alice")
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	if balance != 1000 {
		t.Errorf("Expected balance 1000, got %d", balance)
	}
}

// TestGetBalance_WithPendingTransactions tests balance with pending mempool transactions
func TestGetBalance_WithPendingTransactions(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Add balance to alice
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = db.IncreaseUserBalance(sqlTx, "alice", 1000)
	if err != nil {
		t.Fatalf("Failed to increase balance: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Create pending transaction in mempool
	tx := &blockchain.Transaction{
		From:      "alice",
		To:        "bob",
		Amount:    200,
		Fee:       10,
		Timestamp: utils.GetTimestamp(),
		Status:    "pending",
	}

	mp.AddTransaction(tx)

	// Get balance (should account for pending transaction)
	balance, err := bc.GetBalance("alice")
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	expectedBalance := uint64(1000 - 200 - 10) // confirmed - amount - fee
	if balance != expectedBalance {
		t.Errorf("Expected balance %d, got %d", expectedBalance, balance)
	}
}

// TestUpdateUserBalances tests balance updates from transactions
func TestUpdateUserBalances(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Give alice initial balance
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = db.IncreaseUserBalance(sqlTx, "alice", 1000)
	if err != nil {
		t.Fatalf("Failed to increase balance: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Create transactions
	transactions := []blockchain.Transaction{
		{
			From:      "alice",
			To:        "bob",
			Amount:    200,
			Fee:       10,
			Timestamp: utils.GetTimestamp(),
			Status:    "confirmed",
		},
		{
			To:         "miner",
			Amount:     blockchain.MiningReward,
			Fee:        0,
			Timestamp:  utils.GetTimestamp(),
			Status:     "confirmed",
			IsCoinbase: true,
		},
	}

	// Update balances
	sqlTx2, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx2.Rollback()

	err = bc.UpdateUserBalances(sqlTx2, transactions)
	if err != nil {
		t.Fatalf("Failed to update balances: %v", err)
	}

	err = sqlTx2.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Check alice's balance (1000 - 200 - 10 = 790)
	aliceBalance, err := db.GetConfirmedBalance("alice")
	if err != nil {
		t.Fatalf("Failed to get alice's balance: %v", err)
	}

	if aliceBalance != 790 {
		t.Errorf("Expected alice's balance 790, got %d", aliceBalance)
	}

	// Check Bob's balance (200)
	bobBalance, err := db.GetConfirmedBalance("bob")
	if err != nil {
		t.Fatalf("Failed to get bob's balance: %v", err)
	}

	if bobBalance != 200 {
		t.Errorf("Expected bob's balance 200, got %d", bobBalance)
	}

	// Check miner's balance (mining reward + fee = 10000 + 10 = 10010)
	minerBalance, err := db.GetConfirmedBalance("miner")
	if err != nil {
		t.Fatalf("Failed to get miner's balance: %v", err)
	}

	expectedMinerBalance := uint64(blockchain.MiningReward) + 10
	if minerBalance != expectedMinerBalance {
		t.Errorf("Expected miner's balance %d, got %d", expectedMinerBalance, minerBalance)
	}
}

// TestValidateTransaction tests transaction validation
func TestValidateTransaction(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Give alice balance
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = db.IncreaseUserBalance(sqlTx, "alice", 1000)
	if err != nil {
		t.Fatalf("Failed to increase balance: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Valid transaction
	validTx := &blockchain.Transaction{
		From:      "alice",
		To:        "bob",
		Amount:    200,
		Fee:       10,
		Timestamp: utils.GetTimestamp(),
	}

	err = bc.ValidateTransaction(validTx)
	if err != nil {
		t.Errorf("Valid transaction should pass validation: %v", err)
	}

	// Invalid transaction (insufficient balance)
	invalidTx := &blockchain.Transaction{
		From:      "alice",
		To:        "bob",
		Amount:    2000,
		Fee:       10,
		Timestamp: utils.GetTimestamp(),
	}

	err = bc.ValidateTransaction(invalidTx)
	if err == nil {
		t.Error("Transaction with insufficient balance should fail validation")
	}

	// Coinbase transaction (should always be valid)
	coinbaseTx := &blockchain.Transaction{
		To:         "miner",
		Amount:     blockchain.MiningReward,
		Timestamp:  utils.GetTimestamp(),
		IsCoinbase: true,
	}

	err = bc.ValidateTransaction(coinbaseTx)
	if err != nil {
		t.Errorf("Coinbase transaction should always be valid: %v", err)
	}
}

// TestVerifyBlocks tests verification of multiple blocks
func TestVerifyBlocks(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Get all blocks
	blocks, err := bc.GetAllBlocks()
	if err != nil {
		t.Fatalf("Failed to get all blocks: %v", err)
	}

	// Create a new blockchain instance to test verification
	bc2, err := blockchain.LoadBlockchain(db, blockchain.NewMempool(1048576))
	if err != nil {
		t.Fatalf("Failed to load blockchain: %v", err)
	}

	// Verify blocks
	isValid, err := bc2.VerifyBlocks(blocks)
	if err != nil {
		t.Fatalf("Error verifying blocks: %v", err)
	}

	if !isValid {
		t.Error("All blocks should be valid")
	}
}

// TestVerifyBlocks_EmptyChain tests verification of empty blockchain
func TestVerifyBlocks_EmptyChain(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.LoadBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to load blockchain: %v", err)
	}

	// Verify empty chain
	isValid, err := bc.VerifyBlocks([]blockchain.Block{})
	if err != nil {
		t.Fatalf("Error verifying empty chain: %v", err)
	}

	if !isValid {
		t.Error("Empty chain should be valid")
	}
}

// TestGetBlockById tests retrieving a block by ID
func TestGetBlockById(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Get genesis block
	block, err := bc.GetBlockById(0)
	if err != nil {
		t.Fatalf("Failed to get block by ID: %v", err)
	}

	if block.Id != 0 {
		t.Errorf("Expected block ID 0, got %d", block.Id)
	}

	// Try to get non-existent block
	_, err = bc.GetBlockById(999)
	if err == nil {
		t.Error("Should return error for non-existent block")
	}
}

// TestGetAllBlocks tests retrieving all blocks
func TestGetAllBlocks(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	blocks, err := bc.GetAllBlocks()
	if err != nil {
		t.Fatalf("Failed to get all blocks: %v", err)
	}

	if len(blocks) != 1 {
		t.Errorf("Expected 1 block (genesis), got %d", len(blocks))
	}

	// Add another block with unique transaction
	prevBlock := bc.GetLatestBlock()
	coinbaseTx := blockchain.CreateCoinbaseTx("miner5", blockchain.MiningReward)
	newBlock := &blockchain.Block{
		Id:           prevBlock.Id + 1,
		PrevHash:     prevBlock.Hash,
		Timestamp:    utils.GetTimestamp(),
		Transactions: []blockchain.Transaction{*coinbaseTx},
		Nonce:        0,
	}

	newBlock.ComputeMerkleRoot()
	err = newBlock.HashBlock()
	if err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = bc.AddBlock(sqlTx, newBlock)
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Get all blocks again
	blocks, err = bc.GetAllBlocks()
	if err != nil {
		t.Fatalf("Failed to get all blocks: %v", err)
	}

	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(blocks))
	}
}

// TestAddTransactionToDB tests adding transactions to database
func TestAddTransactionToDB(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Create a keypair for signing
	keyPair, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	// Create transaction
	tx := &blockchain.Transaction{
		From:      keyPair.Address,
		To:        "bob",
		Amount:    100,
		Fee:       10,
		Timestamp: utils.GetTimestamp(),
		Status:    "confirmed",
	}

	err = tx.Sign(keyPair)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Add transaction to database
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = bc.AddTransactionToDB(sqlTx, 1, []blockchain.Transaction{*tx})
	if err != nil {
		t.Fatalf("Failed to add transaction to database: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify transaction was added
	dbTxs, err := db.GetTransactionsByBlockId(1)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	if len(dbTxs) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(dbTxs))
	}

	if dbTxs[0].From != keyPair.Address {
		t.Errorf("Expected From %s, got %s", keyPair.Address, dbTxs[0].From)
	}

	if dbTxs[0].To != "bob" {
		t.Errorf("Expected To 'bob', got %s", dbTxs[0].To)
	}

	if dbTxs[0].Amount != 100 {
		t.Errorf("Expected Amount 100, got %d", dbTxs[0].Amount)
	}
}

// TestVerifyHeaders tests header verification
func TestVerifyHeaders(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Create valid headers
	genesisBlock := bc.GetLatestBlock()
	headers := []blockchain.BlockHeader{
		*genesisBlock.GetHeader(),
	}

	// Verify headers
	isValid, err := bc.VerifyHeaders(headers)
	if err != nil {
		t.Fatalf("Error verifying headers: %v", err)
	}

	if !isValid {
		t.Error("Valid headers should pass verification")
	}
}

// TestVerifyHeaders_EmptyList tests verification of empty header list
func TestVerifyHeaders_EmptyList(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Verify empty list
	isValid, err := bc.VerifyHeaders([]blockchain.BlockHeader{})
	if err != nil {
		t.Fatalf("Error verifying empty headers: %v", err)
	}

	if !isValid {
		t.Error("Empty header list should be valid")
	}
}

// TestVerifyHeaders_InvalidGenesisID tests verification with invalid genesis ID
func TestVerifyHeaders_InvalidGenesisID(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Create header with invalid genesis ID
	invalidHeader := blockchain.BlockHeader{
		Id:       1, // Should be 0 for genesis
		PrevHash: make([]byte, 32),
		Hash:     make([]byte, 32),
	}

	// Verify should fail
	isValid, err := bc.VerifyHeaders([]blockchain.BlockHeader{invalidHeader})
	if err == nil {
		t.Error("Expected error for invalid genesis ID")
	}

	if isValid {
		t.Error("Headers with invalid genesis ID should not be valid")
	}
}

// TestConcurrentAddBlockToMemory tests concurrent access to AddBlockToMemory
func TestConcurrentAddBlockToMemory(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Create multiple blocks with unique transactions
	prevBlock := bc.GetLatestBlock()
	blocks := make([]*blockchain.Block, 10)
	for i := 0; i < 10; i++ {
		// Each block gets a unique coinbase transaction with different timestamp
		coinbaseTx := blockchain.CreateCoinbaseTx("miner"+string(rune(i)), blockchain.MiningReward)
		blocks[i] = &blockchain.Block{
			Id:           prevBlock.Id + int64(i) + 1,
			PrevHash:     prevBlock.Hash,
			Timestamp:    utils.GetTimestamp() + int64(i)*1000, // Add milliseconds to make unique
			Transactions: []blockchain.Transaction{*coinbaseTx},
			Nonce:        int64(i),
		}
		blocks[i].ComputeMerkleRoot()
		err = blocks[i].HashBlock()
		if err != nil {
			t.Fatalf("Failed to hash block: %v", err)
		}
	}

	// Add blocks concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			bc.AddBlockToMemory(blocks[idx])
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all blocks were added
	if len(bc.Blocks) != 11 { // 1 genesis + 10 new blocks
		t.Errorf("Expected 11 blocks, got %d", len(bc.Blocks))
	}
}

// TestConcurrentGetLatestBlock tests concurrent access to GetLatestBlock
func TestConcurrentGetLatestBlock(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Get latest block concurrently
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			block := bc.GetLatestBlock()
			if block == nil {
				t.Error("Latest block should not be nil")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

// TestBlockchainWithMultipleTransactions tests blockchain with multiple transactions
func TestBlockchainWithMultipleTransactions(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Give alice initial balance
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = db.IncreaseUserBalance(sqlTx, "alice", 10000)
	if err != nil {
		t.Fatalf("Failed to increase balance: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Create keypair
	keyPair, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	// Create multiple transactions
	transactions := []blockchain.Transaction{}
	for i := 0; i < 5; i++ {
		tx := &blockchain.Transaction{
			From:      "alice",
			To:        "bob",
			Amount:    uint64(100 * (i + 1)),
			Fee:       10,
			Timestamp: utils.GetTimestamp() + int64(i),
			Status:    "confirmed",
		}
		err = tx.Sign(keyPair)
		if err != nil {
			t.Fatalf("Failed to sign transaction: %v", err)
		}
		transactions = append(transactions, *tx)
	}

	// Create new block with transactions
	prevBlock := bc.GetLatestBlock()
	newBlock := &blockchain.Block{
		Id:           prevBlock.Id + 1,
		PrevHash:     prevBlock.Hash,
		Timestamp:    utils.GetTimestamp(),
		Transactions: transactions,
		Nonce:        0,
	}

	newBlock.ComputeMerkleRoot()
	err = newBlock.HashBlock()
	if err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	// Add block to database
	sqlTx2, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx2.Rollback()

	err = bc.AddBlock(sqlTx2, newBlock)
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	err = sqlTx2.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify transactions were added
	dbTxs, err := db.GetTransactionsByBlockId(2) // Block ID 2 (after genesis)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	if len(dbTxs) != 5 {
		t.Errorf("Expected 5 transactions, got %d", len(dbTxs))
	}
}

// TestGetBalance_ConcurrentAccess tests concurrent balance queries
func TestGetBalance_ConcurrentAccess(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Set initial balance
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = db.IncreaseUserBalance(sqlTx, "alice", 1000)
	if err != nil {
		t.Fatalf("Failed to increase balance: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Query balance concurrently
	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func() {
			balance, err := bc.GetBalance("alice")
			if err != nil {
				t.Errorf("Failed to get balance: %v", err)
			}
			if balance != 1000 {
				t.Errorf("Expected balance 1000, got %d", balance)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		<-done
	}
}

// TestValidateTransaction_EdgeCases tests edge cases in transaction validation
func TestValidateTransaction_EdgeCases(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Give alice exact balance
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = db.IncreaseUserBalance(sqlTx, "alice", 210)
	if err != nil {
		t.Fatalf("Failed to increase balance: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Transaction with exact balance (should pass)
	exactTx := &blockchain.Transaction{
		From:      "alice",
		To:        "bob",
		Amount:    200,
		Fee:       10,
		Timestamp: utils.GetTimestamp(),
	}

	err = bc.ValidateTransaction(exactTx)
	if err != nil {
		t.Errorf("Transaction with exact balance should pass: %v", err)
	}

	// Transaction exceeding balance by 1 (should fail)
	exceedTx := &blockchain.Transaction{
		From:      "alice",
		To:        "bob",
		Amount:    201,
		Fee:       10,
		Timestamp: utils.GetTimestamp(),
	}

	err = bc.ValidateTransaction(exceedTx)
	if err == nil {
		t.Error("Transaction exceeding balance should fail")
	}

	// Zero amount transaction (should pass if balance covers fee)
	zeroTx := &blockchain.Transaction{
		From:      "alice",
		To:        "bob",
		Amount:    0,
		Fee:       10,
		Timestamp: utils.GetTimestamp(),
	}

	err = bc.ValidateTransaction(zeroTx)
	if err != nil {
		t.Errorf("Zero amount transaction should pass if balance covers fee: %v", err)
	}
}

// TestUpdateUserBalances_WithMultipleFees tests balance updates with multiple transaction fees
func TestUpdateUserBalances_WithMultipleFees(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Give alice and bob initial balances
	sqlTx, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx.Rollback()

	err = db.IncreaseUserBalance(sqlTx, "alice", 1000)
	if err != nil {
		t.Fatalf("Failed to increase alice's balance: %v", err)
	}

	err = db.IncreaseUserBalance(sqlTx, "bob", 500)
	if err != nil {
		t.Fatalf("Failed to increase bob's balance: %v", err)
	}

	err = sqlTx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Create multiple transactions with fees
	transactions := []blockchain.Transaction{
		{
			From:      "alice",
			To:        "bob",
			Amount:    100,
			Fee:       5,
			Timestamp: utils.GetTimestamp(),
			Status:    "confirmed",
		},
		{
			From:      "bob",
			To:        "alice",
			Amount:    50,
			Fee:       3,
			Timestamp: utils.GetTimestamp(),
			Status:    "confirmed",
		},
		{
			To:         "miner",
			Amount:     blockchain.MiningReward,
			Fee:        0,
			Timestamp:  utils.GetTimestamp(),
			Status:     "confirmed",
			IsCoinbase: true,
		},
	}

	// Update balances
	sqlTx2, err := db.BeginTx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer sqlTx2.Rollback()

	err = bc.UpdateUserBalances(sqlTx2, transactions)
	if err != nil {
		t.Fatalf("Failed to update balances: %v", err)
	}

	err = sqlTx2.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Check alice's balance: 1000 - 100 - 5 + 50 = 945
	aliceBalance, err := db.GetConfirmedBalance("alice")
	if err != nil {
		t.Fatalf("Failed to get alice's balance: %v", err)
	}

	if aliceBalance != 945 {
		t.Errorf("Expected alice's balance 945, got %d", aliceBalance)
	}

	// Check Bob's balance: 500 + 100 - 50 - 3 = 547
	bobBalance, err := db.GetConfirmedBalance("bob")
	if err != nil {
		t.Fatalf("Failed to get bob's balance: %v", err)
	}

	if bobBalance != 547 {
		t.Errorf("Expected bob's balance 547, got %d", bobBalance)
	}

	// Check miner's balance: mining reward + total fees = 10000 + 5 + 3 = 10008
	minerBalance, err := db.GetConfirmedBalance("miner")
	if err != nil {
		t.Fatalf("Failed to get miner's balance: %v", err)
	}

	expectedMinerBalance := uint64(blockchain.MiningReward) + 5 + 3
	if minerBalance != expectedMinerBalance {
		t.Errorf("Expected miner's balance %d, got %d", expectedMinerBalance, minerBalance)
	}
}
