package p2p

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/p2p/types"
	"slices"
	"time"
)

func (node *Node) handleGetBlockHeaders(requestorAddr string) error {
	if err := node.AddNewPeer(requestorAddr); err != nil {
		log.Printf("Failed to add peer %s: %v", requestorAddr, err)
		// Continue
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	blocks, err := node.Blockchain.GetAllBlocks()
	if err != nil {
		return err
	}

	headers := make([]blockchain.BlockHeader, len(blocks))
	for idx, block := range blocks {
		headers[idx] = *block.GetHeader()
	}

	payload, err := json.Marshal(headers)
	if err != nil {
		return err
	}

	msg := types.NewMessage(types.SendBlockHeadersMsg, node.GetCurrentTcpAddress(), payload)
	return node.WriteMessage(ctx, requestorAddr, msg.Marshal())
}

func (node *Node) handleGetBlock(requestorAddr string, blockId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	block, err := node.Blockchain.GetBlockById(blockId)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(block)
	if err != nil {
		return err
	}

	msg := types.NewMessage(types.SendBlockMsg, node.GetCurrentTcpAddress(), payload)

	return node.WriteMessage(ctx, requestorAddr, msg.Marshal())
}

func (node *Node) handleGetBlockchainData(requestorAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	blocks, err := node.Blockchain.GetAllBlocks()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(blocks)
	if err != nil {
		return err
	}

	msg := types.NewMessage(types.SendBlockHeadersMsg, node.GetCurrentTcpAddress(), payload)

	return node.WriteMessage(ctx, requestorAddr, msg.Marshal())
}

// handleBlockBroadcasting -> Propose the new block
func (node *Node) handleBlockBroadcasting(payload types.Payload) error {
	var block blockchain.Block
	if err := payload.Unmarshal(&block); err != nil {
		return fmt.Errorf("failed to unmarshal broadcast block: %w", err)
	}

	log.Println("handleBlockBroadcasting Current Node: ", node.GetCurrentTcpAddress())

	valid, err := node.Blockchain.VerifyBlock(&block)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("block is corrupted")
	}

	sqlTx, err := node.Blockchain.Database.BeginTx()
	if err != nil {
		return err
	}
	defer sqlTx.Rollback()

	if err := node.Blockchain.AddBlock(sqlTx, &block); err != nil {
		return err
	}

	if err := sqlTx.Commit(); err != nil {
		return err
	}

	node.Blockchain.AddBlockToMemory(&block)

	return nil
}

// handleMempoolBroadcasting -> Propose the new mempool
func (node *Node) handleMempoolBroadcasting(payload types.Payload) error {
	var newMempool blockchain.Mempool
	if err := payload.Unmarshal(&newMempool); err != nil {
		return fmt.Errorf("failed to unmarshal broadcast block: %w", err)
	}

	log.Println("handleMempoolBroadcasting Current Node: ", node.GetCurrentTcpAddress())

	node.Blockchain.Mempool.SyncMempool(&newMempool)

	return nil
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

	if err := node.Blockchain.Database.AddPeer(sqlTx, newPeerAddress); err != nil {
		return err
	}

	if err := sqlTx.Commit(); err != nil {
		return err
	}

	node.addPeerToMemory(newPeerAddress)

	return nil
}

func (node *Node) addPeerToMemory(newPeer string) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	node.Peers = append(node.Peers, newPeer)
}
