package scan_test

import (
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/blockloop/scan"
	"github.com/blockloop/scan/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRowsConvertsColumnNamesToTitleText(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first"}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(true).Once()
	rs.On("Next").Return(false)
	rs.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 1)
		v := args.Get(0)
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf("Brett Jones"))
	}).Return(nil)

	rs.On("Err").Return(nil)

	var item struct {
		First string
	}

	require.NoError(t, scan.Row(&item, rs))
	assert.Equal(t, "Brett Jones", item.First)
	rs.AssertExpectations(t)
}

func TestRowsUsesTagName(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first_and_last_name"}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(true).Once()
	rs.On("Next").Return(false)
	rs.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 1)
		v := args.Get(0)
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf("Brett Jones"))
	}).Return(nil)
	rs.On("Err").Return(nil)

	var item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}

	require.NoError(t, scan.Row(&item, rs))
	assert.Equal(t, "Brett Jones", item.FirstAndLastName)
	rs.AssertExpectations(t)
}

func TestRowsIgnoresUnsetableColumns(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first_and_last_name"}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(true).Once()
	rs.On("Next").Return(false)
	rs.On("Scan", mock.Anything).Return(nil)
	rs.On("Err").Return(nil)

	var item struct {
		// private, unsetable
		firstAndLastName string `db:"first_and_last_name"`
	}

	require.NoError(t, scan.Row(&item, rs))
	rs.AssertExpectations(t)
}

func TestErrorsWhenScanErrors(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first_and_last_name"}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(true).Once()
	rs.On("Next").Return(false)
	scanErr := errors.New("broken")
	rs.On("Scan", mock.Anything).Return(scanErr).Once()

	var item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}

	err := scan.Row(&item, rs)
	assert.Equal(t, scanErr, err)
	rs.AssertExpectations(t)
}

func TestRowsPanicsWhenNotGivenAPointer(t *testing.T) {
	rs := &mocks.RowsScanner{}

	assert.Panics(t, func() {
		scan.Rows("hello", rs)
	})
}

func TestRowsPanicsWhenNotGivenAPointerToSlice(t *testing.T) {
	rs := &mocks.RowsScanner{}

	var item struct{}
	assert.Panics(t, func() {
		scan.Rows(&item, rs)
	})
}

func TestErrorsWhenColumnsReturnsError(t *testing.T) {
	columnsErr := errors.New("broken")
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return(nil, columnsErr)
	rs.On("Close").Return(nil)

	var items []struct {
		Name string
		Age  int
	}
	err := scan.Rows(&items, rs)
	assert.Equal(t, columnsErr, err)
	rs.AssertExpectations(t)
}

func TestDoesNothingWhenNoColumns(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(true).Once()

	var items []struct {
		Name string
		Age  int
	}
	err := scan.Rows(&items, rs)
	assert.NoError(t, err)
	assert.Nil(t, items)
	rs.AssertExpectations(t)
}

func TestDoesNothingWhenNextIsFalse(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"col_int"}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(false)
	rs.On("Err").Return(nil)

	var items []struct {
		Name string
		Age  int
	}
	err := scan.Rows(&items, rs)
	assert.NoError(t, err)
	assert.Nil(t, items)
	rs.AssertExpectations(t)
}

func TestIgnoresColumnsThatDoNotHaveFields(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first", "last", "age"}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(true).Twice()
	rs.On("Next").Return(false)
	rs.On("Err").Return(nil)

	rs.On("Scan", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		reflect.ValueOf(args.Get(0)).Elem().Set(reflect.ValueOf("Brett"))
		reflect.ValueOf(args.Get(1)).Elem().Set(reflect.ValueOf("Jones"))
	}).Return(nil).Once()

	rs.On("Scan", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		reflect.ValueOf(args.Get(0)).Elem().Set(reflect.ValueOf("Fred"))
		reflect.ValueOf(args.Get(1)).Elem().Set(reflect.ValueOf("Jones"))
	}).Return(nil).Once()

	var items []struct {
		First string
		Last  string
	}

	require.NoError(t, scan.Rows(&items, rs))
	require.Len(t, items, 2)
	assert.Equal(t, "Brett", items[0].First)
	assert.Equal(t, "Jones", items[0].Last)
	assert.Equal(t, "Fred", items[1].First)
	assert.Equal(t, "Jones", items[1].Last)
	rs.AssertExpectations(t)
}

