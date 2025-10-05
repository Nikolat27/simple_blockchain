package database

import (
	"database/sql"
)

func (db *Database) AddBlock(sqlTx *sql.Tx, prevHash, hash, merkleRoot string, nonce, timestamp, blockHeight int64) (int64, error) {

	query := `
		INSERT INTO blocks(prev_hash, hash, merkle_root, nonce, timestamp, block_height)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := sqlTx.Exec(query, prevHash, hash, merkleRoot, nonce, timestamp, blockHeight)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (db *Database) GetAllBlocks() (*sql.Rows, error) {
	query := `
			SELECT id, prev_hash, hash, merkle_root, nonce, timestamp, block_height
			FROM blocks ORDER BY block_height
		`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (db *Database) GetBlocksCount() (int64, error) {
	query := `
		SELECT COUNT(*) FROM blocks
	`

	var count int64
	if err := db.DB.QueryRow(query).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (db *Database) GetBlockById(blockId int64) (*sql.Row, error) {
	query := `
		SELECT * FROM blocks WHERE block_height = ?
	`

	// Debug: check how many rows this would return
	var count int
	countQuery := `SELECT COUNT(*) FROM blocks WHERE block_height = ?`

	if err := db.DB.QueryRow(countQuery, blockId).Scan(&count); err != nil {
		return nil, err
	}

	row := db.DB.QueryRow(query, blockId)
	return row, nil
}
