package scan_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/blockloop/scan/v2"
)

func TestCustomScanner(t *testing.T) {
	t.Parallel()

	db := mustDB("sql_scanner", `
		CREATE TABLE test
		(
			id int PRIMARY KEY,
			data int NOT NULL
		)`,
		`INSERT INTO test (id, data) VALUES (1, 123), (2, 234)`,
	)
	t.Cleanup(func() { db.Close() })

	const (
		selectOneQuery = `SELECT data FROM test WHERE id = 1`
		selectAllQuery = `SELECT data FROM test`
	)

	t.Run("scan.Row must work", func(t *testing.T) {
		var data customScanner
		rows, err := db.Query(selectOneQuery)
		require.NoError(t, err)

		err = scan.Row(&data, rows)
		require.NoError(t, err)
		require.Equal(t, customScanner{v: 123}, data)
	})

	t.Run("scan.Rows must work", func(t *testing.T) {
		var data []customScanner
		rows, err := db.Query(selectAllQuery)
		require.NoError(t, err)

		err = scan.Rows(&data, rows)
		require.NoError(t, err)
		require.ElementsMatch(t, []customScanner{{v: 123}, {v: 234}}, data)
	})

}

type customScanner struct {
	v int64
}

func (c *customScanner) Scan(src interface{}) error {
	switch v := src.(type) {
	case int64:
		*c = customScanner{v: v}
	case []byte: // for ramsql
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return fmt.Errorf("parse int: %w", err)
		}
		*c = customScanner{v: n}
	case nil:
		return nil
	default:
		return fmt.Errorf("unsupported type %T", src)
	}

	return nil
}
