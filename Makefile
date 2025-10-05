.PHONY: migrations

runserver:
	@go run cmd/main.go

run-node1:
	go run cmd/main.go --port=8000 --node-port=8080 --dsn=blockchain_db.sqlite

run-node2:
	cp blockchain_db.sqlite node2_blockchain_db.sqlite
	go run cmd/main.go --port=5000 --node-port=7000 --dsn=node2_blockchain_db.sqlite

run-node3:
	cp blockchain_db.sqlite node3_blockchain_db.sqlite
	go run cmd/main.go --port=5001 --node-port=7001 --dsn=node3_blockchain_db.sqlite


build:
	@go build -o blockchain-node ./cmd/main.go

run-tests:
	@go test ./... -v

view-leveldb:
	@leveldb-viewer -db balances

reset-db:
	rm -f blockchain_db.sqlite

migrations:
	goose -dir migrations sqlite3 blockchain_db.sqlite up

rollback:
	goose -dir migrations sqlite3 blockchain_db.sqlite down
