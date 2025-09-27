package LevelDB

import (
	"errors"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDB struct {
	db *leveldb.DB
}

func New(fileAddress string) (*LevelDB, error) {
	if fileAddress == "" {
		fileAddress = "balances"
	}

	dbInstance, err := leveldb.OpenFile(fileAddress, nil)
	if err != nil {
		return nil, err
	}

	return &LevelDB{
		db: dbInstance,
	}, nil
}

func (ld *LevelDB) Get(key, defaultValue []byte) ([]byte, error) {
	value, err := ld.db.Get(key, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return defaultValue, nil
		}

		return nil, err
	}

	return value, nil
}

func (ld *LevelDB) Set(key, value []byte) error {
	return ld.db.Put(key, value, nil)
}

func (ld *LevelDB) IncreaseUserBalance(address []byte, amount int) error {
	// Get current balance
	value, err := ld.db.Get(address, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			// default balance
			value = []byte("0")
		} else {
			return err
		}
	}

	// Convert current balance to int
	currentBalanceInt, err := strconv.Atoi(string(value))
	if err != nil {
		return err
	}

	newBalance := currentBalanceInt + amount

	return ld.db.Put(address, []byte(strconv.Itoa(newBalance)), nil)
}

func (ld *LevelDB) DecreaseUserBalance(address []byte, amount int) error {
	// Get current balance
	value, err := ld.db.Get(address, nil)
	if err != nil {
		return err
	}

	// Convert current balance to int
	currentBalanceInt, err := strconv.Atoi(string(value))
	if err != nil {
		return err
	}

	newBalance := currentBalanceInt - amount

	return ld.db.Put(address, []byte(strconv.Itoa(newBalance)), nil)
}

func (ld *LevelDB) Close() error {
	return ld.db.Close()
}
