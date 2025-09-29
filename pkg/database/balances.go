package database

import (
	"database/sql"
	"errors"
	"fmt"
)

func (db *Database) GetConfirmedBalance(address string) (uint64, error) {
	query := `
		SELECT balance
		FROM balances
		WHERE address = ?
	`

	var balance uint64
	if err := db.db.QueryRow(query, address).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, err
	}

	return balance, nil
}

func (db *Database) IncreaseUserBalance(tx *sql.Tx, address string, amount uint64) error {
	query := `
		INSERT INTO balances(address, balance)
		VALUES (?, ?)
		ON CONFLICT (address) DO UPDATE SET balance = balance + excluded.balance
	`

	if _, err := tx.Exec(query, address, amount); err != nil {
		return err
	}

	return nil
}

func (db *Database) DecreaseUserBalance(tx *sql.Tx, address string, amount uint64) error {
	// First, check if balance is sufficient
	var currentBalance uint64
	checkQuery := `SELECT balance FROM balances WHERE address = ?`

	if err := tx.QueryRow(checkQuery, address).Scan(&currentBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("insufficient balance for address %s", address)
		}
		return err
	}

	if currentBalance < amount {
		return fmt.Errorf("insufficient balance for address %s", address)
	}

	// If sufficient, perform the update
	updateQuery := `UPDATE balances SET balance = balance - ? WHERE address = ?`
	_, err := tx.Exec(updateQuery, amount, address)

	return err
}
