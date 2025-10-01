package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"simple_blockchain/pkg/p2p/types"
)

func (node *Node) parseMessage(senderMsg []byte) error {
	var msg types.Message

	if err := json.Unmarshal(senderMsg, &msg); err != nil {
		return err
	}

	log.Printf("DEBUG: Received message type: %s from: %s", msg.Type, msg.SenderAddress)

	if msg.SenderAddress == "" {
		return errors.New("msg senderAddress field is empty")
	}

	switch msg.Type {
	// Requesting the blockchain`s data
	case types.RequestHeadersMsg:
		return node.handleGetHeaders(msg.SenderAddress)

		// Sending the blockchain`s data to the applicant
	case types.SendHeadersMsg:
		node.payloadCh <- msg.Payload

	case types.RequestBlockMsg:
		var blockId int64
		if err := msg.Payload.Unmarshal(&blockId); err != nil {
			return err
		}

		return node.handleGetBlock(msg.SenderAddress, blockId)

	case types.SendBlockMsg:
		node.payloadCh <- msg.Payload

	case types.BroadcastBlockMsg:
		return node.handleBroadcastBlock(msg.Payload)

	default:
		fmt.Println("meow meow")
	}

	return nil
}
