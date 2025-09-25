package p2p

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/storage"
	"strings"
	"sync"
	"time"
)

type Peer struct {
	Address   string
	LastSeen  time.Time
	Connected bool
	conn      net.Conn
}

type Node struct {
	ID         string
	Address    string
	Peers      map[string]*Peer
	Blockchain *blockchain.Blockchain
	Mempool    *blockchain.Mempool
	Storage    *storage.Storage
	MessageCh  chan *Message

	peerMutex sync.RWMutex
}

func NewNode(address string, blockchain *blockchain.Blockchain, mempool *blockchain.Mempool,
	storage *storage.Storage) *Node {

	return &Node{
		ID:         generateNodeID(),
		Address:    address,
		Peers:      make(map[string]*Peer),
		Blockchain: blockchain,
		Mempool:    mempool,
		Storage:    storage,
		MessageCh:  make(chan *Message, 100),
	}
}

func (node *Node) ConnectToBootstrapNodes(peersList []string, nodeAddr string) {
	for _, peerAddr := range peersList {
		peerAddr = strings.TrimSpace(peerAddr)
		if peerAddr == "" || peerAddr == nodeAddr {
			continue
		}

		node.AddPeer(peerAddr)
		if err := node.ConnectToPeer(peerAddr); err != nil {
			log.Printf("Failed to connect to peer %s: %v\n", peerAddr, err)
		}

	}
}

func (node *Node) AddPeer(address string) {
	node.peerMutex.Lock()
	defer node.peerMutex.Unlock()

	node.Peers[address] = &Peer{
		Address:   address,
		LastSeen:  time.Now(),
		Connected: false,
	}
	node.savePeersToStorage()
}

func (node *Node) savePeersToStorage() {
	if node.Storage == nil {
		return
	}

	// Note: This function assumes the caller already holds the peerMutex
	peerData := make(map[string]*storage.PeerData)
	for addr, peer := range node.Peers {
		peerData[addr] = &storage.PeerData{
			Address:   peer.Address,
			LastSeen:  peer.LastSeen,
			Connected: peer.Connected,
		}
	}

	node.Storage.SavePeers(peerData)
}

func (node *Node) LoadPeersFromStorage() {
	if node.Storage == nil {
		return
	}

	peerData, err := node.Storage.LoadPeers()
	if err != nil {
		fmt.Printf("Error loading peers: %v\n", err)
		return
	}

	node.peerMutex.Lock()
	for addr, data := range peerData {
		node.Peers[addr] = &Peer{
			Address:   data.Address,
			LastSeen:  data.LastSeen,
			Connected: data.Connected,
		}
	}
	node.peerMutex.Unlock()
}

func (node *Node) ConnectToPeer(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	node.peerMutex.Lock()
	if peer, exists := node.Peers[address]; exists {
		peer.Connected = true
		peer.LastSeen = time.Now()
		peer.conn = conn
	}
	node.peerMutex.Unlock()
	node.savePeersToStorage()

	go node.handleConnection(conn)

	return nil
}

func (node *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var connReceivedData = scanner.Bytes()

		var msg Message
		if err := parseMessage(connReceivedData, &msg); err != nil {
			fmt.Printf("Error parsing message: %v\n", err)
			continue
		}

		// Register/update inbound peer based on message source
		if msg.From != "" {
			node.peerMutex.Lock()
			if peer, exists := node.Peers[msg.From]; exists {
				peer.Connected = true
				peer.LastSeen = time.Now()
				peer.conn = conn
			} else {
				node.Peers[msg.From] = &Peer{
					Address:   msg.From,
					LastSeen:  time.Now(),
					Connected: true,
					conn:      conn,
				}
			}
			node.peerMutex.Unlock()
			node.savePeersToStorage()
		}

		select {
		case node.MessageCh <- &msg:
			// send the received msg to the node message channel
		default:
			fmt.Println("Message channel full, dropping message")
		}
	}
}

func parseMessage(receivedMsg []byte, msg *Message) error {
	return json.Unmarshal(receivedMsg, msg)
}

