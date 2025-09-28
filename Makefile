runserver:
	@go run cmd/main.go

build:
	@go build -o blockchain-node ./cmd/main.go

view-leveldb:
	@leveldb-viewer -db balances

migrations:
	goose -dir ./migrations sqlite3 ./blockchain_db.sqlite up

rollback:
	goose -dir ./migrations sqlite3 ./blockchain_db.sqlite down
