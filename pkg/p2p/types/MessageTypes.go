package types

import (
	"encoding/json"
	"log"
)

const (
	RequestHeadersMsg = "get_headers_msg"
	SendHeadersMsg    = "send_headers_msg"
	RequestBlockMsg   = "get_block_msg" // NEW: request specific block
	SendBlockMsg      = "send_block_msg"
	BroadcastBlockMsg = "broadcast_block_msg"
)

type Payload []byte

func (payload *Payload) Unmarshal(v any) error {
	return json.Unmarshal(*payload, v)
}

type Message struct {
	Type          string  `json:"type"`
	SenderAddress string  `json:"sender_address"`
	Payload       Payload `json:"payload"`
}

func NewMessage(typ, senderAddr string, payload Payload) *Message {
	return &Message{
		Type:          typ,
		SenderAddress: senderAddr,
		Payload:       payload,
	}
}

func (msg *Message) Marshal() []byte {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("ERROR marshaling msg")
	}

	return data
}
