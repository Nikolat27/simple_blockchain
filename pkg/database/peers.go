package database

import (
	"database/sql"
	"errors"
)

func (db *Database) AddPeer(sqlTx *sql.Tx, tcpAddress string) error {
	query := `
		INSERT OR IGNORE INTO peers(tcp_address)
		VALUES (?);
	`

	if _, err := sqlTx.Exec(query, tcpAddress); err != nil {
		return err
	}

	return nil
}

func (db *Database) GetPeers() (*sql.Rows, error) {
	query := `
		SELECT tcp_address FROM peers;
	`

	return db.db.Query(query)
}

func (db *Database) LoadPeers() ([]string, error) {
	rows, err := db.db.Query("SELECT tcp_address FROM peers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []string
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, err
		}
		peers = append(peers, addr)
	}

	return peers, rows.Err()
}

func (db *Database) PeerExist(tcpAddress string) (bool, error) {
	query := `
		SELECT 1 FROM peers WHERE tcp_address = ? LIMIT 1
	`

	var exists int
	if err := db.db.QueryRow(query, tcpAddress).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
