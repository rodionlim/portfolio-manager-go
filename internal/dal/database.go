package dal

// Database defines the interface for database operations.
type Database interface {
	Close() error
	Get(key string, v interface{}) error
	Put(key string, v interface{}) error
	Delete(key string) error
	GetAllKeysWithPrefix(prefix string) ([]string, error)
}

const (
	LDB = "leveldb"
	RDB = "rocksdb"
)
