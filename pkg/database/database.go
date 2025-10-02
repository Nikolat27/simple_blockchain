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

func (db *Database) Close() error {
	return db.db.Close()
}

func (db *Database) Version() (string, error) {
	var version string

	if err := db.db.QueryRow("SELECT sqlite_version()").Scan(&version); err != nil {
		return "", err
	}

	return version, nil
}

// BeginTx -> For isolation level
func (db *Database) BeginTx() (*sql.Tx, error) {
	return db.db.Begin()
}

// ClearAllData -> Flush the database
func (db *Database) ClearAllData(sqlTx *sql.Tx) error {
	queries := []string{
		"DELETE FROM transactions",
		"DELETE FROM blocks",
		"DELETE FROM balances",
	}

	for _, query := range queries {
		if _, err := sqlTx.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
