package blockchain

import (
	"sort"
	"sync"
)

type Mempool struct {
	Transactions map[string]Transaction `json:"transactions"`
	MaxCapacity  int64                  `json:"max_capacity"` // 1MB
	Mutex        sync.RWMutex           `json:"-"`
}

const BaseTxFee = 1

func NewMempool(maxCapacity int64) *Mempool {
	// in bytes
	if maxCapacity == 0 {
		maxCapacity = 1048576 // 1 MegaByte
	}

	return &Mempool{
		Transactions: make(map[string]Transaction),
		MaxCapacity:  maxCapacity,
	}
}

func (mp *Mempool) AddTransaction(tx *Transaction) {
	mp.Mutex.Lock()
	defer mp.Mutex.Unlock()

	hash := tx.Hash().EncodeToString()

	mp.Transactions[hash] = *tx
}

// GetTransactionsCopy returns a deep copy of mempool transactions (thread-safe)
func (mp *Mempool) GetTransactionsCopy() map[string]Transaction {
	mp.Mutex.RLock()
	defer mp.Mutex.RUnlock()

	deelCopy := make(map[string]Transaction, len(mp.Transactions))
	for k, v := range mp.Transactions {
		deelCopy[k] = v
	}
	return deelCopy
}

// SortTxsByFee -> Sort transactions in DESC order by their fee
func (mp *Mempool) SortTxsByFee(txs map[string]Transaction) []Transaction {
	sortedTxs := make([]Transaction, 0, len(txs))
	for _, tx := range txs {
		sortedTxs = append(sortedTxs, tx)
	}

	sort.Slice(sortedTxs, func(i, j int) bool {
		return sortedTxs[i].Fee > sortedTxs[j].Fee
	})

	return sortedTxs
}

func (mp *Mempool) CalculateTxFee() uint64 {
	mp.Mutex.RLock()
	defer mp.Mutex.RUnlock()

	congestion := mp.GetCongestion()

	switch congestion {
	case 0: // Low
		return BaseTxFee
	case 1: // Medium
		return BaseTxFee * 2
	case 2: // High
		return BaseTxFee * 4
	default:
		return BaseTxFee
	}
}

func (mp *Mempool) CalculateFee(tx *Transaction) uint64 {
	txFeeRate := mp.CalculateTxFee() // satoshis per byte
	txSize := uint64(tx.Size())

	// total fee = rate * size
	fee := txFeeRate * txSize
	if fee == 0 {
		fee = 1
	}
	return fee
}

func (mp *Mempool) SyncMempool(syncCandidateMempool *Mempool) {
	candidateTxs := syncCandidateMempool.GetTransactionsCopy()

	mp.Mutex.Lock()
	defer mp.Mutex.Unlock()

	if len(candidateTxs) == 0 {
		mp.Clear()
		return
	}

	mp.syncTransactions(candidateTxs)
}

func (mp *Mempool) syncTransactions(newTxs map[string]Transaction) {
	for hash, tx := range newTxs {
		if _, exists := mp.Transactions[hash]; !exists {
			mp.Transactions[hash] = tx
		}
	}
}

func (mp *Mempool) IsEmpty() bool {
	return len(mp.Transactions) == 0
}

func (mp *Mempool) Clear() {
	mp.Transactions = make(map[string]Transaction)
}

func (mp *Mempool) DeleteMinedTransactions(blockTransactions []Transaction) {
	mp.Mutex.Lock()
	defer mp.Mutex.Unlock()

	for _, tx := range blockTransactions {
		// Coinbase transaction are not in the mempool
		if tx.IsCoinbase {
			continue
		}

		hash := tx.Hash().EncodeToString()
		delete(mp.Transactions, hash)
	}
}

func (mp *Mempool) CalculateCurrentSize() int {
	size := 0
	for _, tx := range mp.Transactions {
		size += tx.Size()
	}
	return size
}

// GetCongestion -> Low = 0, Medium = 1, High = 2
func (mp *Mempool) GetCongestion() int {
	currentSize := mp.CalculateCurrentSize()
	ratio := float64(currentSize) / float64(mp.MaxCapacity)

	switch {
	case ratio < 0.25:
		return 0 // Low
	case ratio < 0.75:
		return 1 // Medium
	default:
		return 2 // High
	}
}
