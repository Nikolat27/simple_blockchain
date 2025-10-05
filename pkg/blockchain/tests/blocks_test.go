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

	if err := block.HashBlock(); err != nil {
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
	if err := block.HashBlock(); err != nil {
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
	if err := block.HashBlock(); err != nil {
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

	if err := block.HashBlock(); err != nil {
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

func TestBlock_GetHeader(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Hash:         make([]byte, 32),
		MerkleRoot:   make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Nonce:        12345,
		Transactions: []blockchain.Transaction{},
	}

	header := block.GetHeader()

	if header == nil {
		t.Fatal("Header should not be nil")
	}

	if header.Id != block.Id {
		t.Errorf("Header ID should be %d, got %d", block.Id, header.Id)
	}

	if header.Nonce != block.Nonce {
		t.Errorf("Header Nonce should be %d, got %d", block.Nonce, header.Nonce)
	}

	if header.Timestamp != block.Timestamp {
		t.Errorf("Header Timestamp should be %d, got %d", block.Timestamp, header.Timestamp)
	}
}

func TestBlock_IsValidHash(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Nonce:        0,
		Transactions: []blockchain.Transaction{},
	}

	// Hash the block
	if err := block.HashBlock(); err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	// Most likely won't have valid hash with nonce 0
	// This test just ensures the function doesn't panic
	_ = block.IsValidHash()
}

func TestBlock_SerializeTransactions(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
	}

	// Test with no transactions
	serialized := block.SerializeTransactions()
	if serialized != "" {
		t.Error("Serialized empty transactions should be empty string")
	}

	// Add transactions
	tx1 := createMockTransaction("Alice", "Bob", 100)
	tx2 := createMockTransaction("Bob", "Charlie", 200)
	block.AddTransaction(tx1)
	block.AddTransaction(tx2)

	serialized = block.SerializeTransactions()
	if serialized == "" {
		t.Error("Serialized transactions should not be empty")
	}

	// Verify deterministic serialization
	serialized2 := block.SerializeTransactions()
	if serialized != serialized2 {
		t.Error("Serialization should be deterministic")
	}
}

func TestBlock_ComputeMerkleRoot(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{},
	}

	// Test with empty transactions
	block.ComputeMerkleRoot()
	if block.MerkleRoot == nil {
		t.Error("MerkleRoot should not be nil for empty transactions")
	}

	if len(block.MerkleRoot) != 32 {
		t.Errorf("MerkleRoot should be 32 bytes, got %d", len(block.MerkleRoot))
	}

	emptyMerkleRoot := make([]byte, 32)
	copy(emptyMerkleRoot, block.MerkleRoot)

	// Add one transaction
	tx1 := createMockTransaction("Alice", "Bob", 100)
	block.AddTransaction(tx1)

	// MerkleRoot should change
	if string(block.MerkleRoot) == string(emptyMerkleRoot) {
		t.Error("MerkleRoot should change when transaction is added")
	}

	// Add another transaction
	tx2 := createMockTransaction("Bob", "Charlie", 200)
	block.AddTransaction(tx2)

	// MerkleRoot should be 32 bytes
	if len(block.MerkleRoot) != 32 {
		t.Errorf("MerkleRoot should be 32 bytes, got %d", len(block.MerkleRoot))
	}
}

func TestBlock_HashBlockConsistency(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    1234567890,
		Nonce:        42,
		Transactions: []blockchain.Transaction{},
	}

	// Hash the block twice
	if err := block.HashBlock(); err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	hash1 := make([]byte, len(block.Hash))
	copy(hash1, block.Hash)

	if err := block.HashBlock(); err != nil {
		t.Fatalf("Failed to hash block second time: %v", err)
	}

	// Hashes should be identical
	if string(hash1) != string(block.Hash) {
		t.Error("Hash should be consistent across multiple calls")
	}
}

func TestBlock_ValidateBlockWithDifferentPrevHash(t *testing.T) {
	block := &blockchain.Block{
		Id:           1,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Nonce:        0,
		Transactions: []blockchain.Transaction{},
	}

	// Set a specific prev hash
	for i := range block.PrevHash {
		block.PrevHash[i] = byte(i)
	}

	if err := block.HashBlock(); err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	// Validate with same prev hash
	if !block.ValidateBlock(block.PrevHash) {
		t.Error("Block should be valid with correct prev hash")
	}

	// Create different prev hash
	differentPrevHash := make([]byte, 32)
	for i := range differentPrevHash {
		differentPrevHash[i] = byte(i + 1)
	}

	// Validate with different prev hash
	if block.ValidateBlock(differentPrevHash) {
		t.Error("Block should be invalid with incorrect prev hash")
	}
}
