package handler

import (
	"simple_blockchain/pkg/blockchain"
)

type Handler struct {
	Blockchain *blockchain.Blockchain
	Mempool    *blockchain.Mempool
}

func New(bc *blockchain.Blockchain, mempool *blockchain.Mempool) *Handler {
	return &Handler{
		Blockchain: bc,
		Mempool:    mempool,
	}
}
