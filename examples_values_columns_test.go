package scan_test

import (
	"fmt"

	"github.com/blockloop/scan"
)

func ExampleValues() {
	person := struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}{
		ID:   1,
		Name: "Brett",
	}

	cols := []string{"id", "name"}
	vals := scan.Values(cols, &person)
	fmt.Printf("%+v", vals)
	// Output:
	// [1 Brett]
}

func ExampleColumns() {
	var person struct {
		ID   int `db:"person_id"`
		Name string
	}

	cols := scan.Columns(&person)
	fmt.Printf("%+v", cols)
	// Output:
	// [person_id Name]
}

func ExampleColumns_exclude() {
	var person struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
		Age  string `db:"-"`
	}

	cols := scan.Columns(&person)
	fmt.Printf("%+v", cols)
	// Output:
	// [id name]
}

func ExampleColumnsStrict() {
	var person struct {
		ID   int `db:"id"`
		Name string
		Age  string `db:"age"`
	}

	cols := scan.ColumnsStrict(&person)
	fmt.Printf("%+v", cols)
	// Output:
	// [id age]
}
