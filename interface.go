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
