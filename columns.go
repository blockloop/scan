package scan

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

const dbTag = "db"

var (
	// ErrNotAPointer is returned when a non-pointer is received
	// when a pointer is expected.
	ErrNotAPointer = errors.New("not a pointer")

	// ErrNotAStructPointer is returned when a non-struct pointer
	// is received but a struct pointer was expected
	ErrNotAStructPointer = errors.New("not a struct pointer")

	// ErrNotASlicePointer is returned when receiving an argument
	// that is expected to be a slice pointer, but it is not
	ErrNotASlicePointer = errors.New("not a slice pointer")

	// ErrStructFieldMissing is returned when trying to scan a value
	// to a column which does not match a struct. This means that
	// the struct does not have a field that matches the column
	// specified.
	ErrStructFieldMissing = errors.New("struct field missing")
)

var columnsCache cache = &sync.Map{}

type cacheKey struct {
	Type   reflect.Type
	Strict bool
}

// Columns scans a struct and returns a list of strings
// that represent the assumed column names based on the
// db struct tag, or the field name. Any field or struct
// tag that matches a string within the excluded list
// will be excluded from the result.
func Columns(v interface{}, excluded ...string) ([]string, error) {
	return columns(v, false, excluded...)
}

// ColumnsStrict is identical to Columns, but it only
// searches struct tags and excludes fields not tagged
// with the db struct tag.
func ColumnsStrict(v interface{}, excluded ...string) ([]string, error) {
	return columns(v, true, excluded...)
}

func columns(v interface{}, strict bool, excluded ...string) ([]string, error) {
	model, err := reflectValue(v)
	if err != nil {
		return nil, fmt.Errorf("columns: %w", err)
	}

	key := cacheKey{model.Type(), strict}

	if cache, ok := columnsCache.Load(key); ok {
		cached := cache.([]string)
		res := make([]string, 0, len(cached))

		keep := func(k string) bool {
			for _, c := range excluded {
				if c == k {
					return false
				}
			}
			return true
		}

		for _, k := range cached {
			if keep(k) {
				res = append(res, k)
			}
		}
		return res, nil
	}

	names := columnNames(model, strict, excluded...)
	toCache := append(names, excluded...)
	columnsCache.Store(key, toCache)
	return names, nil
}

func columnNames(model reflect.Value, strict bool, excluded ...string) []string {
	numfield := model.NumField()
	names := make([]string, 0, numfield)

	for i := 0; i < numfield; i++ {
		valField := model.Field(i)
		if !valField.IsValid() || !valField.CanSet() {
			continue
		}

		typeField := model.Type().Field(i)

		if typeField.Type.Kind() == reflect.Struct && !isValidSqlValue(valField) {
			embeddedNames := columnNames(valField, strict, excluded...)
			names = append(names, embeddedNames...)
			continue
		}

		fieldName := typeField.Name
		if tag, hasTag := typeField.Tag.Lookup(dbTag); hasTag {
			if tag == "-" {
				continue
			}
			fieldName = tag
		} else if strict {
			// there's no tag name and we're in strict mode so move on
			continue
		}

		if isExcluded(fieldName, excluded...) {
			continue
		}

		if supportedColumnType(valField) || isValidSqlValue(valField) {
			names = append(names, fieldName)
		}
	}

	return names
}

func isExcluded(name string, excluded ...string) bool {
	for _, ex := range excluded {
		if ex == name {
			return true
		}
	}
	return false
}

func reflectValue(v interface{}) (reflect.Value, error) {
	vType := reflect.TypeOf(v)
	vKind := vType.Kind()
	if vKind != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("%q must be a pointer: %w", vKind.String(), ErrNotAPointer)
	}

	vVal := reflect.Indirect(reflect.ValueOf(v))
	if vVal.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("%q must be a pointer to a struct: %w", vKind.String(), ErrNotAStructPointer)
	}
	return vVal, nil
}

func supportedColumnType(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Interface,
		reflect.String:
		return true
	case reflect.Ptr:
		ptrVal := reflect.New(v.Type().Elem())
		return supportedColumnType(ptrVal.Elem())
	default:
		return false
	}
}

func isValidSqlValue(v reflect.Value) bool {
	// This method covers two cases in which we know the Value can be converted to sql:
	// 1. It returns true for sql.driver's type check for types like time.Time
	// 2. It implements the driver.Valuer interface allowing conversion directly
	//    into sql statements
	if v.Kind() == reflect.Ptr {
		ptrVal := reflect.New(v.Type().Elem())
		return isValidSqlValue(ptrVal.Elem())
	}

	if driver.IsValue(v.Interface()) {
		return true
	}

	valuerType := reflect.TypeOf((*driver.Valuer)(nil)).Elem()
	return v.Type().Implements(valuerType)
}
