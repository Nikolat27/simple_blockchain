package main

import (
	"log"
	"simple_blockchain/pkg/HttpServer"
	"simple_blockchain/pkg/LevelDB"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/database"
	"simple_blockchain/pkg/handler"
)

const Port = "8000"

func main() {
	sqliteDBInstance, err := database.New("sqlite3", "./blockchain_db.sqlite2")
	if err != nil {
		panic(err)
	}
	defer sqliteDBInstance.Close()

	levelDBInstance, err := LevelDB.New("balances")
	if err != nil {
		panic(err)
	}
	defer levelDBInstance.Close()

	var newBc *blockchain.Blockchain
	var newMempool *blockchain.Mempool

	newBc = blockchain.NewBlockchain("genesis-address", levelDBInstance, sqliteDBInstance)
	newMempool = blockchain.NewMempool()

	newHandler := handler.New(newBc, newMempool)

	httpServer := HttpServer.New(Port, newHandler)

	if err := httpServer.Run(); err != nil {
		log.Fatal(err)
	}
}
