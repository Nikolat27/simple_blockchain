// Package handler provides HTTP request handlers for the blockchain API.
package handler

import (
	"github.com/Nikolat27/simple_blockchain/pkg/p2p"
)

// Handler handles HTTP requests for blockchain operations.
// It maintains a reference to the P2P node which provides access to the blockchain and network.
type Handler struct {
	Node *p2p.Node
}

// New creates a new Handler instance with the given P2P node.
func New(node *p2p.Node) *Handler {
	return &Handler{
		Node: node,
	}
}
