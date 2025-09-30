package p2p

import (
	"fmt"
	"slices"
)

func (node *Node) handleNodeJoinNetwork(requestorAddr string) error {
	// Requesting the application node`s blockchain data
	getBlockchainMsg := NewMessage(SendBlockchainDataMsg, node.GetCurrentAddress(), nil)

	// Request blockchain data
	if err := node.WriteMessage(requestorAddr, getBlockchainMsg.Marshal()); err != nil {
		return err
	}

	fmt.Printf("Node: %s verified Node: %s\n", node.GetCurrentAddress(), requestorAddr)
	return nil
}

func (node *Node) handleGetBlockchainData(requestorAddr string) error {
	blocks, err := node.Blockchain.GetAllBlocks()
	if err != nil {
		return err
	}

	msg := NewMessage(GetBlockchainDataMsg, node.GetCurrentAddress(), blocks)

	return node.WriteMessage(requestorAddr, msg.Marshal())
}

func (node *Node) AddNewPeer(newPeerAddress string) error {
	if slices.Contains(node.Peers, newPeerAddress) {
		return nil
	}

	sqlTx, err := node.Blockchain.Database.BeginTx()
	if err != nil {
		return err
	}
	defer sqlTx.Rollback()

	node.Peers = append(node.Peers, newPeerAddress)

	if err := node.Blockchain.Database.AddPeer(sqlTx, newPeerAddress); err != nil {
		return err
	}

	return sqlTx.Commit()
}
