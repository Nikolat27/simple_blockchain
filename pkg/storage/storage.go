package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"simple_blockchain/pkg/blockchain"
	"time"
)

// Storage handles data persistency using JSON files
type Storage struct {
	DataDir string
}

func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	return &Storage{
		DataDir: dataDir,
	}, nil
}

// SaveBlockchain -> Save blockchain to JSON file
func (s *Storage) SaveBlockchain(bc *blockchain.Blockchain) error {
	// Take a snapshot to avoid races while marshalling
	blocks := bc.GetBlocks()
	snapshot := &blockchain.Blockchain{Blocks: blocks}

	data, err := json.MarshalIndent(snapshot, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal blockchain: %w", err)
	}

	filePath := filepath.Join(s.DataDir, "blockchain.json")
	return os.WriteFile(filePath, data, 0644)
}

// LoadBlockchain -> Load blockchain from JSON file
func (s *Storage) LoadBlockchain() (*blockchain.Blockchain, error) {
	filePath := filepath.Join(s.DataDir, "blockchain.json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("blockchain file does not exist")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read blockchain file: %w", err)
	}

	var bc blockchain.Blockchain
	if err := json.Unmarshal(data, &bc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blockchain: %w", err)
	}

	return &bc, nil
}

// SaveMempool -> Save mempool transactions to JSON file
func (s *Storage) SaveMempool(mp *blockchain.Mempool) error {
	transactions := mp.GetTransactions()

	data, err := json.MarshalIndent(transactions, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal mempool: %w", err)
	}

	filePath := filepath.Join(s.DataDir, "mempool.json")
	return os.WriteFile(filePath, data, 0644)
}

// LoadMempool -> Load mempool transactions from JSON file
func (s *Storage) LoadMempool() ([]blockchain.Transaction, error) {
	filePath := filepath.Join(s.DataDir, "mempool.json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []blockchain.Transaction{}, nil // Empty mempool is fine
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mempool file: %w", err)
	}

	var transactions []blockchain.Transaction
	if err := json.Unmarshal(data, &transactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mempool: %w", err)
	}

	return transactions, nil
}

// PeerData represents peer information for JSON storage
type PeerData struct {
	Address   string    `json:"address"`
	LastSeen  time.Time `json:"last_seen"`
	Connected bool      `json:"connected"`
}

// SavePeers -> Save peers to JSON file
func (s *Storage) SavePeers(peers map[string]*PeerData) error {
	peerList := make([]PeerData, 0, len(peers))
	for _, peer := range peers {
		peerList = append(peerList, *peer)
	}

	data, err := json.MarshalIndent(peerList, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal peers: %w", err)
	}

	filePath := filepath.Join(s.DataDir, "peers.json")
	return os.WriteFile(filePath, data, 0644)
}

// LoadPeers -> Load peers from JSON file
func (s *Storage) LoadPeers() (map[string]*PeerData, error) {
	filePath := filepath.Join(s.DataDir, "peers.json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return make(map[string]*PeerData), nil // No saved peers is fine
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read peers file: %w", err)
	}

	var peerList []PeerData
	if err := json.Unmarshal(data, &peerList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal peers: %w", err)
	}

	peers := make(map[string]*PeerData)
	for _, peer := range peerList {
		peerCopy := peer // Create a copy
		peers[peer.Address] = &peerCopy
	}

	return peers, nil
}
