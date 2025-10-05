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

func TestGetTransactionsCopy(t *testing.T) {
	mp := blockchain.NewMempool(1000000)

	tx1 := createMockMempoolTransaction("Alice", "Bob", 100)
	tx2 := createMockMempoolTransaction("Bob", "Charlie", 200)

	mp.AddTransaction(tx1)
	mp.AddTransaction(tx2)

	// Get copy of transactions
	txsCopy := mp.GetTransactionsCopy()

	if len(txsCopy) != 2 {
		t.Errorf("Expected 2 transactions in copy, got %d", len(txsCopy))
	}

	// Verify it's a deep copy by modifying the copy
	for hash := range txsCopy {
		delete(txsCopy, hash)
		break
	}

	// Original mempool should still have 2 transactions
	if len(mp.Transactions) != 2 {
		t.Error("Modifying copy should not affect original mempool")
	}
}

func TestSortTxsByFee(t *testing.T) {
	mp := blockchain.NewMempool(1000000)

	// Create transactions with different fees
	tx1 := createMockMempoolTransaction("Alice", "Bob", 100)
	tx1.Fee = 10

	tx2 := createMockMempoolTransaction("Bob", "Charlie", 200)
	tx2.Fee = 50

	tx3 := createMockMempoolTransaction("Charlie", "Dave", 300)
	tx3.Fee = 25

	mp.AddTransaction(tx1)
	mp.AddTransaction(tx2)
	mp.AddTransaction(tx3)

	// Get transactions and sort by fee
	txs := mp.GetTransactionsCopy()
	sortedTxs := mp.SortTxsByFee(txs)

	if len(sortedTxs) != 3 {
		t.Errorf("Expected 3 sorted transactions, got %d", len(sortedTxs))
	}

	// Verify they are sorted in descending order by fee
	if sortedTxs[0].Fee != 50 {
		t.Errorf("First transaction should have fee 50, got %d", sortedTxs[0].Fee)
	}

	if sortedTxs[1].Fee != 25 {
		t.Errorf("Second transaction should have fee 25, got %d", sortedTxs[1].Fee)
	}

	if sortedTxs[2].Fee != 10 {
		t.Errorf("Third transaction should have fee 10, got %d", sortedTxs[2].Fee)
	}
}

func TestCalculateTxFee(t *testing.T) {
	tests := []struct {
		name               string
		setupMempool       func(*blockchain.Mempool)
		expectedFee        uint64
		expectedCongestion int
	}{
		{
			name: "Low congestion",
			setupMempool: func(mp *blockchain.Mempool) {
				// Add small transaction (< 25% capacity)
				tx := createMockMempoolTransaction("Alice", "Bob", 100)
				mp.AddTransaction(tx)
			},
			expectedFee:        blockchain.BaseTxFee,
			expectedCongestion: 0,
		},
		{
			name: "Medium congestion",
			setupMempool: func(mp *blockchain.Mempool) {
				// Fill mempool to ~50% capacity
				for i := 0; i < 500; i++ {
					tx := createMockMempoolTransaction("Alice", "Bob", uint64(i))
					mp.AddTransaction(tx)
				}
			},
			expectedFee:        blockchain.BaseTxFee * 2,
			expectedCongestion: 1,
		},
		{
			name: "High congestion",
			setupMempool: func(mp *blockchain.Mempool) {
				// Fill mempool to > 75% capacity
				for i := 0; i < 2000; i++ {
					tx := createMockMempoolTransaction("Alice", "Bob", uint64(i))
					mp.AddTransaction(tx)
				}
			},
			expectedFee:        blockchain.BaseTxFee * 4,
			expectedCongestion: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := blockchain.NewMempool(100000) // Set capacity for testing
			tt.setupMempool(mp)

			fee := mp.CalculateTxFee()
			if fee != tt.expectedFee {
				t.Errorf("Expected fee %d, got %d", tt.expectedFee, fee)
			}

			congestion := mp.GetCongestion()
			if congestion != tt.expectedCongestion {
				t.Errorf("Expected congestion %d, got %d", tt.expectedCongestion, congestion)
			}
		})
	}
}

func TestCalculateFee(t *testing.T) {
	mp := blockchain.NewMempool(1000000)

	tx := createMockMempoolTransaction("Alice", "Bob", 100)

	fee := mp.CalculateFee(tx)

	// Fee should be at least 1
	if fee < 1 {
		t.Errorf("Fee should be at least 1, got %d", fee)
	}

	// Fee should be based on transaction size
	expectedMinFee := uint64(tx.Size()) * blockchain.BaseTxFee
	if fee < expectedMinFee {
		t.Errorf("Fee should be at least %d (size * base fee), got %d", expectedMinFee, fee)
	}
}

func TestIsEmpty(t *testing.T) {
	mp := blockchain.NewMempool(1000000)

	// Test empty mempool
	if !mp.IsEmpty() {
		t.Error("New mempool should be empty")
	}

	// Add transaction
	tx := createMockMempoolTransaction("Alice", "Bob", 100)
	mp.AddTransaction(tx)

	if mp.IsEmpty() {
		t.Error("Mempool with transaction should not be empty")
	}

	// Clear mempool
	mp.Clear()

	if !mp.IsEmpty() {
		t.Error("Cleared mempool should be empty")
	}
}

