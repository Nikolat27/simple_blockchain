package p2p

import (
	"fmt"
	"io"
	"log"
	"net"
	"simple_blockchain/pkg/blockchain"
)

type Node struct {
	TcpAddress string
	Peers      []string // slice of TCP addresses

	Blockchain *blockchain.Blockchain `json:"blockchain"`

	BlockchainRespCh chan []blockchain.Block // communication channel
}

func NewNode(address string, bc *blockchain.Blockchain) (*Node, error) {
	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	log.Printf("Node started to listen on: %s address\n", tcpListener.Addr().String())

	node := &Node{
		TcpAddress: address,
		Peers:      make([]string, 0),

		Blockchain: bc,

		BlockchainRespCh: make(chan []blockchain.Block),
	}

	go node.startListening(tcpListener)

	return node, nil
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

	log.Printf("New connection from: %s", conn.RemoteAddr().String())

	data, err := io.ReadAll(conn)
	if err != nil {
		log.Println(err)
	}

	if err := node.parseMessage(data); err != nil {
		log.Println("ERROR parsing message: ", err)
	}
}

// Write -> Writes to TCP connection
func (node *Node) Write(peerAddress string, msg []byte) error {
	conn, err := net.Dial("tcp", peerAddress)
	if err != nil {
		return fmt.Errorf("failed to dial peer %s: %w", peerAddress, err)
	}
	defer conn.Close()

	if _, err := conn.Write(msg); err != nil {
		return err
	}

	return nil
}

func (node *Node) GetCurrentAddress() string {
	return "127.0.0.1" + node.TcpAddress
}

func (node *Node) VerifyBlockchain(chain *blockchain.Blockchain) (bool, error) {
	return node.Blockchain.VerifyBlocks(chain.Blocks)
}
