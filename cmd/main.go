package main

import (
	"flag"
	"log"
	"simple_blockchain/pkg/HttpServer"
	"simple_blockchain/pkg/LevelDB"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/handler"
)

func main() {
	var (
		httpPort = flag.String("http-port", "8000", "Http server port")
	)
	flag.Parse()

	dbInstance, err := LevelDB.New("balances")
	if err != nil {
		panic(err)
	}
	defer dbInstance.Close()

	var newBc *blockchain.Blockchain
	var newMempool *blockchain.Mempool

	newBc = blockchain.NewBlockchain("genesis-address", dbInstance)
	newMempool = blockchain.NewMempool()

	newHandler := handler.New(newBc, newMempool)

	httpServer := HttpServer.New(*httpPort, newHandler)

	if err := httpServer.Run(); err != nil {
		log.Fatal(err)
	}
}
