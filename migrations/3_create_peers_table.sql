-- +goose Up
CREATE TABLE IF NOT EXISTS peers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tcp_address TEXT UNIQUE NOT NULL
);
CREATE INDEX idx_peers_tcp_address ON peers (tcp_address);
-- +goose Down
DROP TABLE IF EXISTS peers;
DROP INDEX IF EXISTS idx_peers_tcp_address