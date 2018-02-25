package scan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValuesScansOnlyCols(t *testing.T) {
	type person struct {
		Name string
		Age  int
	}

	p := &person{Name: "Brett"}
	vals := Values([]string{"Name"}, p)

	assert.EqualValues(t, []interface{}{"Brett"}, vals)
}

func TestValuesScansDBTags(t *testing.T) {
	type person struct {
		Name string `db:"n"`
	}

	p := &person{Name: "Brett"}
	vals := Values([]string{"n"}, p)

	assert.EqualValues(t, []interface{}{"Brett"}, vals)
}

func TestValuesPanicsWhenRetrievingUnexportedValues(t *testing.T) {
	type person struct {
		name string
	}

	assert.Panics(t, func() {
		Values([]string{"name"}, &person{})
	})
}

func TestValuesWorksWithBothTagAndFieldName(t *testing.T) {
	type person struct {
		Name string `db:"n"`
	}

	p := &person{Name: "Brett"}
	vals := Values([]string{"Name", "n"}, p)
	assert.EqualValues(t, []interface{}{"Brett", "Brett"}, vals)
}

// benchmarks

func BenchmarkValuesLargeStructCached(b *testing.B) {
	ls := &largeStruct{}
	cols := Columns(ls)

	for i := 0; i < b.N; i++ {
		Values(cols, ls)
	}
}

type largeStruct struct {
	ID            string  `db:"id"`
	Index         int     `db:"index"`
	UUID          string  `db:"uuid"`
	IsActive      bool    `db:"isActive"`
	Balance       string  `db:"balance"`
	Picture       string  `db:"picture"`
	Age           int     `db:"age"`
	EyeColor      string  `db:"eyeColor"`
	Name          string  `db:"name"`
	Gender        string  `db:"gender"`
	Company       string  `db:"company"`
	Email         string  `db:"email"`
	Phone         string  `db:"phone"`
	Address       string  `db:"address"`
	About         string  `db:"about"`
	Registered    string  `db:"registered"`
	Latitude      float64 `db:"latitude"`
	Longitude     float64 `db:"longitude"`
	Greeting      string  `db:"greeting"`
	FavoriteFruit string  `db:"favoriteFruit"`
}
