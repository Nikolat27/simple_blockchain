package p2p

type Node struct {
	TcpAddress string
	Peers []Node
}
