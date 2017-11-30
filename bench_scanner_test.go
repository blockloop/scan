package scan_test

import (
	"database/sql"
	"testing"

	"github.com/blockloop/scan"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func BenchmarkScanRowOneField(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(b, err)

	_, err = db.Exec(`CREATE TABLE persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name) VALUES ('brett')
	`)
	require.NoError(b, err)

	type item struct {
		First string
	}

	for i := 0; i < b.N; i++ {
		rows, _ := db.Query(`SELECT name FROM persons LIMIT 1`)
		var it item
		scan.Row(&it, rows)
		rows.Close()
	}
}

func BenchmarkDirectScanOneField(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(b, err)

	_, err = db.Exec(`CREATE TABLE persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name) VALUES ('brett')
	`)
	require.NoError(b, err)

	type item struct {
		First string
	}

	for i := 0; i < b.N; i++ {
		rows, _ := db.Query(`SELECT name FROM persons LIMIT 1`)
		var it item
		rows.Next()
		rows.Scan(&it.First)
		rows.Close()
	}
}

func BenchmarkScanRowFiveFields(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(b, err)

	_, err = db.Exec(`CREATE TABLE persons (
		name VARCHAR(120),
		age TINYINT,
		active BOOLEAN,
		city VARCHAR(60),
		state VARCHAR(12)
	);

	INSERT INTO PERSONS (name, age, active, city, state)
	VALUES ('brett', 100, 1, 'dallas', 'tx');
	`)
	require.NoError(b, err)

	type item struct {
		First  string `db:"first"`
		Age    int8   `db:"age"`
		Active bool   `db:"active"`
		City   string `db:"city"`
		State  string `db:"state"`
	}

	for i := 0; i < b.N; i++ {
		rows, _ := db.Query(`SELECT * FROM persons LIMIT 1`)
		var it item
		scan.Row(&it, rows)
		rows.Close()
	}
}

func BenchmarkDirectScanFiveFields(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(b, err)

	_, err = db.Exec(`CREATE TABLE persons (
		name VARCHAR(120),
		age TINYINT,
		active BOOLEAN,
		city VARCHAR(60),
		state VARCHAR(12)
	);

	INSERT INTO PERSONS (name, age, active, city, state)
	VALUES ('brett', 100, 1, 'dallas', 'tx');
	`)
	require.NoError(b, err)

	type item struct {
		First  string `db:"first"`
		Age    int8   `db:"age"`
		Active bool   `db:"active"`
		City   string `db:"city"`
		State  string `db:"state"`
	}

	for i := 0; i < b.N; i++ {
		rows, _ := db.Query(`SELECT "first", "age", "active", "city", "state" FROM persons LIMIT 1`)
		it := item{}
		rows.Next()
		rows.Scan(
			&it.First,
			&it.Age,
			&it.Active,
			&it.City,
			&it.State,
		)
		rows.Close()
	}
}

func BenchmarkScanRowsOneField(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(b, err)

	_, err = db.Exec(`CREATE TABLE persons ( name VARCHAR(120) );

	INSERT INTO PERSONS (name) VALUES ('brett'), ('fred'), ('george'), ('steve')
	`)
	require.NoError(b, err)

	type item struct {
		First string `db:"name"`
	}

	for i := 0; i < b.N; i++ {
		rows, _ := db.Query(`SELECT name FROM persons`)
		var it []item
		scan.Rows(&it, rows)
		rows.Close()
	}
}

func BenchmarkDirectScanManyOneField(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(b, err)

	_, err = db.Exec(`CREATE TABLE persons ( name VARCHAR(120) );

	INSERT INTO PERSONS (name) VALUES ('brett'), ('fred'), ('george'), ('steve')
	`)
	require.NoError(b, err)

	type item struct {
		First string
	}

	for i := 0; i < b.N; i++ {
		rows, _ := db.Query(`SELECT name FROM persons`)
		items := make([]item, 0)
		for rows.Next() {
			var it item
			rows.Scan(&it.First)
			items = append(items, it)
		}
		rows.Close()
	}
}
