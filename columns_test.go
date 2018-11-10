package scan

import (
	"reflect"
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

func TestColumnsIgnoresDashTag(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int    `db:"-"`
	}

	cols := ColumnsStrict(&person{})
	assert.EqualValues(t, []string{"name"}, cols)
}

func TestColumnsReturnsAllFieldNames(t *testing.T) {
	s := new(largeStruct)
	exp := reflect.Indirect(reflect.ValueOf(s)).NumField()

	cols := Columns(s)
	assert.EqualValues(t, exp, len(cols))
}

func TestColumnsReadsFromCacheFirst(t *testing.T) {
	var person struct {
		ID   int64
		Name string
	}

	v := reflect.Indirect(reflect.ValueOf(&person))
	expected := []string{"fake"}
	columnsCache.Store(v, expected)

	assert.EqualValues(t, expected, Columns(&person))
}

func BenchmarkColumnsLargeStruct(b *testing.B) {
	ls := &largeStruct{ID: "test", Index: 88, UUID: "test", IsActive: false, Balance: "test", Picture: "test", Age: 88, EyeColor: "test", Name: "test", Gender: "test", Company: "test", Email: "test", Phone: "test", Address: "test", About: "test", Registered: "test", Latitude: 0.566439688205719, Longitude: 0.48440760374069214, Greeting: "test", FavoriteFruit: "test", AID: "test", AIndex: 19, AUUID: "test", AIsActive: true, ABalance: "test", APicture: "test", AAge: 12, AEyeColor: "test", AName: "test", AGender: "test", ACompany: "test", AEmail: "test", APhone: "test", AAddress: "test", AAbout: "test", ARegistered: "test", ALatitude: 0.16338545083999634, ALongitude: 0.24648870527744293, AGreeting: "test", AFavoriteFruit: "test"}

	for i := 0; i < b.N; i++ {
		Columns(ls)
	}
}
