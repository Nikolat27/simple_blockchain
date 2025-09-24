package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const Difficulty = 4

type Block struct {
	Index        int           `json:"index"`
	PrevHash     []byte        `json:"prev_hash"`
	Hash         []byte        `json:"hash"`
	Timestamp    time.Time     `json:"timestamp"`
	Nonce        int           `json:"nonce"`
	Transactions []Transaction `json:"transactions"`
}

func (block *Block) HashBlock() {
	prevHashStr := hex.EncodeToString(block.PrevHash)

	record := fmt.Sprintf("%d-%s-%s-%d-%v", block.Index, prevHashStr,
		block.Timestamp, block.Nonce, block.Transactions)

	hash := sha256.Sum256([]byte(record))

	block.Hash = hash[:]
}

func (block *Block) IsValidHash() bool {
	hashStr := hex.EncodeToString(block.Hash)
	return strings.HasPrefix(hashStr, strings.Repeat("0", Difficulty))
}
