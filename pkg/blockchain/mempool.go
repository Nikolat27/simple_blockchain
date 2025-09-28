package blockchain

import (
	"sync"
)

type Mempool struct {
	transactions []Transaction
	Mutex        sync.RWMutex // Protects concurrent access to transactions
}

func NewMempool() *Mempool {
	return &Mempool{
		transactions: make([]Transaction, 0),
	}
}

func (mp *Mempool) AddTransaction(tx *Transaction) {
	mp.Mutex.Lock()
	defer mp.Mutex.Unlock()

	mp.transactions = append(mp.transactions, *tx)
}

func (mp *Mempool) GetTransactions() []Transaction {
	mp.Mutex.RLock()
	defer mp.Mutex.RUnlock()

	// Return a copy to prevent external modification
	transactions := make([]Transaction, len(mp.transactions))
	copy(transactions, mp.transactions)

	return transactions
}

func (mp *Mempool) Clear() {
	mp.Mutex.Lock()
	defer mp.Mutex.Unlock()

	mp.transactions = []Transaction{}
}