func TestIgnoresFieldsThatDoNotHaveColumns(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first", "age"}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(true).Twice()
	rs.On("Next").Return(false)
	rs.On("Err").Return(nil)

	rs.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 2)
		reflect.ValueOf(args.Get(0)).Elem().Set(reflect.ValueOf("Brett"))
		reflect.ValueOf(args.Get(1)).Elem().Set(reflect.ValueOf(int8(100)))
	}).Return(nil).Once()

	rs.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 2)
		reflect.ValueOf(args.Get(0)).Elem().Set(reflect.ValueOf("Fred"))
		reflect.ValueOf(args.Get(1)).Elem().Set(reflect.ValueOf(int8(100)))
	}).Return(nil).Once()

	var items []struct {
		First string
		Last  string
		Age   int8
	}

	require.NoError(t, scan.Rows(&items, rs))
	require.Len(t, items, 2)
	assert.EqualValues(t, "Brett", items[0].First)
	assert.EqualValues(t, "", items[0].Last)
	assert.EqualValues(t, 100, items[0].Age)

	assert.EqualValues(t, "Fred", items[1].First)
	assert.EqualValues(t, "", items[1].Last)
	assert.EqualValues(t, 100, items[1].Age)
	rs.AssertExpectations(t)
}

func TestRowScansToPrimativeType(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Close").Return(nil)
	rs.On("Next").Return(true).Once()
	rs.On("Next").Return(false)
	rs.On("Columns").Return([]string{"doesn't matter"}, nil)
	rs.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 1)
	}).Return(nil).Once()
	rs.On("Err").Return(nil)

	var name string
	assert.NoError(t, scan.Row(&name, rs))

	rs.AssertExpectations(t)
}

func TestReturnsScannerError(t *testing.T) {
	scanErr := errors.New("broken")

	rs := &mocks.RowsScanner{}
	rs.On("Err").Return(scanErr)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(false)
	rs.On("Columns").Return([]string{"name"}, nil)

	var persons []struct {
		Name string
	}

	err := scan.Rows(&persons, rs)
	assert.EqualValues(t, scanErr, err)
	rs.AssertExpectations(t)
}

func TestScansPrimitiveSlices(t *testing.T) {
	table := [][]interface{}{
		[]interface{}{1, 2, 3},
		[]interface{}{"brett", "fred", "geoff"},
		[]interface{}{true, false, true},
		[]interface{}{1.0, 1.1, 1.2},
	}

	for _, items := range table {
		queue := newSimpleQueue(items)

		rs := &mocks.RowsScanner{}
		rs.On("Columns").Return([]string{"whatever"}, nil)
		rs.On("Close").Return(nil).Once()
		rs.On("Next").Return(true).Times(len(items))
		rs.On("Next").Return(false)
		rs.On("Err").Return(nil)

		rs.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			v, ok := queue.Pop()
			require.True(t, ok, "pop value")
			reflect.ValueOf(args.Get(0)).Elem().Set(reflect.ValueOf(v))
		}).Return(nil)

		var scanned []interface{}

		require.NoError(t, scan.Rows(&scanned, rs))
		assert.EqualValues(t, items, scanned)
		rs.AssertExpectations(t)
	}
}

func TestErrorsWhenMoreThanOneColumnForPrimitiveSlice(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"fname", "lname"}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(true).Once()

	var fnames []string

	err := scan.Rows(&fnames, rs)
	assert.EqualValues(t, scan.ErrTooManyColumns, err)
	rs.AssertExpectations(t)
}

func TestErrorsWhenScanRowToSlice(t *testing.T) {
	rs := &mocks.RowsScanner{}

	var persons []struct {
		ID int
	}

	err := scan.Row(&persons, rs)
	assert.EqualValues(t, scan.ErrSliceForRow, err)
	rs.AssertExpectations(t)
}

func TestRowReturnsErrNoRowsWhenQueryHasNoRows(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first", "last"}, nil)
	rs.On("Close").Return(nil).Once()
	rs.On("Next").Return(false)
	rs.On("Err").Return(nil)

	var item struct {
		First string
	}

	assert.EqualValues(t, sql.ErrNoRows, scan.Row(&item, rs))
	rs.AssertExpectations(t)
}

func TestRowPanicsWhenItemIsNotAPointer(t *testing.T) {
	rs := &mocks.RowsScanner{}

	var item struct {
		First string
	}

	assert.Panics(t, func() {
		scan.Row(item, rs)
	})
	rs.AssertExpectations(t)
}

