package tests

import (
	"simple_blockchain/pkg/blockchain"
	"testing"
	"time"
)

func TestAddTxMempool(t *testing.T) {
	from := "Alice"
	to := "Bob"
	amount := uint64(50)
	timestamp := time.Now().UTC().UnixMilli()

	newTx := blockchain.NewTransaction(from, to, amount, timestamp)
	newMp := blockchain.NewMempool(1_000_000)

	newMp.AddTransaction(newTx)

	if len(newMp.Transactions) != 1 {
		t.Errorf("Expected 1 transaction in mempool, got %d", len(newMp.Transactions))
	}

	hash := newTx.Hash().EncodeToString()
	if _, exists := newMp.Transactions[hash]; !exists {
		t.Errorf("Transaction with hash %s not found in mempool", hash)
	}

	if newMp.Transactions[hash].From != from {
		t.Errorf("Expected From %s, got %s", from, newMp.Transactions[hash].From)
	}

	if newMp.Transactions[hash].To != to {
		t.Errorf("Expected To %s, got %s", to, newMp.Transactions[hash].To)
	}

	if newMp.Transactions[hash].Amount != amount {
		t.Errorf("Expected Amount %d, got %d", amount, newMp.Transactions[hash].Amount)
	}

	if newMp.Transactions[hash].Timestamp != timestamp {
		t.Errorf("Expected Timestamp %d, got %d", timestamp, newMp.Transactions[hash].Timestamp)
	}
}

func createMockMempoolTransaction(from, to string, amount uint64) *blockchain.Transaction {
	return blockchain.NewTransaction(from, to, amount, time.Now().UTC().UnixMilli())
}

func TestNewMempool(t *testing.T) {
	tests := []struct {
		name         string
		capacity     int64
		wantCapacity int64
	}{
		{"Default capacity", 0, 1048576},
		{"Custom capacity", 2000, 2000},
		{"Negative capacity", -1, 1048576},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := blockchain.NewMempool(tt.capacity)
			if mp.MaxCapacity != tt.wantCapacity {
				t.Errorf("NewMempool() capacity = %v, want %v", mp.MaxCapacity, tt.wantCapacity)
			}
		})
	}
}

func TestAddTransactionToMempool(t *testing.T) {
	mp := blockchain.NewMempool(1000000)
	tx := createMockMempoolTransaction("Alice", "Bob", 100)

	mp.AddTransaction(tx)
	hash := tx.Hash().EncodeToString()

	if len(mp.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(mp.Transactions))
	}

	if _, exists := mp.Transactions[hash]; !exists {
		t.Errorf("Transaction not found in mempool")
	}
}

func TestRemoveTransactionFromMempool(t *testing.T) {
	mp := blockchain.NewMempool(1000000)
	tx := createMockMempoolTransaction("Alice", "Bob", 100)

	mp.AddTransaction(tx)
	hash := tx.Hash().EncodeToString()

	mp.RemoveTransaction(hash)

	if len(mp.Transactions) != 0 {
		t.Errorf("Expected empty mempool, got %d transactions", len(mp.Transactions))
	}

	if _, exists := mp.Transactions[hash]; exists {
		t.Errorf("Transaction should have been removed from mempool")
	}
}

func TestMempoolCapacityLimit(t *testing.T) {
	// Create transactions first to calculate their sizes
	tx1 := createMockMempoolTransaction("Alice", "Bob", 100)
	tx2 := createMockMempoolTransaction("Bob", "Charlie", 200)
	tx3 := createMockMempoolTransaction("Charlie", "Dave", 300)

	// Calculate sizes
	tx1Size := tx1.Size()
	tx2Size := tx2.Size()

	// Set capacity to fit only tx1 + tx2, but not tx3
	maxCapacity := int64(tx1Size + tx2Size + 5) // Add small buffer
	mp := blockchain.NewMempool(maxCapacity)

	// First transaction should not exceed capacity
	if mp.WillExceedCapacity(tx1) {
		t.Error("First transaction should not exceed capacity")
	}
	mp.AddTransaction(tx1)

	// Second transaction should not exceed capacity
	if mp.WillExceedCapacity(tx2) {
		t.Error("Second transaction should not exceed capacity")
	}
	mp.AddTransaction(tx2)

	// Verify we have 2 transactions
	if len(mp.Transactions) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(mp.Transactions))
	}

	// Third transaction should exceed capacity
	if !mp.WillExceedCapacity(tx3) {
		t.Error("Third transaction should exceed capacity")
	}

	// Verify current capacity is within limits
	currentSize := mp.CalculateCurrentSize()
	if currentSize > int(maxCapacity) {
		t.Errorf("Current mempool size %d exceeds capacity %d", currentSize, maxCapacity)
	}
}

