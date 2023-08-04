package scan

import (
	"fmt"
	"reflect"
	"sync"
)

var valuesCache cache = &sync.Map{}

// Values scans a struct and returns the values associated with the columns
// provided. Only simple value types are supported (i.e. Bool, Ints, Uints,
// Floats, Interface, String)
func Values(cols []string, v interface{}) ([]interface{}, error) {
	vals := make([]interface{}, len(cols))
	model, err := reflectValue(v)
	if err != nil {
		return nil, fmt.Errorf("values: %w", err)
	}

	fields := loadFields(model)

	for i, col := range cols {
		j, ok := fields[col]
		if !ok {
			return nil, fmt.Errorf("field %T.%q either does not exist or is unexported: %w", v, col, ErrStructFieldMissing)
		}

		vals[i] = model.FieldByIndex(j).Interface()
	}
	return vals, nil
}

func loadFields(val reflect.Value) map[string][]int {
	if cache, cached := valuesCache.Load(val.Type()); cached {
		return cache.(map[string][]int)
	}
	return writeFieldsCache(val)
}

func writeFieldsCache(val reflect.Value) map[string][]int {
	m := map[string][]int{}
	writeFields(val, m, []int{})
	valuesCache.Store(val.Type(), m)
	return m
}

func writeFields(val reflect.Value, m map[string][]int, index []int) {
	typ := val.Type()
	numfield := val.NumField()

	for i := 0; i < numfield; i++ {
		valField := val.Field(i)
		if !valField.CanSet() {
			continue
		}

		field := typ.Field(i)
		fieldIndex := append(index, field.Index...)

		if field.Type.Kind() == reflect.Struct && !isValidSqlValue(valField) {
			writeFields(valField, m, fieldIndex)
			continue
		}

		m[field.Name] = fieldIndex
		if tag, ok := field.Tag.Lookup(dbTag); ok {
			m[tag] = fieldIndex
		}
	}
}
