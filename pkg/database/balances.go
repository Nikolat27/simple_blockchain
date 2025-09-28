package database

import (
	"database/sql"
	"errors"
	"fmt"
)

func (sqlite *Database) GetConfirmedBalance(address string) (uint64, error) {
	query := `
		SELECT balance
		FROM balances
		WHERE address = ?
	`

	var balance uint64
	if err := sqlite.db.QueryRow(query, address).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, err
	}

	return balance, nil
}

func (sqlite *Database) IncreaseUserBalance(tx *sql.Tx, address string, amount uint64) error {
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

func (sqlite *Database) DecreaseUserBalance(tx *sql.Tx, address string, amount uint64) error {
	query := `
		INSERT INTO balances(address, balance)
		VALUES (?, ?)
		ON CONFLICT (address) DO UPDATE
		   SET balance = CASE
		   WHEN balance - excluded.balance >= 0 THEN balance - excluded.balance
		ELSE balance -- keep it unchanged
		END;
	`

	result, err := tx.Exec(query, address, amount)

	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("insufficient balance for address %s", address)
	}

	return nil
}
