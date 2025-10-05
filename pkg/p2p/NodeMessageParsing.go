package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
	"simple_blockchain/pkg/p2p/types"
)

func (node *Node) parseMessage(senderMsg []byte) error {
	var msg types.Message

	if err := json.Unmarshal(senderMsg, &msg); err != nil {
		return err
	}

	if msg.SenderAddress == "" {
		return errors.New("msg senderAddress field is empty")
	}

	switch msg.Type {
	// Requesting the blockchain`s data
	case types.RequestHeadersMsg:
		return node.handleGetBlockHeaders(msg.SenderAddress)

	// Sending the blockchain`s data to the applicant node
	case types.SendBlockHeadersMsg:
		node.payloadCh <- msg.Payload

	case types.RequestBlockMsg:
		var blockId int64
		if err := msg.Payload.Unmarshal(&blockId); err != nil {
			return err
		}

		return node.handleGetBlock(msg.SenderAddress, blockId)

	case types.SendBlockMsg:
		node.payloadCh <- msg.Payload

	case types.BlockBroadcastMsg:
		return node.handleBlockBroadcasting(msg.Payload)

	case types.MempoolBroadcastMsg:
		return node.handleMempoolBroadcasting(msg.Payload)

	case types.CancelMiningMsg:
		return node.handleCancelMining()

	default:
		fmt.Println("meow meow")
	}

	return nil
}
