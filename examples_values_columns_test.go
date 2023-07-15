package scan_test

import (
	"fmt"

	"github.com/blockloop/scan/v2"
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
	vals, _ := scan.Values(cols, &person)
	fmt.Printf("%+v", vals)
	// Output:
	// [1 Brett]
}

func ExampleValues_nested() {
	type Address struct {
		Street string
		City   string
	}

	person := struct {
		ID   int
		Name string
		Address
	}{
		Name: "Brett",
		ID:   1,
		Address: Address{
			City: "San Francisco",
		},
	}

	cols := []string{"Name", "City"}
	vals, _ := scan.Values(cols, &person)
	fmt.Printf("%+v", vals)
	// Output:
	// [Brett San Francisco]
}

func ExampleColumns() {
	var person struct {
		ID   int `db:"person_id"`
		Name string
	}

	cols, _ := scan.Columns(&person)
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

	cols, _ := scan.Columns(&person)
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

	cols, _ := scan.ColumnsStrict(&person)
	fmt.Printf("%+v", cols)
	// Output:
	// [id age]
}

func ExampleColumnsNested() {
	var person struct {
		ID      int    `db:"person.id"`
		Name    string `db:"person.name"`
		Company struct {
			ID   int `db:"company.id"`
			Name string
		}
	}

	cols, _ := scan.Columns(&person)
	fmt.Printf("%+v", cols)
	// Output:
	// [person.id person.name company.id Name]
}

func ExampleColumnsNestedStrict() {
	var person struct {
		ID      int    `db:"person.id"`
		Name    string `db:"person.name"`
		Company struct {
			ID   int `db:"company.id"`
			Name string
		}
	}

	cols, _ := scan.ColumnsStrict(&person)
	fmt.Printf("%+v", cols)
	// Output:
	// [person.id person.name company.id]
}

func ExampleColumnsNested_exclude() {
	var person struct {
		ID      int    `db:"person.id"`
		Name    string `db:"person.name"`
		Company struct {
			ID   int    `db:"-"`
			Name string `db:"company.name"`
		}
	}

	cols, _ := scan.Columns(&person)
	fmt.Printf("%+v", cols)
	// Output:
	// [person.id person.name company.name]
}
