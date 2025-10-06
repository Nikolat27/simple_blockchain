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

reset-db:
	sudo rm -f blockchain_db.sqlite

migrations:
	goose -dir migrations sqlite3 blockchain_db.sqlite up

rollback:
	goose -dir migrations sqlite3 blockchain_db.sqlite down

tls-cert:
	openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/C=US/ST=State/L=City/O=Org/CN=localhost"

godoc:
	@echo "Starting godoc server..."
	@echo "Documentation will be available at:"
	@echo "  http://localhost:6060/pkg/github.com/Nikolat27/simple_blockchain/pkg/handler/"
	@echo ""
	@godoc -http=:6060
