package p2p

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/p2p/types"
	"slices"
	"strings"
	"time"
)

func (node *Node) handleGetHeaders(requestorAddr string) error {
	log.Printf("DEBUG: handleGetHeaders called with requestorAddr: %s", requestorAddr)
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

	log.Printf("DEBUG: Sending %d headers (blocks 0 to %d)", len(blocks), len(blocks)-1)

	headers := make([]blockchain.BlockHeader, len(blocks))
	for idx, block := range blocks {
		headers[idx] = *block.GetHeader()
	}

	payload, err := json.Marshal(headers)
	if err != nil {
		return err
	}

	msg := types.NewMessage(types.SendHeadersMsg, node.GetCurrentAddress(), payload)
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

	msg := types.NewMessage(types.SendBlockMsg, node.GetCurrentAddress(), payload)
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

	msg := types.NewMessage(types.SendHeadersMsg, node.GetCurrentAddress(), payload)

	return node.WriteMessage(ctx, requestorAddr, msg.Marshal())
}

func (node *Node) handleBroadcastBlock(payload types.Payload) error {
	var block blockchain.Block
	if err := payload.Unmarshal(&block); err != nil {
		return fmt.Errorf("failed to unmarshal broadcast block: %w", err)
	}

	log.Printf("Received broadcast block %d, attempting to validate and add", block.Id)

	if err := node.validateAndAddBlock(&block); err != nil {
		return fmt.Errorf("failed to validate broadcast block: %w", err)
	}

	log.Printf("Successfully processed broadcast block %d", block.Id)
	return nil
}

func (node *Node) validateAndAddBlock(block *blockchain.Block) error {
	log.Printf("DEBUG: Validating block %d", block.Id)

	if block.Id < 0 || len(block.Hash) != 32 {
		return fmt.Errorf("invalid block structure")
	}

	// Check if block already exists
	_, err := node.Blockchain.GetBlockById(block.Id)
	if err == nil {
		log.Printf("DEBUG: Block %d already exists, skipping", block.Id)
		return nil
	} else if err != sql.ErrNoRows {
		log.Printf("DEBUG: Error checking if block %d exists: %v", block.Id, err)
		return fmt.Errorf("database error checking for existing block: %w", err)
	}

	log.Printf("DEBUG: Block %d doesn't exist, proceeding to add it", block.Id)

	sqlTx, err := node.Blockchain.Database.BeginTx()
	if err != nil {
		return err
	}

	defer sqlTx.Rollback()

	if err := node.Blockchain.AddBlock(sqlTx, block); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Printf("Block %d already exists in database", block.Id)
			return nil // Block already exists, not an error
		}
		log.Printf("Failed to add block %d to database: %v", block.Id, err)
		return err
	}

	log.Printf("Successfully added block %d to database", block.Id)

	if err := node.Blockchain.UpdateUserBalances(sqlTx, block.Transactions); err != nil {
		return err
	}

	if err := sqlTx.Commit(); err != nil {
		return err
	}

	// Add to in-memory blockchain if not already there
	node.Blockchain.Mutex.Lock()
	found := false
	for _, existingBlock := range node.Blockchain.Blocks {
		if existingBlock.Id == block.Id {
			found = true
			break
		}
	}
	if !found {
		node.Blockchain.Blocks = append(node.Blockchain.Blocks, *block)
		log.Printf("Added block %d to in-memory blockchain", block.Id)
	}
	node.Blockchain.Mutex.Unlock()

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

	fmt.Println(newPeerAddress)
	node.Peers = append(node.Peers, newPeerAddress)
	fmt.Println(node.Peers)

	if err := node.Blockchain.Database.AddPeer(sqlTx, newPeerAddress); err != nil {
		return err
	}

	return sqlTx.Commit()
}
