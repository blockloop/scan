package scan

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
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
)

// Row scans a single row into a single variable. It requires that you use
// db.Query and not db.QueryRow, because QueryRow does not return column names.
// There is no performance impact in using one over the other. QueryRow only
// defers returning err until Scan is called, which is an unnecessary
// optimization for this library.
func Row(v interface{}, r RowsScanner) error {
	return row(v, r, false)
}

// RowStrict scans a single row into a single variable. It is identical to
// Row, but it ignores fields that do not have a db tag
func RowStrict(v interface{}, r RowsScanner) error {
	return row(v, r, true)
}

func row(v interface{}, r RowsScanner, strict bool) error {
	vType := reflect.TypeOf(v)
	if k := vType.Kind(); k != reflect.Ptr {
		return fmt.Errorf("%q must be a pointer", k.String())
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
	return rows(v, r, false)
}

// RowsStrict scans sql rows into a slice (v) only using db tags
func RowsStrict(v interface{}, r RowsScanner) (outerr error) {
	return rows(v, r, true)
}

func rows(v interface{}, r RowsScanner, strict bool) (outerr error) {
	if AutoClose {
		defer closeRows(r)
	}

	vType := reflect.TypeOf(v)
	if k := vType.Kind(); k != reflect.Ptr {
		return fmt.Errorf("%q must be a pointer", k.String())
	}
	sliceType := vType.Elem()
	if reflect.Slice != sliceType.Kind() {
		return fmt.Errorf("%q must be a slice", sliceType.String())
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

// fieldByName gets a struct's field by first looking up the db struct tag and falling
// back to the field's name in Title case.
func fieldByName(v reflect.Value, name string, strict bool) reflect.Value {
	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		tag, ok := typ.Field(i).Tag.Lookup("db")
		if ok && tag == name {
			return v.Field(i)
		}
	}
	if strict {
		return reflect.ValueOf(nil)
	}
	return v.FieldByName(strings.Title(name))
}

func structPointers(stct reflect.Value, cols []string, strict bool) []interface{} {
	pointers := make([]interface{}, 0, len(cols))
	for _, colName := range cols {
		fieldVal := fieldByName(stct, colName, strict)
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
		log.Printf("failed to close rows: %+v\n", err)
	}
}