func TestGetTransactionByHash(t *testing.T) {
	mp := blockchain.NewMempool(1000000)
	tx := createMockMempoolTransaction("Alice", "Bob", 100)

	mp.AddTransaction(tx)
	hash := tx.Hash().EncodeToString()

	fetchedTx, exists := mp.GetTransaction(hash)
	if !exists {
		t.Error("Expected transaction to exist in mempool")
	}

	if fetchedTx.From != tx.From || fetchedTx.To != tx.To || fetchedTx.Amount != tx.Amount {
		t.Error("Fetched transaction doesn't match original")
	}
}

func TestMempoolClear(t *testing.T) {
	mp := blockchain.NewMempool(1000000)

	tx1 := createMockMempoolTransaction("Alice", "Bob", 100)
	tx2 := createMockMempoolTransaction("Bob", "Charlie", 200)

	mp.AddTransaction(tx1)
	mp.AddTransaction(tx2)

	mp.Clear()

	if len(mp.Transactions) != 0 {
		t.Errorf("Expected empty mempool after clear, got %d transactions", len(mp.Transactions))
	}
}

func TestAddDuplicateTransaction(t *testing.T) {
	mp := blockchain.NewMempool(1000000)
	tx := createMockMempoolTransaction("Alice", "Bob", 100)

	mp.AddTransaction(tx)
	mp.AddTransaction(tx) // Add same transaction twice

	if len(mp.Transactions) != 1 {
		t.Errorf("Expected 1 transaction after duplicate add, got %d", len(mp.Transactions))
	}
}

func TestAddNilTransaction(t *testing.T) {
	mp := blockchain.NewMempool(1000000)

	// This should handle nil gracefully
	mp.AddTransaction(nil)

	if len(mp.Transactions) != 0 {
		t.Errorf("Expected 0 transactions after adding nil, got %d", len(mp.Transactions))
	}
}

func TestCalculateCurrentSize(t *testing.T) {
	mp := blockchain.NewMempool(1000000)

	// Test empty mempool
	if mp.CalculateCurrentSize() != 0 {
		t.Error("Empty mempool should have size 0")
	}

	tx1 := createMockMempoolTransaction("Alice", "Bob", 100)
	tx2 := createMockMempoolTransaction("Bob", "Charlie", 200)

	mp.AddTransaction(tx1)
	size1 := mp.CalculateCurrentSize()

	mp.AddTransaction(tx2)
	size2 := mp.CalculateCurrentSize()

	if size2 <= size1 {
		t.Error("Size should increase after adding second transaction")
	}

	expectedSize := tx1.Size() + tx2.Size()
	if size2 != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size2)
	}
}

func TestWillExceedCapacityNilTransaction(t *testing.T) {
	mp := blockchain.NewMempool(1000)

	// Should handle nil gracefully without panicking
	result := mp.WillExceedCapacity(nil)
	if result {
		t.Error("Nil transaction should not exceed capacity")
	}
}

func TestWillExceedCapacityEmptyMempool(t *testing.T) {
	mp := blockchain.NewMempool(20)                                 // Very small capacity
	largeTx := createMockMempoolTransaction("Alice", "Bob", 999999) // Large transaction

	if !mp.WillExceedCapacity(largeTx) {
		t.Error("Large transaction should exceed capacity on empty mempool")
	}
}
