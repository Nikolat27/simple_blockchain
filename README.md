# Simple Blockchain

A p2p blockchain implementation in Go with POW mining, transaction validation, and secure TLS communication.

## Features

- **Proof-of-Work Mining**: Configurable difficulty with block rewards
- **Transaction Management**: Mempool for pending transactions with fee-based prioritization
- **P2P Network**: TLS-encrypted peer-to-peer communication with DNS-based peer discovery
- **SQLite Storage**: Persistent blockchain data with migration support
- **HTTP API**: RESTful endpoints for blockchain interaction
- **Wallet System**: Public/private key pair generation and transaction signing
- **Balance Tracking**: Real-time balance updates with pending transaction consideration

## Prerequisites

- Go 1.24.0 or higher
- SQLite3
- OpenSSL (for TLS certificate generation)
- Goose (for database migrations)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd simple_blockchain
```

2. Install dependencies:
```bash
go mod download
```

3. Generate TLS certificates:
```bash
make tls-cert
```

4. Run database migrations:
```bash
make migrations
```

## Usage

### Single Node

```bash
make runserver
# or
go run cmd/main.go --port=8000 --node-port=8080 --dsn=blockchain_db.sqlite
```

### Multi-Node Network

Run each command in a separate terminal:

```bash
# Node 1
make run-node1

# Node 2
make run-node2

# Node 3
make run-node3
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/chain` | Get full blockchain |
| GET | `/api/blocks` | Get all blocks |
| GET | `/api/mempool` | View pending transactions |
| GET | `/api/balance?address=<addr>` | Check wallet balance |
| GET | `/api/txs` | Get all transactions |
| GET | `/api/tx/fee` | Get current transaction fee |
| POST | `/api/tx/send` | Send transaction |
| POST | `/api/mine` | Mine new block |
| POST | `/api/keys` | Generate key pair |
| DELETE | `/api/clear` | Clear database |

## Configuration

Command-line flags:

- `--port`: HTTP server port (default: 8000)
- `--node-port`: P2P TCP port (default: 8080)
- `--dsn`: Database file path (default: blockchain_db.sqlite)

Environment variables (`.env`):

- `DB_DRIVER_NAME`: Database driver (sqlite3)

## Project Structure

```
.
├── cmd/                    # Application entry point
├── pkg/
│   ├── blockchain/         # Core blockchain logic
│   ├── CryptoGraphy/       # Key generation and signing
│   ├── database/           # SQLite persistence layer
│   ├── handler/            # HTTP request handlers
│   ├── HttpServer/         # HTTP server and routing
│   ├── p2p/                # Peer-to-peer networking
│   └── utils/              # Utility functions
└── migrations/             # Database schema migrations
```

## Development

Run tests:
```bash
make run-tests
```

Build binary:
```bash
make build
```

Reset database:
```bash
make reset-db
```

## How It Works

1. **Block Creation**: Transactions are collected in the mempool and included in new blocks
2. **Mining**: Proof-of-work algorithm finds valid block hashes meeting difficulty requirements
3. **Validation**: Each block and transaction is cryptographically verified
4. **Consensus**: Nodes synchronize blockchain state through P2P communication
5. **Persistence**: All blocks and transactions are stored in SQLite

## Constants

- Mining Reward: 10,000 units
- Difficulty: 5 leading zeros
- Mempool Size: 1MB

