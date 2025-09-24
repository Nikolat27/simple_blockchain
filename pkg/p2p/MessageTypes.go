package p2p

type MessageType string

const (
	MsgTransaction   MessageType = "transaction"
	MsgBlock         MessageType = "block"
	MsgPeers         MessageType = "peers"
	MsgBlockchainReq MessageType = "blockchain_request"
	MsgBlockchainRsp MessageType = "blockchain_response"
	MsgBlockReq      MessageType = "block_request"
	MsgBlockRsp      MessageType = "block_response"
)

type Message struct {
	Type      MessageType `json:"type"`
	Payload   any         `json:"payload"`
	From      string      `json:"from"`
	Timestamp int64       `json:"timestamp"`
}
