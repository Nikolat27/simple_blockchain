package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"simple_blockchain/pkg/database"
	"strings"
	"time"
)

type Block struct {
	Id           int64         `json:"id"`
	PrevHash     []byte        `json:"prev_hash"`
	Hash         []byte        `json:"Hash"`
	MerkleRoot   []byte        `json:"merkle_root"`
	Timestamp    int64         `json:"timestamp"`
	Nonce        int64         `json:"nonce"`
	Transactions []Transaction `json:"transactions,omitempty"`
}

type BlockHeader struct {
	Id         int64  `json:"id"`
	PrevHash   []byte `json:"prev_hash"`
	Hash       []byte `json:"Hash"`
	MerkleRoot []byte `json:"merkle_root"`
	Timestamp  int64  `json:"timestamp"`
	Nonce      int64  `json:"nonce"`
}

func (block *Block) HashBlock() error {
	if block.MerkleRoot == nil {
		block.ComputeMerkleRoot()
	}

	merkleStr := hex.EncodeToString(block.MerkleRoot)

	prevHashStr := hex.EncodeToString(block.PrevHash)

	record := fmt.Sprintf("%d-%s-%s-%d-%d", block.Id, prevHashStr, merkleStr,
		block.Timestamp, block.Nonce)

	hash := sha256.Sum256([]byte(record))

	block.Hash = hash[:]

	return nil
}

func (block *Block) IsValidHash() bool {
	hashStr := hex.EncodeToString(block.Hash)

	return strings.HasPrefix(hashStr, strings.Repeat("0", Difficulty))
}

// parseDBTransactions -> Convert DB transactions to blockchain transactions
func (block *Block) parseDBTransactions(dbTxs []database.DBTransactionSchema) error {
	block.Transactions = make([]Transaction, len(dbTxs))

	for idx, dbTx := range dbTxs {
		decodedSignature, err := hex.DecodeString(dbTx.Signature)
		if err != nil {
			return err
		}

		block.Transactions[idx] = Transaction{
			From:       dbTx.From,
			To:         dbTx.To,
			Amount:     dbTx.Amount,
			Fee:        dbTx.Fee,
			Timestamp:  dbTx.Timestamp,
			PublicKey:  dbTx.PublicKey,
			Signature:  decodedSignature,
			Status:     dbTx.Status,
			IsCoinbase: dbTx.IsCoinbase,
		}
	}

	return nil
}

// SerializeTransactions -> Deterministic
func (block *Block) SerializeTransactions() string {
	txs := block.Transactions

	parts := make([]string, len(txs))

	for i, tx := range txs {
		if tx.IsCoinbase {
			// For coinbase, include nanosecond precision in serialization
			parts[i] = fmt.Sprintf("%s-%s-%d-%d-%t",
				tx.From, tx.To, tx.Amount, tx.Timestamp, tx.IsCoinbase)
		} else {
			parts[i] = fmt.Sprintf("%s-%s-%d-%x",
				tx.From, tx.To, tx.Amount, tx.Signature)
		}
	}

	return strings.Join(parts, "|")
}

func (block *Block) ComputeMerkleRoot() {
	// empty tree -> Hash of empty byte slice
	if len(block.Transactions) == 0 {
		h := sha256.Sum256([]byte{})
		block.MerkleRoot = h[:]
		return
	}

	// build leaves: sha256 of each transaction serialization
	leaves := make([][]byte, len(block.Transactions))
	for i, tx := range block.Transactions {
		var s string
		if tx.IsCoinbase {
			// For coinbase transactions, include timestamp for uniqueness
			s = fmt.Sprintf("%s-%s-%d-%d-%t",
				tx.From, tx.To, tx.Amount, tx.Timestamp, tx.IsCoinbase)
		} else {
			// For regular transactions, use signature for uniqueness
			s = fmt.Sprintf("%s-%s-%d-%x-%d-%t",
				tx.From, tx.To, tx.Amount, tx.Signature, tx.Timestamp, tx.IsCoinbase)
		}
		h := sha256.Sum256([]byte(s))
		leaf := make([]byte, sha256.Size)
		copy(leaf, h[:])
		leaves[i] = leaf
	}

	// build merkle tree (pairwise hashing, duplicate last if odd)
	for len(leaves) > 1 {
		if len(leaves)%2 == 1 {
			dup := make([]byte, len(leaves[len(leaves)-1]))
			copy(dup, leaves[len(leaves)-1])
			leaves = append(leaves, dup)
		}

		next := make([][]byte, 0, len(leaves)/2)
		for i := 0; i < len(leaves); i += 2 {
			concat := append(leaves[i], leaves[i+1]...)
			h := sha256.Sum256(concat)
			n := make([]byte, sha256.Size)
			copy(n, h[:])
			next = append(next, n)
		}
		leaves = next
	}

	block.MerkleRoot = leaves[0]
}

func (block *Block) GetHeader() *BlockHeader {
	return &BlockHeader{
		Id:         block.Id,
		PrevHash:   block.PrevHash,
		Hash:       block.Hash,
		MerkleRoot: block.MerkleRoot,
		Timestamp:  block.Timestamp,
		Nonce:      block.Nonce,
	}
}

func createGenesisBlock() (*Block, error) {
	block := &Block{
		Id:           0,
		PrevHash:     make([]byte, 32),
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		Nonce:        0,
	}

	block.ComputeMerkleRoot()

	if err := block.HashBlock(); err != nil {
		return nil, err
	}

	return block, nil
}

func (header *BlockHeader) Verify(idx int64, prevHeaderHash []byte) (bool, error) {
	if !header.verifyHash() {
		return false, errors.New("verifyHash error")
	}

	if idx != header.Id {
		return false, errors.New("index does not match the header id")
	}

	if header.Id == 0 {
		// Skip additional checks for genesis
		return true, nil
	}

	if !header.isDifficultyHashValid() {
		return false, errors.New("isDifficultyHashValid error")
	}

	if !bytes.Equal(header.PrevHash, prevHeaderHash) {
		return false, fmt.Errorf("chain continuity broken: block %d prevHash doesn't match", header.Id)
	}

	if header.Timestamp <= 0 {
		return false, errors.New("header timestamp is lower than 0")
	}

	if len(header.MerkleRoot) != 32 {
		return false, errors.New("invalid merkle root length")
	}

	return true, nil
}

func (header *BlockHeader) isDifficultyHashValid() bool {
	hashStr := hex.EncodeToString(header.Hash)
	return strings.HasPrefix(hashStr, strings.Repeat("0", Difficulty))
}

func (header *BlockHeader) verifyHash() bool {
	tempHeader := &BlockHeader{
		Id:         header.Id,
		PrevHash:   header.PrevHash,
		MerkleRoot: header.MerkleRoot,
		Timestamp:  header.Timestamp,
		Nonce:      header.Nonce,
		Hash:       nil, // Will be calculated
	}

	computedHash := tempHeader.computeHeaderHash()

	return bytes.Equal(computedHash, header.Hash)
}

func (header *BlockHeader) computeHeaderHash() []byte {
	prevHashStr := hex.EncodeToString(header.PrevHash)
	merkleStr := hex.EncodeToString(header.MerkleRoot)

	record := fmt.Sprintf("%d-%s-%s-%d-%d", header.Id, prevHashStr, merkleStr,
		header.Timestamp, header.Nonce)

	hash := sha256.Sum256([]byte(record))
	return hash[:]
}