func (node *Node) SendMessage(address string, msg *Message) error {
	node.peerMutex.RLock()

	peer, exists := node.Peers[address]
	node.peerMutex.RUnlock()

	if !exists || !peer.Connected || peer.conn == nil {
		return fmt.Errorf("peer not connected: %s", address)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = peer.conn.Write(append(data, '\n'))
	return err
}

func (node *Node) BroadcastMessage(msg *Message) {
	node.peerMutex.RLock()
	defer node.peerMutex.RUnlock()

	for _, peer := range node.Peers {
		if peer.Connected && peer.conn != nil {
			data, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("Error marshaling message: %v\n", err)
				continue
			}

			peer.conn.Write(append(data, '\n'))
		}
	}
}

func generateNodeID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("node-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func (node *Node) StartMessageProcessor() {
	go func() {
		for msg := range node.MessageCh {
			node.processMessage(msg)
		}
	}()
}

func (node *Node) processMessage(msg *Message) {
	switch msg.Type {
	case MsgTransaction:
		node.handleTransactionMessage(msg)
	case MsgBlock:
		node.handleBlockMessage(msg)
	case MsgPeers:
		node.handlePeersMessage(msg)
	case MsgBlockchainReq:
		node.respondToBlockchainRequest(msg)
	case MsgBlockchainRsp:
		node.handleBlockchainResponse(msg)
	case MsgBlockReq:
		node.respondToBlockRequest(msg)
	case MsgBlockRsp:
		node.handleBlockResponse(msg)
	default:
		fmt.Printf("Unknown message type: %s\n", msg.Type)
	}
}

func (node *Node) handleTransactionMessage(msg *Message) {
	txData, ok := msg.Payload.(map[string]any)
	if !ok {
		fmt.Println("Invalid transaction payload")
		return
	}

	from := getStringFromMap(txData, "from")
	to := getStringFromMap(txData, "to")

	var amount uint64
	if a, ok := txData["amount"].(float64); ok {
		amount = uint64(a)
	}

	publicKey := getStringFromMap(txData, "PublicKey")
	sigB64 := getStringFromMap(txData, "Signature")
	var signature []byte
	if sigB64 != "" {
		if sig, err := base64.StdEncoding.DecodeString(sigB64); err == nil {
			signature = sig
		}
	}

	var timestamp int64
	if ts, ok := txData["Timestamp"].(float64); ok {
		timestamp = int64(ts)
	}

	tx := &blockchain.Transaction{
		From:       from,
		To:         to,
		Amount:     amount,
		Signature:  signature,
		PublicKey:  publicKey,
		Timestamp:  timestamp,
		IsCoinbase: false,
	}

	if tx.Verify() && node.Blockchain.ValidateTransaction(tx) {
		node.Mempool.AddTransaction(tx)
		fmt.Printf("Added transaction from p2p: %s -> %s (%d)\n", from, to, amount)
	} else {
		fmt.Println("Invalid transaction received via p2p")
	}
}

func (node *Node) handleBlockMessage(msg *Message) {
	blockMap, ok := msg.Payload.(map[string]any)
	if !ok {
		fmt.Println("Invalid block payload")
		return
	}

	block := node.reconstructBlockFromMap(blockMap)
	if block == nil {
		fmt.Println("Invalid block structure")
		return
	}

	if !node.Blockchain.VerifyBlock(block) {
		fmt.Println("Received invalid block from peer")
		return
	}

	// Use AcceptBlock which implements longest chain rule
	if node.Blockchain.AcceptBlock(block) {
		fmt.Printf("✅ Accepted block %d from peer %s (hash: %x)\n",
			block.Index, msg.From, block.Hash[:8])

		// Clear mempool if we added to main chain (simplified - in full implementation
		// we should check if any transactions became invalid)
		node.Mempool.Clear()

		// Stop any ongoing mining since we received a new block
		node.Blockchain.StopMining()

		node.broadcastBlockToPeersExcept(block, msg.From)
	} else {
		fmt.Printf("❌ Rejected block %d from peer %s\n", block.Index, msg.From)
		return
	}
}

func (node *Node) broadcastBlockToPeersExcept(block *blockchain.Block, excludeAddr string) {
	node.peerMutex.RLock()
	defer node.peerMutex.RUnlock()

	msg := &Message{
		Type:      MsgBlock,
		Payload:   block,
		From:      node.Address,
		Timestamp: time.Now().Unix(),
	}

	for addr, peer := range node.Peers {
		if addr != excludeAddr && peer.Connected && peer.conn != nil {
			data, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("Error marshaling block message: %v\n", err)
				continue
			}
			peer.conn.Write(append(data, '\n'))
		}
	}
}

