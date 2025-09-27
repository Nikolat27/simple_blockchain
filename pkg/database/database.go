package database

import (
	"database/sql"

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
