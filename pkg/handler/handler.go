package handler

import (
	"simple_blockchain/pkg/blockchain"
)

type Handler struct {
	Blockchain *blockchain.Blockchain
}

func New(bc *blockchain.Blockchain) *Handler {
	return &Handler{
		Blockchain: bc,
	}
}
