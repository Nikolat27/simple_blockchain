package p2p

import (
	"errors"
	"fmt"
	"slices"
)

func (node *Node) handleNodeJoinNetwork(requestorAddr string) error {
	// Requesting the application node`s blockchain data
	getBlockchainMsg := NewMessage(SendBlockchainDataMsg, node.GetCurrentAddress(), nil)

	// Request blockchain data
	if err := node.Write(requestorAddr, getBlockchainMsg.Marshal()); err != nil {
		return err
	}

	// Waits for blockchain data from the node`s channel
	blocks := <-node.blockchainRespCh

	valid, err := node.Blockchain.VerifyBlocks(blocks)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("received blockchain is corrupted")
	}

	node.AddNewPeer(requestorAddr)

	fmt.Printf("Node: %s verified Node: %s\n", node.GetCurrentAddress(), requestorAddr)
	return nil
}

func (node *Node) handleGetBlockchainData(requestorAddr string) error {
	blocks, err := node.Blockchain.GetAllBlocks()
	if err != nil {
		return err
	}

	msg := NewMessage(GetBlockchainDataMsg, node.GetCurrentAddress(), blocks)

	return node.Write(requestorAddr, msg.Marshal())
}

func (node *Node) AddNewPeer(newPeerAddress string) {
	if !slices.Contains(node.Peers, newPeerAddress) {
		node.Peers = append(node.Peers, newPeerAddress)
	}
}
