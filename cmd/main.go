package main

import (
	"log"
	"simple_blockchain/pkg/HttpServer"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/database"
	"simple_blockchain/pkg/handler"
)

const Port = "8000"

func main() {
	dbInstance, err := database.New("sqlite3", "./blockchain_db.sqlite2")
	if err != nil {
		panic(err)
	}
	defer dbInstance.Close()

	var newBc *blockchain.Blockchain
	var newMempool *blockchain.Mempool

	newBc = blockchain.NewBlockchain(dbInstance)
	newMempool = blockchain.NewMempool()

	newHandler := handler.New(newBc, newMempool)

	httpServer := HttpServer.New(Port, newHandler)

	if err := httpServer.Run(); err != nil {
		log.Fatal(err)
	}
}