func TestDeleteMinedTransactions(t *testing.T) {
	mp := blockchain.NewMempool(1000000)

	// Add regular transactions
	tx1 := createMockMempoolTransaction("Alice", "Bob", 100)
	tx2 := createMockMempoolTransaction("Bob", "Charlie", 200)
	tx3 := createMockMempoolTransaction("Charlie", "Dave", 300)

	mp.AddTransaction(tx1)
	mp.AddTransaction(tx2)
	mp.AddTransaction(tx3)

	if len(mp.Transactions) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(mp.Transactions))
	}

	// Create coinbase transaction
	coinbaseTx := blockchain.CreateCoinbaseTx("miner", blockchain.MiningReward)

	// Simulate mining: delete tx1 and tx2 from mempool
	minedTxs := []blockchain.Transaction{*tx1, *tx2, *coinbaseTx}
	mp.DeleteMinedTransactions(minedTxs)

	// Only tx3 should remain
	if len(mp.Transactions) != 1 {
		t.Errorf("Expected 1 transaction remaining, got %d", len(mp.Transactions))
	}

	// Verify tx3 is still there
	hash3 := tx3.Hash().EncodeToString()
	if _, exists := mp.Transactions[hash3]; !exists {
		t.Error("Transaction 3 should still be in mempool")
	}

	// Verify tx1 and tx2 are gone
	hash1 := tx1.Hash().EncodeToString()
	if _, exists := mp.Transactions[hash1]; exists {
		t.Error("Transaction 1 should be removed from mempool")
	}

	hash2 := tx2.Hash().EncodeToString()
	if _, exists := mp.Transactions[hash2]; exists {
		t.Error("Transaction 2 should be removed from mempool")
	}
}

func TestSyncMempool(t *testing.T) {
	mp1 := blockchain.NewMempool(1000000)
	mp2 := blockchain.NewMempool(1000000)

	// Add transactions to mp1
	tx1 := createMockMempoolTransaction("Alice", "Bob", 100)
	tx2 := createMockMempoolTransaction("Bob", "Charlie", 200)
	mp1.AddTransaction(tx1)
	mp1.AddTransaction(tx2)

	// Add different transaction to mp2
	tx3 := createMockMempoolTransaction("Charlie", "Dave", 300)
	mp2.AddTransaction(tx3)

	// Sync mp2 with mp1
	mp2.SyncMempool(mp1)

	// mp2 should now have all three transactions
	if len(mp2.Transactions) != 3 {
		t.Errorf("Expected 3 transactions after sync, got %d", len(mp2.Transactions))
	}

	// Verify all transactions are present
	hash1 := tx1.Hash().EncodeToString()
	hash2 := tx2.Hash().EncodeToString()
	hash3 := tx3.Hash().EncodeToString()

	if _, exists := mp2.Transactions[hash1]; !exists {
		t.Error("Transaction 1 should be in synced mempool")
	}

	if _, exists := mp2.Transactions[hash2]; !exists {
		t.Error("Transaction 2 should be in synced mempool")
	}

	if _, exists := mp2.Transactions[hash3]; !exists {
		t.Error("Transaction 3 should be in synced mempool")
	}
}

func TestSyncMempoolWithEmpty(t *testing.T) {
	mp1 := blockchain.NewMempool(1000000)
	mp2 := blockchain.NewMempool(1000000)

	// Add transactions to mp1
	tx1 := createMockMempoolTransaction("Alice", "Bob", 100)
	tx2 := createMockMempoolTransaction("Bob", "Charlie", 200)
	mp1.AddTransaction(tx1)
	mp1.AddTransaction(tx2)

	// Sync mp1 with empty mp2
	mp1.SyncMempool(mp2)

	// mp1 should be cleared
	if len(mp1.Transactions) != 0 {
		t.Errorf("Expected empty mempool after syncing with empty mempool, got %d transactions", len(mp1.Transactions))
	}
}

func TestGetCongestion(t *testing.T) {
	tests := []struct {
		name               string
		capacity           int64
		fillPercentage     float64
		expectedCongestion int
	}{
		{"Low congestion - 10%", 100000, 0.10, 0},
		{"Low congestion - 24%", 100000, 0.24, 0},
		{"Medium congestion - 25%", 100000, 0.25, 1},
		{"Medium congestion - 50%", 100000, 0.50, 1},
		{"Medium congestion - 74%", 100000, 0.74, 1},
		{"High congestion - 75%", 100000, 0.75, 2},
		{"High congestion - 90%", 100000, 0.90, 2},
		{"High congestion - 100%", 100000, 1.00, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := blockchain.NewMempool(tt.capacity)

			// Fill mempool to desired percentage
			targetSize := int(float64(tt.capacity) * tt.fillPercentage)
			for mp.CalculateCurrentSize() < targetSize {
				tx := createMockMempoolTransaction("Alice", "Bob", uint64(mp.CalculateCurrentSize()))
				if mp.WillExceedCapacity(tx) {
					break
				}
				mp.AddTransaction(tx)
			}

			congestion := mp.GetCongestion()
			if congestion != tt.expectedCongestion {
				t.Errorf("Expected congestion %d, got %d (size: %d, capacity: %d)",
					tt.expectedCongestion, congestion, mp.CalculateCurrentSize(), tt.capacity)
			}
		})
	}
}

func TestWillExceedCapacityWithExistingTransaction(t *testing.T) {
	mp := blockchain.NewMempool(1000)

	tx := createMockMempoolTransaction("Alice", "Bob", 100)
	mp.AddTransaction(tx)

	// Adding the same transaction again should not exceed capacity
	if mp.WillExceedCapacity(tx) {
		t.Error("Adding existing transaction should not exceed capacity")
	}
}
