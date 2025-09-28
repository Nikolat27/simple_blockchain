package database

import (
	"database/sql"
)

// DBTransactionSchema represents a transaction as stored in the database
type DBTransactionSchema struct {
	From       string
	To         string
	Amount     uint64
	Timestamp  int64
	PublicKey  string
	Signature  []byte
	Status     string
	IsCoinbase bool
}

func (sqlite *Database) GetTransactionsByBlockId(blockId int) ([]DBTransactionSchema, error) {
	query := `
		SELECT sender, recipient, amount, timestamp, public_key, signature, status, is_coin_base
		FROM transactions
		WHERE block_id = ?
		ORDER BY id ASC
	`

	rows, err := sqlite.db.Query(query, blockId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []DBTransactionSchema
	for rows.Next() {
		var tx DBTransactionSchema
		var sender sql.NullString
		var publicKey sql.NullString
		var signature sql.NullString

		err := rows.Scan(&sender, &tx.To, &tx.Amount, &tx.Timestamp,
			&publicKey, &signature, &tx.Status, &tx.IsCoinbase)
		if err != nil {
			return nil, err
		}

		if sender.Valid {
			tx.From = sender.String
		}
		if publicKey.Valid {
			tx.PublicKey = publicKey.String
		}
		if signature.Valid {
			tx.Signature = []byte(signature.String)
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (sqlite *Database) AddTransaction(tx DBTransactionSchema, blockId int) error {
	query := `
		INSERT INTO transactions(block_id, sender, recipient, amount, timestamp, public_key, signature, status, is_coin_base)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var sender interface{} = nil
	if tx.From != "" {
		sender = tx.From
	}

	var publicKey interface{} = nil
	if tx.PublicKey != "" {
		publicKey = tx.PublicKey
	}

	var signature interface{} = nil
	if len(tx.Signature) > 0 {
		signature = string(tx.Signature)
	}

	_, err := sqlite.db.Exec(query, blockId, sender, tx.To, tx.Amount, tx.Timestamp,
		publicKey, signature, tx.Status, tx.IsCoinbase)

	return err
}