// BlockchainSyncInfo represents blockchain synchronization information
type BlockchainSyncInfo struct {
	LatestBlockHeight int    `json:"latest_block_height"`
	LatestBlockHash   string `json:"latest_block_hash"`
}

func (node *Node) SyncBlockchain() {
	node.peerMutex.RLock()
	peerCount := len(node.Peers)
	node.peerMutex.RUnlock()

	if peerCount == 0 {
		fmt.Println("No peers available for blockchain sync")
		return
	}

	fmt.Println("🔄 Starting blockchain synchronization...")
	node.requestBlockchainInfo()
	time.Sleep(2 * time.Second)

	localHeight := len(node.Blockchain.GetBlocks()) - 1
	fmt.Printf("📊 Local blockchain height: %d\n", localHeight)
}

func (node *Node) requestBlockchainInfo() {
	msg := &Message{
		Type:      MsgBlockchainReq,
		Payload:   "status",
		From:      node.Address,
		Timestamp: time.Now().Unix(),
	}

	node.BroadcastMessage(msg)
}

func (node *Node) respondToBlockchainRequest(msg *Message) {
	blocks := node.Blockchain.GetBlocks()
	latestBlock := blocks[len(blocks)-1]

	info := BlockchainSyncInfo{
		LatestBlockHeight: latestBlock.Index,
		LatestBlockHash:   hex.EncodeToString(latestBlock.Hash),
	}

	responseMsg := &Message{
		Type:      MsgBlockchainRsp,
		Payload:   info,
		From:      node.Address,
		Timestamp: time.Now().Unix(),
	}

	node.SendMessage(msg.From, responseMsg)
}

func (node *Node) handleBlockchainResponse(msg *Message) {
	info, ok := msg.Payload.(map[string]any)
	if !ok {
		fmt.Println("Invalid blockchain response format")
		return
	}

	peerHeight := int(info["latest_block_height"].(float64))
	localHeight := len(node.Blockchain.GetBlocks()) - 1

	fmt.Printf("📡 Peer %s has blockchain height: %d (local: %d)\n",
		msg.From, peerHeight, localHeight)

	if peerHeight > localHeight {
		fmt.Printf("🔄 Local chain behind by %d blocks, requesting sync from %s\n",
			peerHeight-localHeight, msg.From)
		// Request missing blocks
		node.requestMissingBlocks(msg.From, localHeight+1, peerHeight)
	} else if peerHeight < localHeight {
		fmt.Printf("📤 Local chain ahead, peer %s is behind\n", msg.From)
	} else {
		fmt.Printf("✅ Local chain in sync with peer %s\n", msg.From)
	}
}

func (node *Node) requestMissingBlocks(peerAddr string, startIndex, endIndex int) {
	fmt.Printf("📥 Requesting blocks %d to %d from %s\n", startIndex, endIndex, peerAddr)
	batchSize := 10
	for i := startIndex; i <= endIndex; i += batchSize {
		end := i + batchSize - 1
		if end > endIndex {
			end = endIndex
		}

		blockIndices := make([]int, 0, batchSize)
		for j := i; j <= end; j++ {
			blockIndices = append(blockIndices, j)
		}

		msg := &Message{
			Type:      MsgBlockReq,
			Payload:   blockIndices,
			From:      node.Address,
			Timestamp: time.Now().Unix(),
		}

		node.SendMessage(peerAddr, msg)

		// Small delay between batches
		time.Sleep(100 * time.Millisecond)
	}
}

func (node *Node) respondToBlockRequest(msg *Message) {
	indices, ok := msg.Payload.([]any)
	if !ok {
		fmt.Println("Invalid block request format")
		return
	}

	blocks := node.Blockchain.GetBlocks()
	responseBlocks := make([]*blockchain.Block, 0)

	for _, idx := range indices {
		index := int(idx.(float64))
		if index >= 0 && index < len(blocks) {
			// Create a copy to avoid modifying original
			block := blocks[index]
			responseBlocks = append(responseBlocks, &block)
		}
	}

	if len(responseBlocks) > 0 {
		responseMsg := &Message{
			Type:      MsgBlockRsp,
			Payload:   responseBlocks,
			From:      node.Address,
			Timestamp: time.Now().Unix(),
		}

		node.SendMessage(msg.From, responseMsg)
		fmt.Printf("📤 Sent %d blocks to %s\n", len(responseBlocks), msg.From)
	}
}

