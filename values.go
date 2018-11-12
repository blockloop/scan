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
		return nil, fmt.Errorf("Values: %v", err)
	}

	fields := loadFields(model)

	for i, col := range cols {
		j, ok := fields[col]
		if !ok {
			return nil, fmt.Errorf("field %T.%q either does not exist or is unexported", v, col)
		}

		vals[i] = model.Field(j).Interface()
	}
	return vals, nil
}

func loadFields(val reflect.Value) map[string]int {
	if cache, cached := valuesCache.Load(val); cached {
		return cache.(map[string]int)
	}
	return writeFieldsCache(val)
}

func writeFieldsCache(val reflect.Value) map[string]int {
	typ := val.Type()
	numfield := val.NumField()
	m := map[string]int{}

	for i := 0; i < numfield; i++ {
		if !val.Field(i).CanSet() {
			continue
		}

		field := typ.Field(i)
		m[field.Name] = i
		if tag, ok := field.Tag.Lookup(dbTag); ok {
			m[tag] = i
		}
	}
	valuesCache.Store(val, m)
	return m
}
