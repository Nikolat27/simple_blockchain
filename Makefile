runserver:
	@go run cmd/main.go

build:
	@go build -o blockchain-node ./cmd/main.go

view-leveldb:
	@leveldb-viewer -db balances

clear:
	rm -rf data
	rm -rf balances
