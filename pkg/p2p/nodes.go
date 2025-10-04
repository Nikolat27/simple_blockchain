package p2p

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/p2p/resolver"
	"simple_blockchain/pkg/p2p/types"
	"sync"
	"time"
)

type Node struct {
	TcpAddress string
	Peers      []string // Slice of TCP addresses

	Blockchain *blockchain.Blockchain `json:"blockchain"`
	payloadCh  chan types.Payload     // Communication channel

	mutex sync.RWMutex
}

// SetupNode -> Node
func SetupNode(address string, bc *blockchain.Blockchain) (*Node, error) {
	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	log.Printf("Node started to listen on: %s address\n", tcpListener.Addr().String())

	node := &Node{
		TcpAddress: address,
		Peers:      make([]string, 0),

		Blockchain: bc,

		payloadCh: make(chan types.Payload),
	}

	go node.startListening(tcpListener)

	return node, nil
}

// Bootstrap -> Discovers node using DNS resolving
func (node *Node) Bootstrap() {
	seedPeers := resolver.ResolveSeedNodes()

	for _, peer := range seedPeers {
		if peer == node.GetCurrentTcpAddress() {
			continue
		}

		go func(peerAddr string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := node.ConnectAndSync(ctx, peerAddr); err != nil {
				log.Println("Failed to connect: ", err)
			}
		}(peer)
	}
}

func (node *Node) ConnectAndSync(ctx context.Context, peerAddress string) error {
	msg := types.NewMessage(types.RequestHeadersMsg, node.GetCurrentTcpAddress(), types.Payload{})
	if err := node.WriteMessage(ctx, peerAddress, msg.Marshal()); err != nil {
		return err
	}

	// Receive headers
	select {
	case <-ctx.Done():
		return ctx.Err()
	case payload := <-node.payloadCh:
		var headers []blockchain.BlockHeader

		if err := payload.Unmarshal(&headers); err != nil {
			return err
		}

		if valid, err := node.Blockchain.VerifyHeaders(headers); err != nil {
			return err
		} else if !valid {
			return fmt.Errorf("peer %s send corrupted blockchain", peerAddress)
		}

		if err := node.DownloadMissingBlocks(ctx, peerAddress, headers); err != nil {
			return err
		}

		fmt.Printf("Successfully synced %d blocks with peer: %s\n", len(headers), peerAddress)
		return node.AddNewPeer(peerAddress)
	case <-time.After(60 * time.Second):
		return errors.New("ERROR get payload Timeout exceeded")
	}
}

func (node *Node) DownloadMissingBlocks(ctx context.Context, peerAddress string,
	headers []blockchain.BlockHeader) error {

	localBlocks, err := node.Blockchain.GetAllBlocks()
	if err != nil {
		return err
	}

	localBlockIds := make(map[int64]bool)
	for _, block := range localBlocks {
		localBlockIds[block.Id] = true
	}

	downloaded := 0
	for _, header := range headers {
		if localBlockIds[header.Id] {
			continue
		}

		if err := node.downloadBlock(ctx, peerAddress, header.Id); err != nil {
			return fmt.Errorf("failed to download block %d: %w", header.Id, err)
		}
		downloaded++
	}

	return nil
}

func (node *Node) downloadBlock(ctx context.Context, peerAddress string, blockId int64) error {
	payload, err := json.Marshal(blockId)
	if err != nil {
		return err
	}

	msg := types.NewMessage(types.RequestBlockMsg, node.GetCurrentTcpAddress(), payload)
	if err := node.WriteMessage(ctx, peerAddress, msg.Marshal()); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case payload := <-node.payloadCh:
		var block blockchain.Block
		if err := payload.Unmarshal(&block); err != nil {
			return err
		}

		if block.Id != blockId {
			return fmt.Errorf("received block ID mismatch: expected %d, got %d", blockId, block.Id)
		}

		sqlTx, err := node.Blockchain.Database.BeginTx()
		if err != nil {
			return err
		}

		defer sqlTx.Rollback()

		if err := node.Blockchain.AddBlock(sqlTx, &block); err != nil {
			return err
		}

		if err := sqlTx.Commit(); err != nil {
			return err
		}

		node.Blockchain.AddBlockToMemory(&block)

		return nil
		case <-time.After(60 * time.Second):
			return errors.New("ERROR get block payload timeout exceeded")
	}
}

func (node *Node) startListening(tcpListener net.Listener) error {
	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			return err
		}

		go node.handleListening(conn)
	}
}

// handleListening -> Node must consistently listen (read) to the TCP connection
func (node *Node) handleListening(conn net.Conn) {
	defer conn.Close()

	data, err := io.ReadAll(conn)
	if err != nil {
		log.Println(err)
	}

	if err := node.parseMessage(data); err != nil {
		log.Println("parsing message: ", err)
	}
}

// WriteMessage -> Writes to TCP connection
func (node *Node) WriteMessage(ctx context.Context, peerAddress string, msg []byte) error {
	var dialer net.Dialer

	conn, err := dialer.DialContext(ctx, "tcp", peerAddress)
	if err != nil {
		return fmt.Errorf("failed to dial peer %s: %w", peerAddress, err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return err
		}
	}

	if _, err := conn.Write(msg); err != nil {
		return err
	}

	return nil
}

func (node *Node) BroadcastBlock(block *blockchain.Block) error {
	payload, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	newMessage := types.NewMessage(types.BlockBroadcastMsg, node.GetCurrentTcpAddress(), payload)

	node.sendToAllPeers(newMessage.Marshal())

	return nil
}

func (node *Node) BroadcastMempool(mempool *blockchain.Mempool) error {
	payload, err := json.Marshal(mempool)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	newMessage := types.NewMessage(types.MempoolBroadcastMsg, node.GetCurrentTcpAddress(), payload)

	node.sendToAllPeers(newMessage.Marshal())

	return nil
}

func (node *Node) sendToAllPeers(newMessage []byte) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	fmt.Println("available peers: ", node.Peers)

	for _, peerAddr := range node.Peers {
		if peerAddr == node.GetCurrentTcpAddress() {
			continue
		}

		go func(addr string) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := node.WriteMessage(ctx, addr, newMessage); err != nil {
				log.Printf("Failed to broadcast block to %s: %v", addr, err)
			}
		}(peerAddr)
	}
}

func (node *Node) GetCurrentTcpAddress() string {
	return "127.0.0.1" + node.TcpAddress
}
