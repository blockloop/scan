package scan_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/blockloop/scan"
	_ "github.com/proullon/ramsql/driver"
)

func mustDB(name string, queries ...string) *sql.DB {
	db, err := sql.Open("ramsql", name)
	if err != nil {
		panic(err)
	}

	for _, s := range queries {
		_, err = db.Exec(s)
		if err != nil {
			panic(err)
		}
	}
	return db
}

func exampleDB() *sql.DB {
	return mustDB(fmt.Sprintf("%d", rand.Uint64()), `CREATE TABLE persons (
		id INTEGER NOT NULL,
		name VARCHAR(120)
	);`,
		`INSERT INTO persons (id, name) VALUES (1, 'brett');`,
		`INSERT INTO persons (id, name) VALUES (2, 'fred');`,
	)
}

func ExampleRow() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT id,name FROM persons LIMIT 1")
	if err != nil {
		panic(err)
	}

	var person struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	err = scan.Row(&person, rows)
	if err != nil {
		panic(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(&person)
	if err != nil {
		panic(err)
	}
	// Output:
	// {"ID":1,"Name":"brett"}
}

func ExampleRowStrict() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT id,name FROM persons LIMIT 1")
	if err != nil {
		panic(err)
	}

	var person struct {
		ID   int
		Name string `db:"name"`
	}

	err = scan.RowStrict(&person, rows)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(os.Stdout).Encode(&person)
	// Output:
	// {"ID":0,"Name":"brett"}
}

func ExampleRow_scalar() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT name FROM persons LIMIT 1")
	if err != nil {
		panic(err)
	}

	var name string

	err = scan.Row(&name, rows)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%q", name)
	// Output:
	// "brett"
}

func ExampleRows() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT id,name FROM persons ORDER BY id ASC")
	if err != nil {
		panic(err)
	}

	var persons []struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	err = scan.Rows(&persons, rows)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(os.Stdout).Encode(&persons)
	// Output:
	// [{"ID":1,"Name":"brett"},{"ID":2,"Name":"fred"}]
}

func ExampleRowsStrict() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT id,name FROM persons ORDER BY id ASC")
	if err != nil {
		panic(err)
	}

	var persons []struct {
		ID   int
		Name string `db:"name"`
	}

	err = scan.Rows(&persons, rows)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(os.Stdout).Encode(&persons)
	// Output:
	// [{"ID":0,"Name":"brett"},{"ID":0,"Name":"fred"}]
}

func ExampleRows_primitive() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT name FROM persons ORDER BY id ASC")
	if err != nil {
		panic(err)
	}

	var names []string
	err = scan.Rows(&names, rows)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(os.Stdout).Encode(&names)
	// Output:
	// ["brett","fred"]
}
