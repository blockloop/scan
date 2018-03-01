package scan

import "database/sql"

// RowsScanner is a database scanner for many rows. It is most commonly the
// result of *sql.DB Query(...)
type RowsScanner interface {
	Close() error
	Scan(dest ...interface{}) error
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
	Err() error
	Next() bool
}

// cache is an interface for a sync.Map that is used for cache internally
type cache interface {
	Delete(key interface{})
	Load(key interface{}) (value interface{}, ok bool)
	LoadOrStore(key interface{}, value interface{}) (actual interface{}, loaded bool)
	Range(f func(key interface{}, value interface{}) bool)
	Store(key interface{}, value interface{})
}
