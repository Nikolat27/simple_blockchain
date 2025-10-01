.PHONY: migrations

runserver:
	@go run cmd/main.go

run-node1:
	go run cmd/main.go --port=8000 --node-port=8080

run-node2:
	cp blockchain_db.sqlite node2_blockchain_db.sqlite
	go run cmd/main.go --port=5000 --node-port=7000

build:
	@go build -o blockchain-node ./cmd/main.go

view-leveldb:
	@leveldb-viewer -db balances

reset-db:
	rm -f blockchain_db.sqlite

migrations:
	goose -dir migrations sqlite3 blockchain_db.sqlite up

rollback:
	goose -dir migrations sqlite3 blockchain_db.sqlite down