func (node *Node) handleBlockResponse(msg *Message) {
	blocks, ok := msg.Payload.([]any)
	if !ok {
		fmt.Println("Invalid block response format")
		return
	}

	localBlocks := node.Blockchain.GetBlocks()
	localHeight := len(localBlocks) - 1

	addedBlocks := 0
	for _, blockData := range blocks {
		blockMap, ok := blockData.(map[string]any)
		if !ok {
			continue
		}

		block := node.reconstructBlockFromMap(blockMap)
		if block == nil {
			continue
		}

		if block.Index == localHeight+1 && hex.EncodeToString(block.PrevHash) == hex.EncodeToString(localBlocks[localHeight].Hash) {
			if node.Blockchain.VerifyBlock(block) {
				node.Blockchain.AddBlock(block)
				localHeight++
				addedBlocks++
				node.Mempool.Clear()
				fmt.Printf("✅ Added synced block %d from %s\n", block.Index, msg.From)
			}
		}
	}

	if addedBlocks > 0 {
		fmt.Printf("🔄 Successfully synced %d blocks from %s\n", addedBlocks, msg.From)
	}
}

func (node *Node) reconstructBlockFromMap(blockMap map[string]any) *blockchain.Block {
	index := int(blockMap["index"].(float64))
	prevHashStr := blockMap["prev_hash"].(string)
	hashStr := blockMap["hash"].(string)
	timestampStr := blockMap["timestamp"].(string)
	nonce := int(blockMap["nonce"].(float64))

	prevHash, _ := base64.StdEncoding.DecodeString(prevHashStr)
	hash, _ := base64.StdEncoding.DecodeString(hashStr)
	timestamp, _ := time.Parse(time.RFC3339Nano, timestampStr)

	transactions := make([]blockchain.Transaction, 0)
	if txs, ok := blockMap["transactions"].([]any); ok {
		for _, txData := range txs {
			if txMap, ok := txData.(map[string]any); ok {
				tx := blockchain.Transaction{
					From:       getStringFromMap(txMap, "from"),
					To:         getStringFromMap(txMap, "to"),
					Amount:     uint64(txMap["amount"].(float64)),
					Timestamp:  int64(txMap["Timestamp"].(float64)),
					PublicKey:  getStringFromMap(txMap, "PublicKey"),
					Signature:  []byte(getStringFromMap(txMap, "Signature")),
					Status:     getStringFromMap(txMap, "status"),
					IsCoinbase: txMap["is_coinbase"].(bool),
				}
				transactions = append(transactions, tx)
			}
		}
	}

	return &blockchain.Block{
		Index:        index,
		PrevHash:     prevHash,
		Hash:         hash,
		Timestamp:    timestamp,
		Nonce:        nonce,
		Transactions: transactions,
	}
}

func getStringFromMap(m map[string]any, key string) string {
	if val, ok := m[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (node *Node) handlePeersMessage(msg *Message) {
	peersData, ok := msg.Payload.([]any)
	if !ok {
		fmt.Println("Invalid peers payload")
		return
	}

	for _, peerAddr := range peersData {
		if addr, ok := peerAddr.(string); ok && addr != node.Address {
			node.AddPeer(addr)
		}
	}
}

func (node *Node) BroadcastTransaction(tx *blockchain.Transaction) {
	msg := &Message{
		Type:      MsgTransaction,
		Payload:   tx,
		From:      node.Address,
		Timestamp: time.Now().Unix(),
	}

	node.BroadcastMessage(msg)
}

func (node *Node) BroadcastBlock(block *blockchain.Block) {
	msg := &Message{
		Type:      MsgBlock,
		Payload:   block,
		From:      node.Address,
		Timestamp: time.Now().Unix(),
	}

	node.BroadcastMessage(msg)
}
