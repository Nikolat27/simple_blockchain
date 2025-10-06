package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Nikolat27/simple_blockchain/pkg/HttpServer"
	"github.com/Nikolat27/simple_blockchain/pkg/blockchain"
	"github.com/Nikolat27/simple_blockchain/pkg/database"
	"github.com/Nikolat27/simple_blockchain/pkg/handler"
	"github.com/Nikolat27/simple_blockchain/pkg/p2p"
	"github.com/Nikolat27/simple_blockchain/pkg/utils"
)

func main() {
	httpPort := flag.String("port", "8000", "http port")
	tcpPort := flag.String("node-port", "8080", "tcp port")
	dbDSN := flag.String("dsn", "blockchain_db.sqlite", "database data source name")

	flag.Parse()

	peerAddress := fmt.Sprintf(":%s", *tcpPort)

	if err := utils.LoadEnv(); err != nil {
		panic(err)
	}

	dbDriverName := os.Getenv("DB_DRIVER_NAME")
	dataSourceName := *dbDSN

	dbInstance, err := database.New(dbDriverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	defer dbInstance.Close()

	mempool := blockchain.NewMempool(1048576)

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

	tlsConfig, err := utils.InitTLS("cert.pem", "key.pem")
	if err != nil {
		panic(err)
	}

	// Start node
	node, err := p2p.SetupNode(peerAddress, bc, tlsConfig)
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
