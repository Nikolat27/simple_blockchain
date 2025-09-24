package p2p

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/storage"
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

func NewNode(address string, blockchain *blockchain.Blockchain, mempool *blockchain.Mempool, storage *storage.Storage) *Node {
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

func (n *Node) AddPeer(address string) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	n.Peers[address] = &Peer{
		Address:   address,
		LastSeen:  time.Now(),
		Connected: false,
	}
	n.savePeersToStorage()
}

func (n *Node) savePeersToStorage() {
	if n.Storage == nil {
		return
	}

	n.peerMutex.RLock()
	peerData := make(map[string]*storage.PeerData)
	for addr, peer := range n.Peers {
		peerData[addr] = &storage.PeerData{
			Address:   peer.Address,
			LastSeen:  peer.LastSeen,
			Connected: peer.Connected,
		}
	}
	n.peerMutex.RUnlock()

	if err := n.Storage.SavePeers(peerData); err != nil {
		fmt.Printf("Error saving peers: %v\n", err)
	}
}

func (n *Node) LoadPeersFromStorage() {
	if n.Storage == nil {
		return
	}

	peerData, err := n.Storage.LoadPeers()
	if err != nil {
		fmt.Printf("Error loading peers: %v\n", err)
		return
	}

	n.peerMutex.Lock()
	for addr, data := range peerData {
		n.Peers[addr] = &Peer{
			Address:   data.Address,
			LastSeen:  data.LastSeen,
			Connected: data.Connected,
		}
	}
	n.peerMutex.Unlock()
}

func (n *Node) ConnectToPeer(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	n.peerMutex.Lock()
	if peer, exists := n.Peers[address]; exists {
		peer.Connected = true
		peer.LastSeen = time.Now()
		peer.conn = conn
	}
	n.peerMutex.Unlock()
	n.savePeersToStorage()

	go n.handleConnection(conn)

	return nil
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			fmt.Printf("Error parsing message: %v\n", err)
			continue
		}

		// Register/update inbound peer based on message source
		if msg.From != "" {
			n.peerMutex.Lock()
			if peer, exists := n.Peers[msg.From]; exists {
				peer.Connected = true
				peer.LastSeen = time.Now()
				peer.conn = conn
			} else {
				n.Peers[msg.From] = &Peer{
					Address:   msg.From,
					LastSeen:  time.Now(),
					Connected: true,
					conn:      conn,
				}
			}
			n.peerMutex.Unlock()
			n.savePeersToStorage()
		}

		select {
		case n.MessageCh <- &msg:
		default:
			fmt.Println("Message channel full, dropping message")
		}
	}
}

