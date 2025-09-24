runserver:
	@go run cmd/main.go

build:
	@go build -o blockchain-node ./cmd/main.go
