package database

import (
	"database/sql"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func New(driverName, dataSourceName string) (*Database, error) {
	if driverName == "" {
		driverName = "sqlite3"
	}

	if dataSourceName == "" {
		dataSourceName = "./blockchain_db.sqlite"
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
}

func (sqlite *Database) Close() error {
	return sqlite.db.Close()
}

func (sqlite *Database) Version() (string, error) {
	var version string

	if err := sqlite.db.QueryRow("SELECT sqlite_version()").Scan(&version); err != nil {
		return "", err
	}

	return version, nil
}

// Begin -> For isolation level
func (sqlite *Database) Begin() (*sql.Tx, error) {
	return sqlite.db.Begin()
}

// ClearAllData removes all blockchain data (used when corruption is detected)
func (sqlite *Database) ClearAllData() error {
	queries := []string{
		"DELETE FROM transactions",
		"DELETE FROM blocks",
		"DELETE FROM balances",
	}

	for _, query := range queries {
		if _, err := sqlite.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
