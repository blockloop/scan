package scnr

import (
	"database/sql"
	"reflect"
	"strings"
)

// Row scans a single row into a single variable
func Row(v interface{}, rows RowsScanner) error {
	vType := reflect.TypeOf(v).Elem()
	vVal := reflect.ValueOf(v).Elem()

	sl := reflect.New(reflect.SliceOf(vType))
	err := Rows(sl.Interface(), rows)
	if err != nil {
		return err
	}

	sl = sl.Elem()

	if sl.Len() > 0 {
		vVal.Set(sl.Index(0))
	}

	return nil
}

// Rows scans sql rows into a slice (v)
func Rows(v interface{}, rows RowsScanner) error {
	vType := reflect.TypeOf(v)
	if k := vType.Kind(); k != reflect.Ptr {
		panic(k.String() + ": must be a pointer to a slice")
	}
	sliceType := vType.Elem()
	if reflect.Slice != sliceType.Kind() {
		panic(sliceType.String() + ": must be a pointer to a slice")
	}

	sliceVal := reflect.Indirect(reflect.ValueOf(v))
	itemType := sliceType.Elem()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		sliceItem := reflect.New(itemType).Elem()

		pointers := make([]interface{}, 0, len(cols))
		for _, colName := range cols {
			fieldVal := fieldByName(sliceItem, itemType, colName)
			if !fieldVal.IsValid() || !fieldVal.CanSet() {
				continue
			}

			pointers = append(pointers, fieldVal.Addr().Interface())
		}

		if len(pointers) == 0 {
			return nil
		}

		err := rows.Scan(pointers...)
		if err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, sliceItem))
	}
	return nil
}

func fieldByName(v reflect.Value, elemType reflect.Type, name string) reflect.Value {
	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		tag, ok := typ.Field(i).Tag.Lookup("db")
		if ok && tag == name {
			return v.Field(i)
		}
	}
	return v.FieldByName(strings.Title(name))
}

// RowsScanner is a database scanner for many rows
type RowsScanner interface {
	Scan(dest ...interface{}) error
	Close() error
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
	Err() error
	Next() bool
	NextResultSet() bool
}
