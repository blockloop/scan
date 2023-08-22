package scan_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/blockloop/scan/v2"
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
	return mustDB(fmt.Sprintf("%d", rand.Uint64()), `CREATE TABLE person (
		id INTEGER NOT NULL,
		name VARCHAR(120)
	);`,
		`INSERT INTO person (id, name) VALUES (1, 'brett', 1);`,
		`INSERT INTO person (id, name) VALUES (2, 'fred', 1);`,
		`INSERT INTO person (id) VALUES (3);`,
	)
}

func exampleNestedDB() *sql.DB {
	return mustDB(fmt.Sprintf("%d", rand.Uint64()), `
	CREATE TABLE IF NOT EXISTS person (
		id INT NOT NULL,
		name VARCHAR(120),
		company_id INT
	);`,
		`CREATE TABLE IF NOT EXISTS company (
		id INTEGER NOT NULL,
		name VARCHAR(120)
	);
	`,
		`INSERT INTO person (id, name, company_id) VALUES (1, 'brett', 1);`,
		`INSERT INTO person (id, name, company_id) VALUES (2, 'fred', 1);`,
		`INSERT INTO company (id, name) VALUES (1, 'costco');`,
	)
}

func ExampleRow() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT id,name FROM person LIMIT 1")
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

func ExampleRowNested() {
	db := exampleNestedDB()
	defer db.Close()
	rows, err := db.Query(`
		SELECT person.id,person.name,company.name FROM person
		JOIN company on company.id = person.company_id
		LIMIT 1
	`)
	if err != nil {
		panic(err)
	}

	var person struct {
		ID      int    `db:"person.id"`
		Name    string `db:"person.name"`
		Company struct {
			Name string `db:"company.name"`
		}
	}

	err = scan.RowStrict(&person, rows)
	if err != nil {
		panic(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(&person)
	if err != nil {
		panic(err)
	}
	// Output:
	// {"ID":1,"Name":"brett","Company":{"Name":"costco"}}
}

func ExampleRowStrict() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT id,name FROM person LIMIT 1")
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

func ExampleRowPtr() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT id,name FROM person where id = 3 LIMIT 1")
	if err != nil {
		panic(err)
	}

	var person struct {
		ID   int
		Name *string `db:"name"`
	}

	err = scan.RowStrict(&person, rows)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(os.Stdout).Encode(&person)
	// Output:
	// {"ID":0,"Name":null}
}

func ExampleRowPtrType() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT id,name FROM person where id = 3 LIMIT 1")
	if err != nil {
		panic(err)
	}

	type NullableString *string
	var person struct {
		ID   int
		Name NullableString `db:"name"`
	}

	err = scan.RowStrict(&person, rows)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(os.Stdout).Encode(&person)
	// Output:
	// {"ID":0,"Name":null}
}

func ExampleRow_scalar() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT name FROM person LIMIT 1")
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
	rows, err := db.Query("SELECT id,name FROM person ORDER BY id ASC")
	if err != nil {
		panic(err)
	}

	var persons []struct {
		ID   int     `db:"id"`
		Name *string `db:"name"`
	}

	err = scan.Rows(&persons, rows)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(os.Stdout).Encode(&persons)
	// Output:
	// [{"ID":1,"Name":"brett"},{"ID":2,"Name":"fred"},{"ID":3,"Name":null}]
}

func ExampleRowsStrict() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT id,name FROM person ORDER BY id ASC")
	if err != nil {
		panic(err)
	}

	var persons []struct {
		ID   int
		Name *string `db:"name"`
	}

	err = scan.Rows(&persons, rows)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(os.Stdout).Encode(&persons)
	// Output:
	// [{"ID":0,"Name":"brett"},{"ID":0,"Name":"fred"},{"ID":0,"Name":null}]
}

func ExampleRows_primitive() {
	db := exampleDB()
	defer db.Close()
	rows, err := db.Query("SELECT name FROM person WHERE name IS NOT NULL ORDER BY id ASC")
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
