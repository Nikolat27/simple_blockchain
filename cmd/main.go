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

const Port = "8000"

func main() {
	tcpPort := flag.String("peerAddress", "8080", "tcp port")

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

	// Try to load existing blockchain
	bc, err := blockchain.LoadBlockchain(dbInstance, mempool)
	if err != nil {
		panic(err)
	}

	// If no blocks exist, create new blockchain with genesis block
	if len(bc.Blocks) == 0 {
		bc, err = blockchain.NewBlockchain(dbInstance, mempool)
		if err != nil {
			panic(err)
		}
	}

	newNode, err := p2p.NewNode(peerAddress, bc)
	if err != nil {
		panic(err)
	}

	// http handlers
	newHandler := handler.New(newNode)

	httpServer := HttpServer.New(Port, newHandler)

	if err := httpServer.Run(); err != nil {
		log.Fatal(err)
	}
}
