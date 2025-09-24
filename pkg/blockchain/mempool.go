package blockchain

import "sync"

type Mempool struct {
	transactions []Transaction
	mu           sync.RWMutex // Protects concurrent access to transactions
}

func NewMempool() *Mempool {
	return &Mempool{
		transactions: make([]Transaction, 0),
	}
}

func (mp *Mempool) AddTransaction(tx *Transaction) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.transactions = append(mp.transactions, *tx)
}

func (mp *Mempool) GetTransactions() []Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	// Return a copy to prevent external modification
	transactions := make([]Transaction, len(mp.transactions))
	copy(transactions, mp.transactions)
	return transactions
}

func (mp *Mempool) Clear() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.transactions = []Transaction{}
}
