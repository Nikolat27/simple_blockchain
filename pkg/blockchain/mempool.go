package blockchain

import (
	"sort"
	"sync"
)

type Mempool struct {
	Transactions map[string]Transaction `json:"transactions"`
	Mutex        sync.RWMutex
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

func (mp *Mempool) GetTransactions() map[string]Transaction {
	mp.Mutex.RLock()
	defer mp.Mutex.RUnlock()

	return mp.Transactions
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

	mp.Transactions = make(map[string]Transaction)
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