func (n *Node) SendMessage(address string, msg *Message) error {
	n.peerMutex.RLock()

	peer, exists := n.Peers[address]
	n.peerMutex.RUnlock()

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

func (n *Node) BroadcastMessage(msg *Message) {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	for _, peer := range n.Peers {
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

func (n *Node) StartMessageProcessor() {
	go func() {
		for msg := range n.MessageCh {
			n.processMessage(msg)
		}
	}()
}

func (n *Node) processMessage(msg *Message) {
	switch msg.Type {
	case MsgTransaction:
		n.handleTransactionMessage(msg)
	case MsgBlock:
		n.handleBlockMessage(msg)
	case MsgPeers:
		n.handlePeersMessage(msg)
	case MsgBlockchainReq:
		n.respondToBlockchainRequest(msg)
	case MsgBlockchainRsp:
		n.handleBlockchainResponse(msg)
	case MsgBlockReq:
		n.respondToBlockRequest(msg)
	case MsgBlockRsp:
		n.handleBlockResponse(msg)
	default:
		fmt.Printf("Unknown message type: %s\n", msg.Type)
	}
}

func (n *Node) handleTransactionMessage(msg *Message) {
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

	if tx.Verify() && n.Blockchain.ValidateTransaction(tx) {
		n.Mempool.AddTransaction(tx)
		fmt.Printf("Added transaction from p2p: %s -> %s (%d)\n", from, to, amount)
	} else {
		fmt.Println("Invalid transaction received via p2p")
	}
}

func (n *Node) handleBlockMessage(msg *Message) {
	blockMap, ok := msg.Payload.(map[string]any)
	if !ok {
		fmt.Println("Invalid block payload")
		return
	}

	block := n.reconstructBlockFromMap(blockMap)
	if block == nil {
		fmt.Println("Invalid block structure")
		return
	}

	if !n.Blockchain.VerifyBlock(block) {
		fmt.Println("Received invalid block from peer")
		return
	}

	blocks := n.Blockchain.GetBlocks()
	if len(blocks) == 0 {
		fmt.Println("No blocks in chain to compare")
		return
	}

	latestBlock := blocks[len(blocks)-1]
	if hex.EncodeToString(block.PrevHash) != hex.EncodeToString(latestBlock.Hash) {
		fmt.Printf("Block prevHash doesn't match latest block. Expected: %x, Got: %x\n",
			latestBlock.Hash, block.PrevHash)
		return
	}

	if block.Index != latestBlock.Index+1 {
		fmt.Printf("Block index mismatch. Expected: %d, Got: %d\n",
			latestBlock.Index+1, block.Index)
		return
	}

	n.Blockchain.AddBlock(block)
	n.Mempool.Clear()

	fmt.Printf("✅ Added block %d from peer %s (hash: %x)\n",
		block.Index, msg.From, block.Hash[:8])
	n.broadcastBlockToPeersExcept(block, msg.From)
}

func (n *Node) broadcastBlockToPeersExcept(block *blockchain.Block, excludeAddr string) {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	msg := &Message{
		Type:      MsgBlock,
		Payload:   block,
		From:      n.Address,
		Timestamp: time.Now().Unix(),
	}

	for addr, peer := range n.Peers {
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

func (n *Node) SyncBlockchain() {
	n.peerMutex.RLock()
	peerCount := len(n.Peers)
	n.peerMutex.RUnlock()

	if peerCount == 0 {
		fmt.Println("No peers available for blockchain sync")
		return
	}

	fmt.Println("🔄 Starting blockchain synchronization...")
	n.requestBlockchainInfo()
	time.Sleep(2 * time.Second)

	localHeight := len(n.Blockchain.GetBlocks()) - 1
	fmt.Printf("📊 Local blockchain height: %d\n", localHeight)
}

func (n *Node) requestBlockchainInfo() {
	msg := &Message{
		Type:      MsgBlockchainReq,
		Payload:   "status",
		From:      n.Address,
		Timestamp: time.Now().Unix(),
	}

	n.BroadcastMessage(msg)
}

func (n *Node) respondToBlockchainRequest(msg *Message) {
	blocks := n.Blockchain.GetBlocks()
	latestBlock := blocks[len(blocks)-1]

	info := BlockchainSyncInfo{
		LatestBlockHeight: latestBlock.Index,
		LatestBlockHash:   hex.EncodeToString(latestBlock.Hash),
	}

	responseMsg := &Message{
		Type:      MsgBlockchainRsp,
		Payload:   info,
		From:      n.Address,
		Timestamp: time.Now().Unix(),
	}

	n.SendMessage(msg.From, responseMsg)
}

func (n *Node) handleBlockchainResponse(msg *Message) {
	info, ok := msg.Payload.(map[string]any)
	if !ok {
		fmt.Println("Invalid blockchain response format")
		return
	}

	peerHeight := int(info["latest_block_height"].(float64))
	localHeight := len(n.Blockchain.GetBlocks()) - 1

	fmt.Printf("📡 Peer %s has blockchain height: %d (local: %d)\n",
		msg.From, peerHeight, localHeight)

	if peerHeight > localHeight {
		fmt.Printf("🔄 Local chain behind by %d blocks, requesting sync from %s\n",
			peerHeight-localHeight, msg.From)
		// Request missing blocks
		n.requestMissingBlocks(msg.From, localHeight+1, peerHeight)
	} else if peerHeight < localHeight {
		fmt.Printf("📤 Local chain ahead, peer %s is behind\n", msg.From)
	} else {
		fmt.Printf("✅ Local chain in sync with peer %s\n", msg.From)
	}
}

func (n *Node) requestMissingBlocks(peerAddr string, startIndex, endIndex int) {
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
			From:      n.Address,
			Timestamp: time.Now().Unix(),
		}

		n.SendMessage(peerAddr, msg)

		// Small delay between batches
		time.Sleep(100 * time.Millisecond)
	}
}

func (n *Node) respondToBlockRequest(msg *Message) {
	indices, ok := msg.Payload.([]any)
	if !ok {
		fmt.Println("Invalid block request format")
		return
	}

	blocks := n.Blockchain.GetBlocks()
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
			From:      n.Address,
			Timestamp: time.Now().Unix(),
		}

		n.SendMessage(msg.From, responseMsg)
		fmt.Printf("📤 Sent %d blocks to %s\n", len(responseBlocks), msg.From)
	}
}

func (n *Node) handleBlockResponse(msg *Message) {
	blocks, ok := msg.Payload.([]any)
	if !ok {
		fmt.Println("Invalid block response format")
		return
	}

	localBlocks := n.Blockchain.GetBlocks()
	localHeight := len(localBlocks) - 1

	addedBlocks := 0
	for _, blockData := range blocks {
		blockMap, ok := blockData.(map[string]any)
		if !ok {
			continue
		}

		block := n.reconstructBlockFromMap(blockMap)
		if block == nil {
			continue
		}

		if block.Index == localHeight+1 && hex.EncodeToString(block.PrevHash) == hex.EncodeToString(localBlocks[localHeight].Hash) {
			if n.Blockchain.VerifyBlock(block) {
				n.Blockchain.AddBlock(block)
				localHeight++
				addedBlocks++
				n.Mempool.Clear()
				fmt.Printf("✅ Added synced block %d from %s\n", block.Index, msg.From)
			}
		}
	}

	if addedBlocks > 0 {
		fmt.Printf("🔄 Successfully synced %d blocks from %s\n", addedBlocks, msg.From)
	}
}

func (n *Node) reconstructBlockFromMap(blockMap map[string]any) *blockchain.Block {
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

func (n *Node) handlePeersMessage(msg *Message) {
	peersData, ok := msg.Payload.([]any)
	if !ok {
		fmt.Println("Invalid peers payload")
		return
	}

	for _, peerAddr := range peersData {
		if addr, ok := peerAddr.(string); ok && addr != n.Address {
			n.AddPeer(addr)
		}
	}
}

func (n *Node) BroadcastTransaction(tx *blockchain.Transaction) {
	msg := &Message{
		Type:      MsgTransaction,
		Payload:   tx,
		From:      n.Address,
		Timestamp: time.Now().Unix(),
	}

	n.BroadcastMessage(msg)
}

func (n *Node) BroadcastBlock(block *blockchain.Block) {
	msg := &Message{
		Type:      MsgBlock,
		Payload:   block,
		From:      n.Address,
		Timestamp: time.Now().Unix(),
	}

	n.BroadcastMessage(msg)
}
