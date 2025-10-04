package blockchain

import (
	"sort"
	"sync"
)

type Mempool struct {
	Transactions map[string]Transaction `json:"transactions"`
	Mutex        sync.RWMutex           `json:"-"`
}

const BaseTxFee = 25  // 0.25%
const HighTxFee = 200 // 2%

func NewMempool() *Mempool {
	return &Mempool{
		Transactions: make(map[string]Transaction),
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

	if len(mp.Transactions) > 100 {
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
