package scan

import (
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

	if cache, ok := columnsCache.Load(model); ok {
		return cache.([]string), nil
	}

	numfield := model.NumField()
	names := make([]string, 0, numfield)

	isExcluded := func(name string) bool {
		for _, ex := range excluded {
			if ex == name {
				return true
			}
		}
		return false
	}

	for i := 0; i < numfield; i++ {
		valField := model.Field(i)
		if !valField.IsValid() || !valField.CanSet() {
			continue
		}

		typeField := model.Type().Field(i)
		if tag, ok := typeField.Tag.Lookup(dbTag); ok {
			if tag != "-" && !isExcluded(tag) {
				names = append(names, tag)
			}
			continue
		}

		if strict {
			continue
		}

		if isExcluded(typeField.Name) || !supportedColumnType(valField.Kind()) {
			continue
		}

		names = append(names, typeField.Name)
	}

	columnsCache.Store(model, names)
	return names, nil
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

func supportedColumnType(k reflect.Kind) bool {
	switch k {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Interface,
		reflect.String:
		return true
	default:
		return false
	}
}
