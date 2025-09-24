package p2p

import (
	"fmt"
	"log"
	"net"
)

type NetworkServer struct {
	Node     *Node
	listener net.Listener
}

func NewNetworkServer(node *Node) *NetworkServer {
	return &NetworkServer{
		Node: node,
	}
}

func (s *NetworkServer) Start(port string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	s.listener = listener

	go s.acceptConnections()
	return nil
}

func (s *NetworkServer) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go s.Node.handleConnection(conn)
	}
}

func (s *NetworkServer) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}
