package scan

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCustomScanner(t *testing.T) {
	t.Parallel()

	dsn, cleanup, err := testDatabaseDSN()
	require.NoError(t, err)
	t.Cleanup(cleanup)

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	const createTableQuery = `
CREATE TABLE test
(
	id int PRIMARY KEY AUTO_INCREMENT,
	data int NOT NULL
);

INSERT INTO test (id, data)
VALUES (1, 123), (2, 234);
`

	const (
		selectOneQuery = `SELECT data FROM test WHERE id = 1`
		selectAllQuery = `SELECT data FROM test`
	)

	_, err = db.Exec(createTableQuery)
	require.NoError(t, err)

	t.Run("scan.Row must work", func(t *testing.T) {
		var data customScanner
		rows, err := db.Query(selectOneQuery)
		require.NoError(t, err)

		err = Row(&data, rows)
		require.NoError(t, err)
		require.Equal(t, customScanner{v: 123}, data)
	})

	t.Run("scan.Rows must work", func(t *testing.T) {
		var data []customScanner
		rows, err := db.Query(selectAllQuery)
		require.NoError(t, err)

		err = Rows(&data, rows)
		require.NoError(t, err)
		require.ElementsMatch(t, []customScanner{{v: 123}, {v: 234}}, data)
	})

}

type customScanner struct {
	v int64
}

func (c *customScanner) Scan(src any) error {
	log.Println("custom scan")
	switch v := src.(type) {
	case int64:
		*c = customScanner{v: v}
	case nil:
		return nil
	default:
		return fmt.Errorf("unsupported type %T", src)
	}

	return nil
}

func testDatabaseDSN() (string, func(), error) {
	mysqlContainer, err := testcontainers.GenericContainer(context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "percona/percona-server:8.0.36-28.1-multi",
				ExposedPorts: []string{"3306/tcp"},
				WaitingFor: wait.ForAll(
					wait.ForListeningPort("3306/tcp"),
					wait.ForLog("mysqld: ready for connections"),
				),
				Env: map[string]string{
					"MYSQL_DATABASE":             "test",
					"MYSQL_ALLOW_EMPTY_PASSWORD": "yes",
				},
			},
			Started: true,
		})
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		_ = mysqlContainer.Terminate(context.Background())
	}

	host, err := mysqlContainer.Host(context.Background())
	if err != nil {
		cleanup()
		return "", nil, err
	}

	port, err := mysqlContainer.MappedPort(context.Background(), "3306")
	if err != nil {
		cleanup()
		return "", nil, err
	}

	dsn := fmt.Sprintf("root@(%s:%d)/test?tls=false&multiStatements=true", host, port.Int())
	return dsn, cleanup, nil
}
