package p2p

import (
	"encoding/json"
	"log"
	"simple_blockchain/pkg/blockchain"
)

const (
	JoinNetworkMsg        = "join_network_msg"
	SendBlockchainDataMsg = "get_blockchain_msg"
	GetBlockchainDataMsg  = "get_blockchain_data_msg"
)

type Message struct {
	Type          string             `json:"type"`
	SenderAddress string             `json:"sender_address"`
	Blocks        []blockchain.Block `json:"blocks"`
}

func NewMessage(typ, senderAddr string, blocks []blockchain.Block) *Message {
	return &Message{
		Type:          typ,
		SenderAddress: senderAddr,
		Blocks:        blocks,
	}
}

func (msg *Message) Marshal() []byte {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("ERROR marshaling msg")
	}

	return data
}
