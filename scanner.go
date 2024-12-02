package scan

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"reflect"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	// ErrTooManyColumns indicates that a select query returned multiple columns and
	// attempted to bind to a slice of a primitive type. For example, trying to bind
	// `select col1, col2 from mutable` to []string
	ErrTooManyColumns = errors.New("too many columns returned for primitive slice")

	// ErrSliceForRow occurs when trying to use Row on a slice
	ErrSliceForRow = errors.New("cannot scan Row into slice")

	// AutoClose is true when scan should automatically close Scanner when the scan
	// is complete. If you set it to false, then you must defer rows.Close() manually
	AutoClose = true

	// OnAutoCloseError can be used to log errors which are returned from rows.Close()
	// By default this is a NOOP function
	OnAutoCloseError = func(error) {}

	// ScannerMapper transforms database field names into struct/map field names
	// E.g. you can set function for convert snake_case into CamelCase
	ScannerMapper = func(name string) string { return cases.Title(language.English).String(name) }
)

// Row scans a single row into a single variable. It requires that you use
// db.Query and not db.QueryRow, because QueryRow does not return column names.
// There is no performance impact in using one over the other. QueryRow only
// defers returning err until Scan is called, which is an unnecessary
// optimization for this library.
func Row(v interface{}, r RowsScanner) error {
	if AutoClose {
		defer closeRows(r)
	}

	return row(v, r, false)
}

// RowStrict scans a single row into a single variable. It is identical to
// Row, but it ignores fields that do not have a db tag
func RowStrict(v interface{}, r RowsScanner) error {
	if AutoClose {
		defer closeRows(r)
	}

	return row(v, r, true)
}

func row(v interface{}, r RowsScanner, strict bool) error {
	vType := reflect.TypeOf(v)
	if k := vType.Kind(); k != reflect.Ptr {
		return fmt.Errorf("%q must be a pointer: %w", k.String(), ErrNotAPointer)
	}

	vType = vType.Elem()
	vVal := reflect.ValueOf(v).Elem()
	if vType.Kind() == reflect.Slice {
		return ErrSliceForRow
	}

	sl := reflect.New(reflect.SliceOf(vType))
	err := rows(sl.Interface(), r, strict)
	if err != nil {
		return err
	}

	sl = sl.Elem()

	if sl.Len() == 0 {
		return sql.ErrNoRows
	}

	vVal.Set(sl.Index(0))

	return nil
}

// Rows scans sql rows into a slice (v)
func Rows(v interface{}, r RowsScanner) (outerr error) {
	if AutoClose {
		defer closeRows(r)
	}

	return rows(v, r, false)
}

// RowsStrict scans sql rows into a slice (v) only using db tags
func RowsStrict(v interface{}, r RowsScanner) (outerr error) {
	if AutoClose {
		defer closeRows(r)
	}

	return rows(v, r, true)
}

func rows(v interface{}, r RowsScanner, strict bool) (outerr error) {
	vType := reflect.TypeOf(v)
	if k := vType.Kind(); k != reflect.Ptr {
		return fmt.Errorf("%q must be a pointer: %w", k.String(), ErrNotAPointer)
	}
	sliceType := vType.Elem()
	if reflect.Slice != sliceType.Kind() {
		return fmt.Errorf("%q must be a slice: %w", sliceType.String(), ErrNotASlicePointer)
	}

	sliceVal := reflect.Indirect(reflect.ValueOf(v))
	itemType := sliceType.Elem()

	cols, err := r.Columns()
	if err != nil {
		return err
	}

	isPrimitive := itemType.Kind() != reflect.Struct

	for r.Next() {
		sliceItem := reflect.New(itemType).Elem()

		var pointers []interface{}
		if isPrimitive {
			if len(cols) > 1 {
				return ErrTooManyColumns
			}
			pointers = []interface{}{sliceItem.Addr().Interface()}
		} else {
			pointers = structPointers(sliceItem, cols, strict)
		}

		if len(pointers) == 0 {
			return nil
		}

		err := r.Scan(pointers...)
		if err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, sliceItem))
	}
	return r.Err()
}

// Initialization the tags from struct.
func initFieldTag(sliceItem reflect.Value, fieldTagMap *map[string]reflect.Value) {
	typ := sliceItem.Type()
	for i := 0; i < sliceItem.NumField(); i++ {
		if typ.Field(i).Anonymous || typ.Field(i).Type.Kind() == reflect.Struct {
			// found an embedded struct
			sliceItemOfAnonymous := sliceItem.Field(i)
			initFieldTag(sliceItemOfAnonymous, fieldTagMap)
		}
		tag, ok := typ.Field(i).Tag.Lookup("db")
		if ok && tag != "" {
			(*fieldTagMap)[tag] = sliceItem.Field(i)
		}
	}
}

var sqlScannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

func structPointers(sliceItem reflect.Value, cols []string, strict bool) []interface{} {
	pointers := make([]interface{}, 0, len(cols))
	fieldTag := make(map[string]reflect.Value, len(cols))
	initFieldTag(sliceItem, &fieldTag)

	for _, colName := range cols {
		var fieldVal reflect.Value
		if v, ok := fieldTag[colName]; ok {
			fieldVal = v
		} else {
			if strict {
				fieldVal = reflect.ValueOf(nil)
			} else {
				fieldVal = sliceItem.FieldByName(ScannerMapper(colName))
				if fieldVal == (reflect.Value{}) {
					if sliceItem.Addr().Type().Implements(sqlScannerType) {
						// probably this is a custom struct that implements sql.Scanner.
						// do our best and don't set "nothing" as a pointer
						fieldVal = sliceItem
					}
				}
			}
		}
		if !fieldVal.IsValid() || !fieldVal.CanSet() {
			// have to add if we found a column because Scan() requires
			// len(cols) arguments or it will error. This way we can scan to
			// a useless pointer
			var nothing interface{}
			pointers = append(pointers, &nothing)
			continue
		}

		pointers = append(pointers, fieldVal.Addr().Interface())
	}
	return pointers
}

func closeRows(c io.Closer) {
	if err := c.Close(); err != nil {
		if OnAutoCloseError != nil {
			OnAutoCloseError(err)
		}
	}
}