func TestColumnsListsAllFields(t *testing.T) {
	var person struct {
		FirstName string
		Age       int
		Active    bool
	}

	cols := scan.Columns(&person)
	assert.EqualValues(t, []string{"FirstName", "Age", "Active"}, cols)
}

func TestColumnsPrefersDBTag(t *testing.T) {
	var person struct {
		FirstName string `db:"first_name"`
		Age       int
		Active    bool
	}

	cols := scan.Columns(&person)
	assert.EqualValues(t, []string{"first_name", "Age", "Active"}, cols)
}

func TestColumnsSkipsDashDBTag(t *testing.T) {
	var person struct {
		FirstName string `db:"-"`
		Age       int
		Active    bool
	}

	cols := scan.Columns(&person)
	assert.EqualValues(t, []string{"Age", "Active"}, cols)
}

func TestColumnsSkipsComplexTypeFields(t *testing.T) {
	var person struct {
		FirstName string
		Age       int
		Active    bool
		Address   struct {
			Line1 string
			City  string
			State string
		}
		One   map[string]interface{}
		Two   []string
		Three chan bool
	}

	cols := scan.Columns(&person)
	assert.EqualValues(t, []string{"FirstName", "Age", "Active"}, cols)
}

func TestColumnsUsesValueWhenNotAPointer(t *testing.T) {
	var person struct {
		FirstName *string
	}

	cols := scan.Columns(person)
	assert.EqualValues(t, []string{"FirstName"}, cols)
}

func TestColumnsPanicsWhenNotAStruct(t *testing.T) {
	var a string

	assert.Panics(t, func() {
		scan.Columns(a)
	})
}

func TestValuesGetsAllFieldValues(t *testing.T) {
	type person struct {
		Name string
		Age  int
	}

	p := &person{
		Name: "brett",
		Age:  100,
	}

	vals := scan.Values(p)
	assert.EqualValues(t, []interface{}{"brett", 100}, vals)
}

func TestValuesSkipsComplexTypes(t *testing.T) {
	type person struct {
		FirstName string
		Age       int
		Active    bool
		Address   struct {
			Line1 string
			City  string
			State string
		}
		One   map[string]interface{}
		Two   []string
		Three chan bool
	}

	p := &person{
		FirstName: "brett",
		Age:       100,
		Active:    false,
		One:       map[string]interface{}{"Hello": "world"},
		Two:       []string{"abcd"},
		Three:     make(chan bool, 0),
	}

	vals := scan.Values(p)
	assert.EqualValues(t, []interface{}{"brett", 100, false}, vals)
}

func TestValuesSkipsDashDBTag(t *testing.T) {
	type person struct {
		FirstName string `db:"-"`
		Age       int
		Active    bool
	}
	p := &person{
		FirstName: "brett",
		Age:       100,
		Active:    true,
	}

	vals := scan.Values(p)
	assert.EqualValues(t, []interface{}{100, true}, vals)
}

func TestValuesUsesNilForEmptyValues(t *testing.T) {
	type person struct {
		FirstName *string
	}
	p := &person{}

	vals := scan.Values(p)
	assert.EqualValues(t, []interface{}{(*string)(nil)}, vals)
}

func TestValuesWorksWhenNotAPointer(t *testing.T) {
	type person struct {
		FirstName string
	}
	p := person{FirstName: "brett"}

	vals := scan.Values(p)
	assert.EqualValues(t, []interface{}{"brett"}, vals)
}

func TestValuesPanicsWhenNotAStruct(t *testing.T) {
	assert.Panics(t, func() {
		scan.Values("brett")
	})
}

type simpleQueue struct {
	items []interface{}
	m     *sync.Mutex
}

func newSimpleQueue(items []interface{}) *simpleQueue {
	return &simpleQueue{
		items: items,
		m:     &sync.Mutex{},
	}
}

func (q *simpleQueue) Push(v interface{}) {
	q.m.Lock()
	defer q.m.Unlock()
	q.items = append([]interface{}{v}, q.items...)
}

func (q *simpleQueue) Pop() (v interface{}, ok bool) {
	q.m.Lock()
	defer q.m.Unlock()
	if len(q.items) == 0 {
		return nil, false
	}

	v = q.items[0]
	q.items = q.items[1:]
	return v, true
}
