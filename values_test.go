package scan

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValuesScansOnlyCols(t *testing.T) {
	type person struct {
		Name string
		Age  int
	}

	p := &person{Name: "Brett"}
	vals, err := Values([]string{"Name"}, p)
	require.NoError(t, err)

	assert.EqualValues(t, []interface{}{"Brett"}, vals)
}

func TestValuesScansDBTags(t *testing.T) {
	type person struct {
		Name string `db:"n"`
	}

	p := &person{Name: "Brett"}
	vals, err := Values([]string{"n"}, p)
	require.NoError(t, err)

	assert.EqualValues(t, []interface{}{"Brett"}, vals)
}

func TestValuesScansNestedFields(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}

	type Person struct {
		Name string
		Age  int
		Address
	}

	p := &Person{Name: "Brett", Address: Address{Street: "123 Main St", City: "San Francisco"}}

	vals, err := Values([]string{"Name", "Street", "City"}, p)
	require.NoError(t, err)

	assert.EqualValues(t, []interface{}{"Brett", "123 Main St", "San Francisco"}, vals)
}

func TestValuesReturnsErrorWhenPassingNonPointer(t *testing.T) {
	_, err := Values([]string{"Name"}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pointer")
}

func TestValuesReturnsErrorWhenRetrievingUnexportedValues(t *testing.T) {
	type person struct {
		name string
	}

	fieldName := "name"
	_, err := Values([]string{fieldName}, &person{name: "Brett"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), fieldName)
}

func TestValuesWorksWithBothTagAndFieldName(t *testing.T) {
	type person struct {
		Name string `db:"n"`
	}

	p := &person{Name: "Brett"}
	vals, err := Values([]string{"Name", "n"}, p)
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{"Brett", "Brett"}, vals)
}

func TestValuesReturnsAllFieldNames(t *testing.T) {
	s := new(largeStruct)
	exp := reflect.Indirect(reflect.ValueOf(s)).NumField()

	cols, err := Columns(s)
	assert.NoError(t, err)
	vals, err := Values(cols, s)
	require.NoError(t, err)
	assert.EqualValues(t, exp, len(vals))
}

func TestValuesReadsFromCacheFirst(t *testing.T) {
	person := struct {
		Name string
	}{
		Name: "Brett",
	}

	v := reflect.Indirect(reflect.ValueOf(&person)).Type()
	valuesCache.Store(v, map[string][]int{"fake": {0}})

	vals, err := Values([]string{"fake"}, &person)
	require.NoError(t, err)
	assert.EqualValues(t, []interface{}{"Brett"}, vals)
}

// benchmarks

func BenchmarkValuesLargeStruct(b *testing.B) {
	ls := &largeStruct{ID: "test", Index: 88, UUID: "test", IsActive: false, Balance: "test", Picture: "test", Age: 88, EyeColor: "test", Name: "test", Gender: "test", Company: "test", Email: "test", Phone: "test", Address: "test", About: "test", Registered: "test", Latitude: 0.566439688205719, Longitude: 0.48440760374069214, Greeting: "test", FavoriteFruit: "test", AID: "test", AIndex: 19, AUUID: "test", AIsActive: true, ABalance: "test", APicture: "test", AAge: 12, AEyeColor: "test", AName: "test", AGender: "test", ACompany: "test", AEmail: "test", APhone: "test", AAddress: "test", AAbout: "test", ARegistered: "test", ALatitude: 0.16338545083999634, ALongitude: 0.24648870527744293, AGreeting: "test", AFavoriteFruit: "test"}
	cols, err := Columns(ls)
	assert.NoError(b, err)

	for i := 0; i < b.N; i++ {
		Values(cols, ls)
	}
}

type largeStruct struct {
	ID             string  `db:"id"`
	Index          int     `db:"index"`
	UUID           string  `db:"uuid"`
	IsActive       bool    `db:"isActive"`
	Balance        string  `db:"balance"`
	Picture        string  `db:"picture"`
	Age            int     `db:"age"`
	EyeColor       string  `db:"eyeColor"`
	Name           string  `db:"name"`
	Gender         string  `db:"gender"`
	Company        string  `db:"company"`
	Email          string  `db:"email"`
	Phone          string  `db:"phone"`
	Address        string  `db:"address"`
	About          string  `db:"about"`
	Registered     string  `db:"registered"`
	Latitude       float64 `db:"latitude"`
	Longitude      float64 `db:"longitude"`
	Greeting       string  `db:"greeting"`
	FavoriteFruit  string  `db:"favoriteFruit"`
	AID            string  `db:"aid"`
	AIndex         int     `db:"aindex"`
	AUUID          string  `db:"auuid"`
	AIsActive      bool    `db:"aisActive"`
	ABalance       string  `db:"abalance"`
	APicture       string  `db:"apicture"`
	AAge           int     `db:"aage"`
	AEyeColor      string  `db:"aeyeColor"`
	AName          string  `db:"aname"`
	AGender        string  `db:"agender"`
	ACompany       string  `db:"acompany"`
	AEmail         string  `db:"aemail"`
	APhone         string  `db:"aphone"`
	AAddress       string  `db:"aaddress"`
	AAbout         string  `db:"aabout"`
	ARegistered    string  `db:"aregistered"`
	ALatitude      float64 `db:"alatitude"`
	ALongitude     float64 `db:"alongitude"`
	AGreeting      string  `db:"agreeting"`
	AFavoriteFruit string  `db:"afavoriteFruit"`
}
