package tests

import (
	"simple_blockchain/pkg/blockchain"
	"testing"
	"time"
)

func TestBlock_HashBlock(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    1234567890,
		Nonce:        0,
		Transactions: []blockchain.Transaction{},
	}

	err := block.HashBlock()
	if err != nil {
		t.Errorf("HashBlock failed: %v", err)
	}

	if len(block.Hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(block.Hash))
	}

	if block.MerkleRoot == nil {
		t.Error("MerkleRoot should be computed during HashBlock")
	}
}

func TestBlock_AddTransaction(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
	}

	tx := createMockTransaction("Alice", "Bob", 100)
	block.AddTransaction(tx)

	if len(block.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(block.Transactions))
	}
}

func TestBlock_ValidateBlock(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Nonce:        0,
		Transactions: []blockchain.Transaction{},
	}

	// Hash the block first
	err := block.HashBlock()
	if err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	// Test validation
	isValid := block.ValidateBlock(make([]byte, 32))
	if !isValid {
		t.Error("Block should be valid")
	}

	// Test with wrong previous hash
	wrongPrevHash := make([]byte, 32)
	wrongPrevHash[0] = 1
	isValid = block.ValidateBlock(wrongPrevHash)
	if isValid {
		t.Error("Block should be invalid with wrong previous hash")
	}
}

func TestBlock_CalculateSize(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
	}

	// Add some transactions
	tx1 := createMockTransaction("Alice", "Bob", 100)
	tx2 := createMockTransaction("Bob", "Charlie", 200)
	block.AddTransaction(tx1)
	block.AddTransaction(tx2)

	size := block.CalculateSize()
	if size <= 0 {
		t.Error("Block size should be greater than 0")
	}

	// Size should increase with more transactions
	tx3 := createMockTransaction("Charlie", "Dave", 300)
	block.AddTransaction(tx3)
	newSize := block.CalculateSize()

	if newSize <= size {
		t.Error("Block size should increase with more transactions")
	}
}

func TestBlock_MerkleRootCalculation(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
	}

	// Test with no transactions
	err := block.HashBlock()
	if err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	if block.MerkleRoot == nil {
		t.Error("MerkleRoot should be calculated even with no transactions")
	}

	// Test with transactions
	tx1 := createMockTransaction("Alice", "Bob", 100)
	tx2 := createMockTransaction("Bob", "Charlie", 200)
	block.AddTransaction(tx1)
	block.AddTransaction(tx2)

	err = block.HashBlock()
	if err != nil {
		t.Fatalf("Failed to hash block with transactions: %v", err)
	}

	if block.MerkleRoot == nil {
		t.Error("MerkleRoot should be calculated with transactions")
	}
}

// Helper function to create mock transactions
func createMockTransaction(from, to string, amount int64) blockchain.Transaction {
	return blockchain.Transaction{
		From:      from,
		To:        to,
		Timestamp: time.Now().Unix(),
	}
}
