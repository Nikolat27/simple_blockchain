package main

import (
	"fmt"
	"log"
	"os"
	"simple_blockchain/pkg/HttpServer"
	"simple_blockchain/pkg/blockchain"
	"simple_blockchain/pkg/database"
	"simple_blockchain/pkg/handler"

	"github.com/joho/godotenv"
)

const Port = "8000"

func main() {
	if err := loadEnv(); err != nil {
		panic(err)
	}

	dbDriverName := os.Getenv("DB_DRIVER_NAME")
	dataSourceName := os.Getenv("DATA_SOURCE_NAME")

	dbInstance, err := database.New(dbDriverName, dataSourceName)
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

func loadEnv() error {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		file, err := os.Create(".env")
		if err != nil {
			return fmt.Errorf("error creating .env file: %v", err)
		}

		defer file.Close()
	}

	return godotenv.Load()
}
