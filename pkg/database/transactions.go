package database

import (
	"database/sql"
)

// DBTransactionSchema represents a transaction as stored in the database
type DBTransactionSchema struct {
	From       string
	To         string
	Amount     uint64
	Fee        uint64
	Timestamp  int64
	PublicKey  string
	Signature  string
	Status     string
	IsCoinbase bool
}

func (db *Database) GetTransactionsByBlockId(blockId int) ([]DBTransactionSchema, error) {
	query := `
			SELECT sender, recipient, amount, fee, timestamp, public_key, signature, status, is_coin_base
			FROM transactions
			WHERE block_id = ?
			ORDER BY id
		`

	rows, err := db.db.Query(query, blockId)
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

		err := rows.Scan(&sender, &tx.To, &tx.Amount, &tx.Fee, &tx.Timestamp,
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
			tx.Signature = signature.String
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (db *Database) AddTransaction(dbTx *sql.Tx, tx DBTransactionSchema, blockId int) error {
	query := `
		INSERT INTO transactions(block_id, sender, recipient, amount, fee, timestamp, public_key, signature, status, is_coin_base)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var sender any = nil
	if tx.From != "" {
		sender = tx.From
	}

	var publicKey any = nil
	if tx.PublicKey != "" {
		publicKey = tx.PublicKey
	}

	var signature any = nil
	if len(tx.Signature) > 0 {
		signature = tx.Signature
	}

	_, err := dbTx.Exec(query, blockId, sender, tx.To, tx.Amount, tx.Fee, tx.Timestamp,
		publicKey, signature, tx.Status, tx.IsCoinbase)

	return err
}
