package blockchain

import (
	"sort"
	"sync"
)

type Mempool struct {
	transactions []Transaction
	Mutex        sync.RWMutex
}

const BaseTxFee = 25  // 0.25%
const HighTxFee = 200 // 2%

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

// SortTxsByFee -> sort transactions in DESC order by their fee
func (mp *Mempool) SortTxsByFee(txs *[]Transaction) {
	sort.Slice(*txs, func(i, j int) bool {
		return (*txs)[i].Fee > (*txs)[j].Fee
	})
}

func (mp *Mempool) Clear() {
	mp.Mutex.Lock()
	defer mp.Mutex.Unlock()

	mp.transactions = []Transaction{}
}

func (mp *Mempool) CalculateTxFee() uint64 {
	mp.Mutex.RLock()
	defer mp.Mutex.RLock()

	if len(mp.transactions) > 100 {
		return HighTxFee
	}

	return BaseTxFee
}

func (mp *Mempool) CalculateFee(amount uint64) uint64 {
	txFee := mp.CalculateTxFee()

	feeAmount := (amount * txFee) / 10000

	if feeAmount == 0 {
		feeAmount = 1
	}

	return feeAmount
}
