package handler

import (
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/p2p"
)

type Handler struct {
	Blockchain *blockchain.Blockchain
	Mempool    *blockchain.Mempool
	P2PNode    *p2p.Node
}

func NewWithP2P(bc *blockchain.Blockchain, mempool *blockchain.Mempool, p2pNode *p2p.Node) *Handler {
	return &Handler{
		Blockchain: bc,
		Mempool:    mempool,
		P2PNode:    p2pNode,
	}
}
