package dal

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDB struct {
	db *leveldb.DB
}

func NewLevelDB(dbPath string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open LevelDB: %w", err)
	}
	return &LevelDB{db}, nil
}

func (l *LevelDB) Close() error {
	return l.db.Close()
}

func (l *LevelDB) Get(key string, v interface{}) error {
	data, err := l.db.Get([]byte(key), nil)
	if err != nil {
		return fmt.Errorf("failed to get data for key %s: %w", key, err)
	}

	if err = json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal data for key %s: %w", key, err)
	}

	return nil
}

func (l *LevelDB) Put(key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal data for key %s: %w", key, err)
	}

	err = l.db.Put([]byte(key), data, nil)
	if err != nil {
		return fmt.Errorf("failed to put data for key %s: %w", key, err)
	}

	return nil
}

func (l *LevelDB) Delete(key string) error {
	err := l.db.Delete([]byte(key), nil)
	if err != nil {
		return fmt.Errorf("failed to delete data for key %s: %w", key, err)
	}

	return nil
}

func (l *LevelDB) GetAllTickers() ([]string, error) {
	iter := l.db.NewIterator(nil, nil)
	defer iter.Release()

	var tickers []string
	for iter.Next() {
		key := iter.Key()
		if string(key) == "ticker" {
			value := iter.Value()
			ticker := string(value)
			tickers = append(tickers, ticker)
			log.Printf("Retrieved ticker: %s\n", ticker)
		}
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("failed to iterate over database: %w", err)
	}

	return tickers, nil
}
