package scan

import "reflect"

const dbTag = "db"

// Columns scans a struct and returns a list of strings
// that represent the assumed column names based on the
// db struct tag, or the field name. Any field or struct
// tag that matches a string within the excluded list
// will be excluded from the result
func Columns(v interface{}, excluded ...string) []string {
	return columns(v, false, excluded...)
}

// ColumnsStrict is identical to Columns, but it only
// searches struct tags and excludes fields not tagged
// with the db struct tag
func ColumnsStrict(v interface{}, excluded ...string) []string {
	return columns(v, true, excluded...)
}

func columns(v interface{}, strict bool, excluded ...string) []string {
	vVal := mustReflectValue(v)

	numfield := vVal.NumField()
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
		valField := vVal.Field(i)
		if !valField.IsValid() || !valField.CanSet() {
			continue
		}

		typeField := vVal.Type().Field(i)
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

	return names
}

func mustReflectValue(v interface{}) reflect.Value {
	vType := reflect.TypeOf(v)
	vKind := vType.Kind()
	if vKind != reflect.Ptr {
		panic(vKind.String() + ": must be a pointer")
	}

	vVal := reflect.Indirect(reflect.ValueOf(v))
	if vVal.Kind() != reflect.Struct {
		panic(vKind.String() + ": must be a pointer to a struct")
	}
	return vVal
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

type scanner interface {
	Scan(...interface{}) error
}
