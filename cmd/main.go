package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"simple_blockchain/pkg/HttpServer"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/database"
	"simple_blockchain/pkg/handler"
	"simple_blockchain/pkg/p2p"
	"simple_blockchain/pkg/utils"
)

func main() {
	httpPort := flag.String("port", "8000", "http port")
	tcpPort := flag.String("node-port", "8080", "tcp port")

	flag.Parse()

	peerAddress := fmt.Sprintf(":%s", *tcpPort)

	if err := utils.LoadEnv(); err != nil {
		panic(err)
	}

	dbDriverName := os.Getenv("DB_DRIVER_NAME")
	dataSourceName := os.Getenv("DATA_SOURCE_NAME")

	dbInstance, err := database.New(dbDriverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	defer dbInstance.Close()

	mempool := blockchain.NewMempool()

	// Load or initialize blockchain
	bc, err := blockchain.LoadBlockchain(dbInstance, mempool)
	if err != nil {
		panic(err)
	}

	if len(bc.Blocks) == 0 {
		bc, err = blockchain.NewBlockchain(dbInstance, mempool)
		if err != nil {
			panic(err)
		}
	}

	// Start node
	node, err := p2p.SetupNode(peerAddress, bc)
	if err != nil {
		panic(err)
	}

	// Load existing peers from DB if any
	allPeers, err := node.Blockchain.Database.LoadPeers()
	if err != nil {
		panic(err)
	}

	node.Peers = allPeers

	// Bootstrap node with DNS seeds
	go node.Bootstrap()

	// HTTP handlers
	newHandler := handler.New(node)
	httpServer := HttpServer.New(*httpPort, newHandler)

	if err := httpServer.Run(); err != nil {
		log.Fatal(err)
	}
}
