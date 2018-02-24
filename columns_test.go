package scan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumnsPanicsWhenNotAPointer(t *testing.T) {
	assert.Panics(t, func() {
		Columns(1)
	})
}

func TestColumnsPanicsWhenNotAStruct(t *testing.T) {
	var num int
	assert.Panics(t, func() {
		Columns(&num)
	})
}

func TestColumnsReturnsFieldNames(t *testing.T) {
	type person struct {
		Name string
	}

	cols := Columns(&person{})
	assert.EqualValues(t, []string{"Name"}, cols)
}

func TestColumnsReturnsStructTags(t *testing.T) {
	type person struct {
		Name string `db:"name"`
	}

	cols := Columns(&person{})
	assert.EqualValues(t, []string{"name"}, cols)
}

func TestColumnsReturnsStructTagsAndFieldNames(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int
	}

	cols := Columns(&person{})
	assert.EqualValues(t, []string{"name", "Age"}, cols)
}

func TestColumnsIgnoresPrivateFields(t *testing.T) {
	type person struct {
		name string `db:"name"`
		Age  int
	}

	cols := Columns(&person{})
	assert.EqualValues(t, []string{"Age"}, cols)
}

func TestColumnsAddsComplexTypesWhenStructTag(t *testing.T) {
	type person struct {
		Address struct {
			Street string
		} `db:"address"`
	}

	cols := Columns(&person{})
	assert.EqualValues(t, []string{"address"}, cols)
}

func TestColumnsIgnoresComplexTypesWhenNoStructTag(t *testing.T) {
	type person struct {
		Address struct {
			Street string
		}
	}

	cols := Columns(&person{})
	assert.EqualValues(t, []string{}, cols)
}

func TestColumnsExcludesFields(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int    `db:"age"`
	}

	cols := ColumnsStrict(&person{}, "name")
	assert.EqualValues(t, []string{"age"}, cols)
}

func TestColumnsStrictExcludesUntaggedFields(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int
	}

	cols := ColumnsStrict(&person{})
	assert.EqualValues(t, []string{"name"}, cols)
}
