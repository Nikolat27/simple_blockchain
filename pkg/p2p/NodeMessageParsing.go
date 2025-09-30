package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
)

func (node *Node) parseMessage(senderMsg []byte) error {
	var msg Message

	if err := json.Unmarshal(senderMsg, &msg); err != nil {
		return err
	}

	if msg.SenderAddress == "" {
		return errors.New("msg senderAddress field is empty")
	}

	switch msg.Type {

	case JoinNetworkMsg:
		return node.handleNodeJoinNetwork(msg.SenderAddress)

		// Requesting the blockchain`s data
	case SendBlockchainDataMsg:
		return node.handleGetBlockchainData(msg.SenderAddress)

		// Sending the blockchain`s data to the applicant
	case GetBlockchainDataMsg:
		node.BlockchainRespCh <- msg.Blocks

	default:
		fmt.Println("meow meow")
	}

	return nil
}
