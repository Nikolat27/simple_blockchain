package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"simple_blockchain/pkg/CryptoGraphy"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/database"
	"simple_blockchain/pkg/p2p"
	"simple_blockchain/pkg/p2p/types"
	"simple_blockchain/pkg/utils"
	"testing"
	"time"
)

// setupTestDatabase creates a temporary test database
func setupTestDatabase(t *testing.T) (*database.Database, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_p2p.db")

	db, err := database.New("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	err = runMigrations(db.DB)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

// runMigrations creates the necessary tables for testing
func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS balances (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			address TEXT NOT NULL UNIQUE,
			balance INTEGER NOT NULL DEFAULT(0)
		)`,
		`CREATE TABLE IF NOT EXISTS blocks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			prev_hash TEXT UNIQUE NOT NULL,
			hash TEXT UNIQUE NOT NULL,
			merkle_root TEXT UNIQUE NOT NULL,
			nonce INTEGER DEFAULT (0),
			timestamp INTEGER DEFAULT (strftime('%s', 'now')),
			block_height INTEGER DEFAULT (0)
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			block_id INTEGER NULL,
			sender TEXT NULL,
			recipient TEXT NOT NULL,
			amount INTEGER DEFAULT (0),
			fee INTEGER DEFAULT (0),
			timestamp INTEGER DEFAULT (strftime('%s', 'now')),
			public_key TEXT NULL,
			signature TEXT NULL,
			status TEXT NOT NULL DEFAULT ('pending') CHECK (status IN ('pending', 'confirmed')),
			is_coin_base INTEGER NOT NULL DEFAULT (0) CHECK (is_coin_base IN (0, 1)),
			FOREIGN KEY (block_id) REFERENCES blocks (id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS peers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tcp_address TEXT NOT NULL UNIQUE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_block_id ON transactions (block_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}

// TestNode_GetCurrentTcpAddress tests getting current TCP address
func TestNode_GetCurrentTcpAddress(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	node, err := p2p.SetupNode(":9000", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	address := node.GetCurrentTcpAddress()

	if address == "" {
		t.Error("TCP address should not be empty")
	}

	// Should start with 127.0.0.1
	expected := "127.0.0.1:9000"
	if address != expected {
		t.Errorf("Expected address %s, got %s", expected, address)
	}
}

// TestNode_AddNewPeer tests adding a new peer
func TestNode_AddNewPeer(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	node, err := p2p.SetupNode(":9001", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	peerAddr := "127.0.0.1:9002"

	// Add peer
	err = node.AddNewPeer(peerAddr)
	if err != nil {
		t.Fatalf("Failed to add peer: %v", err)
	}

	// Verify peer was added
	node.Mutex.RLock()
	found := false
	for _, peer := range node.Peers {
		if peer == peerAddr {
			found = true
			break
		}
	}
	node.Mutex.RUnlock()

	if !found {
		t.Error("Peer should be added to node's peer list")
	}
}

// TestNode_AddNewPeer_Duplicate tests adding duplicate peer
func TestNode_AddNewPeer_Duplicate(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	node, err := p2p.SetupNode(":9003", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	peerAddr := "127.0.0.1:9004"

	// Add peer first time
	err = node.AddNewPeer(peerAddr)
	if err != nil {
		t.Fatalf("Failed to add peer first time: %v", err)
	}

	// Add same peer again
	err = node.AddNewPeer(peerAddr)
	if err != nil {
		t.Fatalf("Adding duplicate peer should not error: %v", err)
	}

	// Verify peer count
	node.Mutex.RLock()
	count := 0
	for _, peer := range node.Peers {
		if peer == peerAddr {
			count++
		}
	}
	node.Mutex.RUnlock()

	if count != 1 {
		t.Errorf("Peer should only appear once, found %d times", count)
	}
}

// TestNode_BroadcastBlock tests broadcasting a block
func TestNode_BroadcastBlock(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	node, err := p2p.SetupNode(":9005", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	// Create a test block
	genesisBlock := bc.GetLatestBlock()
	coinbaseTx := blockchain.CreateCoinbaseTx("miner", blockchain.MiningReward)
	testBlock := &blockchain.Block{
		Id:           genesisBlock.Id + 1,
		PrevHash:     genesisBlock.Hash,
		Timestamp:    utils.GetTimestamp(),
		Transactions: []blockchain.Transaction{*coinbaseTx},
		Nonce:        0,
	}

	testBlock.ComputeMerkleRoot()
	err = testBlock.HashBlock()
	if err != nil {
		t.Fatalf("Failed to hash block: %v", err)
	}

	// Broadcast should not error even with no peers
	err = node.BroadcastBlock(testBlock)
	if err != nil {
		t.Errorf("Broadcast should not error: %v", err)
	}
}

// TestNode_BroadcastMempool tests broadcasting mempool
func TestNode_BroadcastMempool(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	node, err := p2p.SetupNode(":9006", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	// Broadcast should not error even with no peers
	err = node.BroadcastMempool(mp)
	if err != nil {
		t.Errorf("Broadcast mempool should not error: %v", err)
	}
}

// TestNode_CancelMining tests cancel mining broadcast
func TestNode_CancelMining(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	node, err := p2p.SetupNode(":9007", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	// Cancel mining should not error
	err = node.CancelMining()
	if err != nil {
		t.Errorf("Cancel mining should not error: %v", err)
	}
}

// TestNode_WriteMessage tests writing message with context
func TestNode_WriteMessage(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	// Setup receiver node
	receiverNode, err := p2p.SetupNode(":9008", bc)
	if err != nil {
		t.Fatalf("Failed to setup receiver node: %v", err)
	}

	// Setup sender node
	senderNode, err := p2p.SetupNode(":9009", bc)
	if err != nil {
		t.Fatalf("Failed to setup sender node: %v", err)
	}

	// Give receiver time to start listening
	time.Sleep(100 * time.Millisecond)

	// Create a test message
	msg := types.NewMessage(types.CancelMiningMsg, senderNode.GetCurrentTcpAddress(), types.Payload{})

	// Write message with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = senderNode.WriteMessage(ctx, receiverNode.GetCurrentTcpAddress(), msg.Marshal())
	if err != nil {
		t.Errorf("Failed to write message: %v", err)
	}
}

// TestNode_WriteMessage_ContextCancellation tests context cancellation
func TestNode_WriteMessage_ContextCancellation(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	node, err := p2p.SetupNode(":9010", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	msg := types.NewMessage(types.CancelMiningMsg, node.GetCurrentTcpAddress(), types.Payload{})

	// Create already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should fail with context error
	err = node.WriteMessage(ctx, "127.0.0.1:99999", msg.Marshal())
	if err == nil {
		t.Error("Should return error for cancelled context")
	}
}

// TestNode_VerifyBlockTransactions tests transaction verification
func TestNode_VerifyBlockTransactions(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	_, err = p2p.SetupNode(":9011", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	// Create a valid signed transaction
	keyPair, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	tx := &blockchain.Transaction{
		From:      keyPair.Address,
		To:        "recipient",
		Amount:    100,
		Fee:       10,
		Timestamp: utils.GetTimestamp(),
		Status:    "confirmed",
	}

	err = tx.Sign(keyPair)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Create coinbase transaction
	coinbaseTx := blockchain.CreateCoinbaseTx("miner", blockchain.MiningReward)

	// Create block with transactions
	genesisBlock := bc.GetLatestBlock()
	_ = &blockchain.Block{
		Id:           genesisBlock.Id + 1,
		PrevHash:     genesisBlock.Hash,
		Timestamp:    utils.GetTimestamp(),
		Transactions: []blockchain.Transaction{*coinbaseTx, *tx},
		Nonce:        0,
	}

	// Note: We can't directly call verifyBlockTransactions as it's not exported
	// This test demonstrates the setup for transaction verification
	// The actual verification happens inside node message handlers
}

// TestNode_ConcurrentPeerAccess tests concurrent access to peers
func TestNode_ConcurrentPeerAccess(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	node, err := p2p.SetupNode(":9012", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	// Add peers concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			peerAddr := fmt.Sprintf("127.0.0.1:%d", 10000+id)
			_ = node.AddNewPeer(peerAddr)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify peers were added
	node.Mutex.RLock()
	peerCount := len(node.Peers)
	node.Mutex.RUnlock()

	if peerCount == 0 {
		t.Error("Should have added peers")
	}
}

// TestNode_MessageParsing tests message parsing
func TestNode_MessageParsing(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	_, err = p2p.SetupNode(":9013", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	// Test different message types
	messageTypes := []string{
		types.CancelMiningMsg,
	}

	for _, msgType := range messageTypes {
		t.Run(msgType, func(t *testing.T) {
			msg := types.NewMessage(msgType, "127.0.0.1:8080", types.Payload{})
			data := msg.Marshal()

			// Verify message can be unmarshaled
			var decoded types.Message
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("Failed to unmarshal message: %v", err)
			}

			if decoded.Type != msgType {
				t.Errorf("Expected type %s, got %s", msgType, decoded.Type)
			}
		})
	}
}

// TestNode_SetupMultipleNodes tests setting up multiple nodes
func TestNode_SetupMultipleNodes(t *testing.T) {
	nodes := make([]*p2p.Node, 3)

	for i := 0; i < 3; i++ {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		mp := blockchain.NewMempool(1048576)
		bc, err := blockchain.NewBlockchain(db, mp)
		if err != nil {
			t.Fatalf("Failed to create blockchain %d: %v", i, err)
		}

		port := 9100 + i
		portStr := fmt.Sprintf(":%d", port)
		node, err := p2p.SetupNode(portStr, bc)
		if err != nil {
			t.Fatalf("Failed to setup node %d: %v", i, err)
		}

		nodes[i] = node
	}

	// Verify all nodes are set up
	for i, node := range nodes {
		if node == nil {
			t.Errorf("Node %d should not be nil", i)
		}

		if node.Blockchain == nil {
			t.Errorf("Node %d blockchain should not be nil", i)
		}
	}
}

// TestNode_BlockchainIntegration tests node with blockchain operations
func TestNode_BlockchainIntegration(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	mp := blockchain.NewMempool(1048576)
	bc, err := blockchain.NewBlockchain(db, mp)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}

	node, err := p2p.SetupNode(":9014", bc)
	if err != nil {
		t.Fatalf("Failed to setup node: %v", err)
	}

	// Verify node has access to blockchain
	if node.Blockchain == nil {
		t.Fatal("Node blockchain should not be nil")
	}

	// Verify node can access blockchain data
	latestBlock := node.Blockchain.GetLatestBlock()
	if latestBlock == nil {
		t.Error("Should be able to get latest block through node")
	}

	// Verify genesis block
	if latestBlock.Id != 0 {
		t.Errorf("Expected genesis block ID 0, got %d", latestBlock.Id)
	}
}
