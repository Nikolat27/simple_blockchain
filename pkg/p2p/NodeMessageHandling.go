package p2p

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"simple_blockchain/pkg/CryptoGraphy"
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

	newMessage := types.NewMessage(types.SendBlockHeadersMsg, node.GetCurrentTcpAddress(), payload)

	return node.WriteMessage(ctx, requestorAddr, newMessage.Marshal())
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

func (node *Node) verifyBlockTransactions(block *blockchain.Block) error {
	for _, tx := range block.Transactions {
		if tx.IsCoinbase {
			continue
		}

		if !tx.Verify() {
			return fmt.Errorf("transaction %x has invalid signature", tx.Hash())
		}

		derivedAddress, err := CryptoGraphy.DeriveAddressFromPublicKey(tx.PublicKey)
		if err != nil {
			return err
		}

		if derivedAddress != tx.From {
			return fmt.Errorf("transaction %x sender address mismatch", tx.Hash())
		}

		if tx.Amount == 0 {
			return fmt.Errorf("transaction %x has zero amount", tx.Hash())
		}

		if tx.Amount+tx.Fee <= 0 {
			return fmt.Errorf("transaction %x has invalid amount/fee combination", tx.Hash())
		}
	}

	return nil
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

	if err := node.verifyBlockTransactions(&block); err != nil {
		return fmt.Errorf("block contains invalid transactions: %w", err)
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

	node.Blockchain.Mempool.DeleteMinedTransactions(block.Transactions)

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

func (node *Node) handleCancelMining() error {
	log.Println("handleCancelMining Current Node: ", node.GetCurrentTcpAddress())

	node.Blockchain.CancelMiningCh <- true

	return nil
}

func (node *Node) AddNewPeer(newPeerAddress string) error {
	node.Mutex.RLock()
	if slices.Contains(node.Peers, newPeerAddress) {
		return nil
	}
	node.Mutex.RUnlock()

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

	node.Mutex.Lock()
	node.addPeerToMemory(newPeerAddress)
	node.Mutex.Unlock()

	return nil
}

func (node *Node) addPeerToMemory(newPeer string) {
	node.Peers = append(node.Peers, newPeer)
}
