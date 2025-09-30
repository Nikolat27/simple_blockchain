runserver:
	@go run cmd/main.go

run-node1:
	go run cmd/main.go --port=8000 --node-port=8080

run-node2:
	go run cmd/main.go --port=5000 --node-port=7000

build:
	@go build -o blockchain-node ./cmd/main.go

view-leveldb:
	@leveldb-viewer -db balances

migrations:
	goose -dir migrations sqlite3 ./blockchain_db.sqlite up

rollback:
	goose -dir ./migrations sqlite3 ./blockchain_db.sqlite down
