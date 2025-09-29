package database

import "database/sql"

func (sqlite *Database) AddBlock(prevHash, hash string, nonce, timestamp, blockHeight int64) (int64, error) {
	query := `
		INSERT INTO blocks(prev_hash, hash, nonce, timestamp, block_height)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := sqlite.db.Exec(query, prevHash, hash, nonce, timestamp, blockHeight)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (sqlite *Database) GetAllBlocks() (*sql.Rows, error) {
	query := `
		SELECT id, prev_hash, hash, nonce, timestamp, block_height FROM blocks ORDER BY block_height ASC
	`

	rows, err := sqlite.db.Query(query)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
