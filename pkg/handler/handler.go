package handler

import (
	"simple_blockchain/pkg/p2p"
)

type Handler struct {
	Node *p2p.Node
}

func New(node *p2p.Node) *Handler {
	return &Handler{
		Node: node,
	}
}
