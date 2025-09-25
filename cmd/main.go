package main

import (
	"flag"
	"log"
	"simple_blockchain/pkg/HttpServer"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/handler"
	"simple_blockchain/pkg/p2p"
	"simple_blockchain/pkg/storage"
	"strings"
	"time"
)

func main() {
	var (
		nodePort    = flag.String("port", "9000", "P2P server port")
		httpPort    = flag.String("http-port", "8000", "Http server port")
		peers       = flag.String("peers", "", "Comma-separated list of peer addresses (e.g., 'localhost:9001,localhost:9002')")
		genesisAddr = flag.String("genesis", "genesis-wallet", "Genesis wallet address")
		dataDir     = flag.String("data-dir", "./data", "Data directory for persistence")
	)
	flag.Parse()

	genesisAddress := *genesisAddr

	store, err := storage.New(*dataDir)
	if err != nil {
		log.Printf("ERROR creating new storage: %v\n", err)
		return
	}

	var newBc *blockchain.Blockchain
	var newMempool *blockchain.Mempool

	if savedBc, err := store.LoadBlockchain(); err != nil {
		log.Printf("Creating new blockchain: %v", err)
		newBc = blockchain.NewBlockchain(genesisAddress)
		newMempool = blockchain.NewMempool()
	} else {
		log.Printf("Loaded existing blockchain with %d blocks", len(savedBc.Blocks))
		newBc = savedBc

		newMempool = blockchain.NewMempool()
		if transactions, err := store.LoadMempoolTransactions(); err != nil {
			log.Printf("Error loading mempool: %v\n", err)
			return
		} else {
			log.Printf("Loaded %d transactions into mempool", len(transactions))
			for _, tx := range transactions {
				newMempool.AddTransaction(&tx)
			}
		}
	}

	nodeAddr := "localhost:" + *nodePort
	p2pNode := p2p.NewNode(nodeAddr, newBc, newMempool, store)
	p2pNode.StartMessageProcessor()

	// Load saved peers
	p2pNode.LoadPeersFromStorage()

	// Connect to bootstrap peers
	if *peers != "" {
		peerList := strings.Split(*peers, ",")
		for _, peerAddr := range peerList {
			peerAddr = strings.TrimSpace(peerAddr)
			if peerAddr == "" || peerAddr == nodeAddr {
				continue
			}

			p2pNode.AddPeer(peerAddr)
			if err := p2pNode.ConnectToPeer(peerAddr); err != nil {
				log.Printf("Failed to connect to peer %s: %v", peerAddr, err)
			}

		}
	}

	// Start blockchain synchronization
	go func() {
		// Wait a bit for connections to establish
		time.Sleep(3 * time.Second)
		p2pNode.SyncBlockchain()
	}()

	p2pServer := p2p.NewNetworkServer(p2pNode)

	if err := p2pServer.Start(*nodePort); err != nil {
		log.Printf("Failed to start p2p server: %v\n", err)
	}

	// Start periodic saving
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := store.SaveBlockchain(newBc); err != nil {
				log.Printf("Error saving blockchain: %v", err)
			}
			if err := store.SaveMempool(newMempool); err != nil {
				log.Printf("Error saving mempool: %v", err)
			}
		}
	}()

	newHandler := handler.NewWithP2P(newBc, newMempool, p2pNode)

	httpServer := HttpServer.NewHttpServer(*httpPort, newHandler)

	log.Printf("Node started - P2P port: %s, HTTP port: %s, Data dir: %s", *nodePort, *httpPort, *dataDir)
	if err := httpServer.Run(); err != nil {
		log.Fatal(err)
	}
}
