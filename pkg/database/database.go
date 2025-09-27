package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	DB *sql.DB
}

func New(driverName, dataSourceName string) (*Database, error) {
	if driverName == "" {
		driverName = "sqlite3"
	}

	if dataSourceName == "" {
		dataSourceName = "./blockchain_db.sqlite2"
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return &Database{
		DB: db,
	}, nil
}

func (sqlite *Database) Close() error {
	return sqlite.DB.Close()
}

func (sqlite *Database) Version() (string, error) {
	var version string

	if err := sqlite.DB.QueryRow("SELECT sqlite_version()").Scan(&version); err != nil {
		return "", err
	}

	return version, nil
}

func (sqlite *Database) GetBalance(address string) (uint64, error) {
	query := `
		SELECT balance
		FROM balances
		WHERE address = ?
	`

	var balance uint64
	if err := sqlite.DB.QueryRow(query, address).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, err
	}

	return balance, nil
}

func (sqlite *Database) IncreaseUserBalance(address string, amount uint64) error {
	query := ` 
		INSERT INTO balances(address, balance)
		VALUES (?, ?)
		ON CONFLICT (address) DO UPDATE SET balance = balance + excluded.balance	
	`

	_, err := sqlite.DB.Exec(query, address, amount)

	if err != nil {
		return err
	}

	return nil
}

func (sqlite *Database) DecreaseUserBalance(address string, amount uint64) error {
	query := `
		INSERT INTO balances(address, balance)
		VALUES (?, ?)
		ON CONFLICT (address) DO UPDATE SET balance = balance + excluded.balance
	`

	result, err := sqlite.DB.Exec(query, address, amount)

	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("insufficient balance for address %s", address)
	}

	return nil
}
