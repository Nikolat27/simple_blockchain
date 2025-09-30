-- +goose Up
CREATE TABLE IF NOT EXISTS blocks
(
    id           INTEGER Primary Key AUTOINCREMENT,
    prev_hash    TEXT UNIQUE NOT NULL,
    hash         TEXT UNIQUE NOT NULL,
    nonce        INTEGER DEFAULT (0),
    timestamp    INTEGER DEFAULT (strftime('%s', 'now')),
    block_height INTEGER DEFAULT (0)
);

CREATE TABLE IF NOT EXISTS transactions
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    block_id     INTEGER NULL,
    sender       TEXT    NULL,
    recipient    TEXT    NOT NULL,
    amount       INTEGER          DEFAULT (0),
    fee          INTEGER          DEFAULT (0),
    timestamp    INTEGER          DEFAULT (strftime('%s', 'now')),
    public_key   TEXT    NULL,
    signature    TEXT    NULL,
    status       TEXT    NOT NULL DEFAULT ('pending') CHECK (status IN ('pending', 'confirmed')),
    is_coin_base INTEGER NOT NULL DEFAULT (0) CHECK (is_coin_base IN (0, 1)),
    FOREIGN KEY (block_id) REFERENCES blocks (id) ON DELETE SET NULL
);

-- indexes
CREATE INDEX idx_transactions_block_id ON transactions (block_id);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_block_id;

DROP TABLE IF EXISTS blocks;
DROP TABLE IF EXISTS transactions;
