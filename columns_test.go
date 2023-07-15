package scan

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColumnsErrorsWhenNotAPointer(t *testing.T) {
	_, err := Columns(1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pointer")
}

func TestColumnsErrorsWhenNotAStruct(t *testing.T) {
	var num int
	_, err := Columns(&num)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "struct")
}

func TestColumnsReturnsFieldNames(t *testing.T) {
	type person struct {
		Name string
	}

	cols, err := Columns(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"Name"}, cols)
}

func TestColumnsReturnsStructTags(t *testing.T) {
	type person struct {
		Name string `db:"name"`
	}

	cols, err := Columns(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"name"}, cols)
}

func TestColumnsReturnsStructTagsAndFieldNames(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int
	}

	cols, err := Columns(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"name", "Age"}, cols)
}

func TestColumnsIgnoresPrivateFields(t *testing.T) {
	type person struct {
		name string `db:"name"`
		Age  int
	}

	cols, err := Columns(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"Age"}, cols)
}

func TestColumnsAddsComplexTypesWhenNoStructTag(t *testing.T) {
	type person struct {
		Address struct {
			Street string
		}
	}

	cols, err := Columns(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"Street"}, cols)
}

func TestColumnsAddsComplexTypesWhenStructTag(t *testing.T) {
	type person struct {
		Address struct {
			Street string `db:"address.street"`
		}
	}

	cols, err := Columns(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"address.street"}, cols)
}

func TestColumnsDoesNotAddStructTag(t *testing.T) {
	type person struct {
		Address struct {
			Street string `db:"address.street"`
		} `db:"address"`
	}

	cols, err := Columns(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"address.street"}, cols)
}

func TestColumnsStrictAddsComplexTypesWhenStructTag(t *testing.T) {
	type person struct {
		Address struct {
			Street string `db:"address.street"`
		}
	}

	cols, err := ColumnsStrict(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"address.street"}, cols)
}

func TestColumnsStrictDoesNotAddComplexTypesWhenNoStructTag(t *testing.T) {
	type person struct {
		Address struct {
			Street string
		}
	}

	cols, err := ColumnsStrict(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{}, cols)
}

func TestColumnsStrictDoesNotAddComplexTypesWhenStructTagIgnored(t *testing.T) {
	type person struct {
		Address struct {
			Street string `db:"address.street"`
		} `db:"-"`
	}

	cols, err := ColumnsStrict(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{}, cols)
}

func TestColumnsIgnoresComplexTypesWhenNoStructTag(t *testing.T) {
	type person struct {
		Address struct {
			Street string
		}
	}

	cols, err := Columns(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"Street"}, cols)
}

func TestColumnsExcludesFields(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int    `db:"age"`
	}

	cols, err := ColumnsStrict(&person{}, "name")
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"age"}, cols)
}

func TestColumnsExcludesFieldsFromCache(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int    `db:"age"`
	}

	// Create the cache first
	thing := &person{}
	cols1, err := Columns(thing)
	assert.NoError(t, err)
	assert.Equal(t, []string{"name", "age"}, cols1)

	// verify that all cached fields aren't returned
	cols2, err := Columns(thing, "age")
	assert.NoError(t, err)
	assert.Equal(t, []string{"name"}, cols2)
}

func TestColumnsIncludesFieldsCachedFromFirstExclude(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int    `db:"age"`
	}

	// Create the cache first
	thing := &person{}
	cols1, err := Columns(thing, "age")
	assert.NoError(t, err)
	assert.Equal(t, []string{"name"}, cols1)

	// verify that all fields are returned
	cols2, err := Columns(thing)
	assert.NoError(t, err)
	assert.Equal(t, []string{"name", "age"}, cols2)
}

func TestColumnsStrictExcludesUntaggedFields(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int
	}

	cols, err := ColumnsStrict(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"name"}, cols)
}

func TestColumnsIgnoresDashTag(t *testing.T) {
	type person struct {
		Name string `db:"name"`
		Age  int    `db:"-"`
	}

	cols, err := ColumnsStrict(&person{})
	assert.NoError(t, err)
	assert.EqualValues(t, []string{"name"}, cols)
}

func TestColumnsReturnsAllFieldNames(t *testing.T) {
	s := new(largeStruct)
	exp := reflect.Indirect(reflect.ValueOf(s)).NumField()

	cols, err := Columns(s)
	assert.NoError(t, err)
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

	cols, err := Columns(&person)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, cols)
}

func BenchmarkColumnsLargeStruct(b *testing.B) {
	ls := &largeStruct{ID: "test", Index: 88, UUID: "test", IsActive: false, Balance: "test", Picture: "test", Age: 88, EyeColor: "test", Name: "test", Gender: "test", Company: "test", Email: "test", Phone: "test", Address: "test", About: "test", Registered: "test", Latitude: 0.566439688205719, Longitude: 0.48440760374069214, Greeting: "test", FavoriteFruit: "test", AID: "test", AIndex: 19, AUUID: "test", AIsActive: true, ABalance: "test", APicture: "test", AAge: 12, AEyeColor: "test", AName: "test", AGender: "test", ACompany: "test", AEmail: "test", APhone: "test", AAddress: "test", AAbout: "test", ARegistered: "test", ALatitude: 0.16338545083999634, ALongitude: 0.24648870527744293, AGreeting: "test", AFavoriteFruit: "test"}

	for i := 0; i < b.N; i++ {
		Columns(ls)
	}
}
