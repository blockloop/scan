package scan_test

import (
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/blockloop/scan/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRowsConvertsColumnNamesToTitleText(t *testing.T) {
	var item struct {
		First string
	}

	expected := "Brett Jones"
	rows := fakeRowsWithRecords(t, []string{"First"},
		[]interface{}{expected},
	)

	require.NoError(t, scan.Row(&item, rows))
	assert.Equal(t, 1, rows.ScanCallCount())
	assert.Equal(t, expected, item.First)
}

func TestRowsUsesTagName(t *testing.T) {
	expected := "Brett Jones"
	rows := fakeRowsWithRecords(t, []string{"first_and_last_name"},
		[]interface{}{expected},
	)

	var item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}

	require.NoError(t, scan.Row(&item, rows))
	assert.Equal(t, 1, rows.ScanCallCount())
	assert.Equal(t, expected, item.FirstAndLastName)
}

func TestRowsIgnoresUnsetableColumns(t *testing.T) {
	expected := "Brett Jones"
	rows := fakeRowsWithRecords(t, []string{"first_and_last_name"},
		[]interface{}{expected},
	)

	var item struct {
		// private, unsetable
		firstAndLastName string `db:"first_and_last_name"`
	}

	require.NoError(t, scan.Row(&item, rows))
	assert.NotEqual(t, expected, item.firstAndLastName)
}

func TestErrorsWhenScanErrors(t *testing.T) {
	expected := errors.New("asdf")
	rows := fakeRowsWithColumns(t, 1, "first_and_last_name")
	rows.ScanStub = func(...interface{}) error {
		return expected
	}

	var item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}

	err := scan.Row(&item, rows)
	assert.Equal(t, expected, err)
}

func TestRowsErrorsWhenNotGivenAPointer(t *testing.T) {
	rows := fakeRowsWithColumns(t, 1, "name")

	err := scan.Rows("hello", rows)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pointer")
}

func TestRowsErrorsWhenNotGivenAPointerToSlice(t *testing.T) {
	rows := fakeRowsWithColumns(t, 1, "name")

	var item struct{}
	err := scan.Rows(&item, rows)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slice")
}

func TestErrorsWhenColumnsReturnsError(t *testing.T) {
	expected := errors.New("asdf")
	rows := &FakeRowsScanner{
		ColumnsStub: func() ([]string, error) {
			return nil, expected
		},
	}

	var items []struct {
		Name string
		Age  int
	}
	err := scan.Rows(&items, rows)
	assert.Equal(t, expected, err)
}

func TestDoesNothingWhenNoColumns(t *testing.T) {
	rows := fakeRowsWithColumns(t, 1)

	var items []struct {
		Name string
		Age  int
	}
	err := scan.Rows(&items, rows)
	assert.NoError(t, err)
	assert.Nil(t, items)
}

func TestDoesNothingWhenNextIsFalse(t *testing.T) {
	rows := fakeRowsWithColumns(t, 0, "Name")

	var items []struct {
		Name string
		Age  int
	}
	err := scan.Rows(&items, rows)
	assert.NoError(t, err)
	assert.Nil(t, items)
}

func TestIgnoresColumnsThatDoNotHaveFields(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"First", "Last", "Age"},
		[]interface{}{"Brett", "Jones"},
		[]interface{}{"Fred", "Jones"},
	)

	var items []struct {
		First string
		Last  string
	}

	require.NoError(t, scan.Rows(&items, rows))
	require.Len(t, items, 2)
	assert.Equal(t, "Brett", items[0].First)
	assert.Equal(t, "Jones", items[0].Last)
	assert.Equal(t, "Fred", items[1].First)
	assert.Equal(t, "Jones", items[1].Last)
}

func TestIgnoresFieldsThatDoNotHaveColumns(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"first", "age"},
		[]interface{}{"Brett", int8(40)},
		[]interface{}{"Fred", int8(50)},
	)

	var items []struct {
		First string
		Last  string
		Age   int8
	}

	require.NoError(t, scan.Rows(&items, rows))
	require.Len(t, items, 2)
	assert.EqualValues(t, "Brett", items[0].First)
	assert.EqualValues(t, "", items[0].Last)
	assert.EqualValues(t, 40, items[0].Age)

	assert.EqualValues(t, "Fred", items[1].First)
	assert.EqualValues(t, "", items[1].Last)
	assert.EqualValues(t, 50, items[1].Age)
}

func TestRowScansToPrimitiveType(t *testing.T) {
	expected := "Bob"
	rows := fakeRowsWithRecords(t, []string{"name"},
		[]interface{}{expected},
	)

	var name string
	assert.NoError(t, scan.Row(&name, rows))
	assert.Equal(t, expected, name)
}

func TestReturnsScannerError(t *testing.T) {
	scanErr := errors.New("broken")

	rows := fakeRowsWithColumns(t, 1, "Name")
	rows.ErrReturns(scanErr)

	var persons []struct {
		Name string
	}

	err := scan.Rows(&persons, rows)
	assert.EqualValues(t, scanErr, err)
}

func TestScansPrimitiveSlices(t *testing.T) {
	table := [][]interface{}{
		{1, 2, 3},
		{"brett", "fred", "geoff"},
		{true, false},
		{1.0, 1.1, 1.2},
	}

	for _, items := range table {
		// each item in items is a single value which needs to be converted
		// to a single row with a scalar value
		dbrows := make([][]interface{}, len(items))
		for i, item := range items {
			dbrows[i] = []interface{}{item}
		}
		rows := fakeRowsWithRecords(t, []string{"a"}, dbrows...)

		var scanned []interface{}

		require.NoError(t, scan.Rows(&scanned, rows))
		assert.EqualValues(t, items, scanned)
	}
}

func TestErrorsWhenMoreThanOneColumnForPrimitiveSlice(t *testing.T) {
	rows := fakeRowsWithColumns(t, 1, "fname", "lname")

	var fnames []string

	err := scan.Rows(&fnames, rows)
	assert.EqualValues(t, scan.ErrTooManyColumns, err)
}

func TestErrorsWhenScanRowToSlice(t *testing.T) {
	rows := &FakeRowsScanner{}

	var persons []struct {
		ID int
	}

	err := scan.Row(&persons, rows)
	assert.EqualValues(t, scan.ErrSliceForRow, err)
}

func TestRowReturnsErrNoRowsWhenQueryHasNoRows(t *testing.T) {
	rows := fakeRowsWithColumns(t, 0, "First")

	var item struct {
		First string
	}

	assert.EqualValues(t, sql.ErrNoRows, scan.Row(&item, rows))
}

func TestRowErrorsWhenItemIsNotAPointer(t *testing.T) {
	rows := &FakeRowsScanner{}

	var item struct {
		First string
	}

	err := scan.Row(item, rows)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pointer")
}

func TestRowStrictIgnoresFieldsWithoutDBTag(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"First", "Last"},
		[]interface{}{"Brett", "Jones"},
	)

	var item struct {
		First string `db:"First"`
		Last  string
	}

	require.NoError(t, scan.RowStrict(&item, rows))
	assert.Equal(t, "Brett", item.First)
	assert.Equal(t, "", item.Last)
}

func TestRowScansNestedFields(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"p.First", "p.Last"},
		[]interface{}{"Brett", "Jones"},
	)

	var res struct {
		Item struct {
			First string `db:"p.First"`
			Last  string `db:"p.Last"`
		}
	}

	require.NoError(t, scan.Row(&res, rows))
	assert.Equal(t, "Brett", res.Item.First)
	assert.Equal(t, "Jones", res.Item.Last)
}

func TestRowStrictScansNestedFields(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"p.First", "p.Last"},
		[]interface{}{"Brett", "Jones"},
	)

	var res struct {
		Item struct {
			First string `db:"p.First"`
			Last  string `db:"p.Last"`
		}
	}

	require.NoError(t, scan.RowStrict(&res, rows))
	assert.Equal(t, "Brett", res.Item.First)
	assert.Equal(t, "Jones", res.Item.Last)
}

func TestRowsStrictIgnoresFieldsWithoutDBTag(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"First", "Last"},
		[]interface{}{"Brett", "Jones"},
		[]interface{}{"Fred", "Jones"},
	)

	var items []struct {
		First string `db:"First"`
		Last  string
	}

	require.NoError(t, scan.RowsStrict(&items, rows))
	require.Len(t, items, 2)
	assert.Equal(t, "Brett", items[0].First)
	assert.Equal(t, "", items[0].Last)
	assert.Equal(t, "Fred", items[1].First)
	assert.Equal(t, "", items[1].Last)
}

func TestRowCloses(t *testing.T) {
	rows := fakeRowsWithRecords(t, []string{"name"},
		[]interface{}{"Bob"},
	)

	var name string
	assert.NoError(t, scan.Row(&name, rows))
	assert.EqualValues(t, 1, rows.CloseCallCount())
}

func Test_OnAutoCloseErrorIsCalledWhenRowsCloseErrors(t *testing.T) {
	expected := sql.ErrTxDone
	calls := int32(0)

	scan.OnAutoCloseError = func(err error) {
		assert.Equal(t, expected, err)
		atomic.AddInt32(&calls, 1)
	}

	rows := fakeRowsWithRecords(t, []string{"name"},
		[]interface{}{"Bob"},
	)

	rows.CloseReturns(expected)

	var name string
	assert.NoError(t, scan.Row(&name, rows))
	assert.EqualValues(t, 1, calls)
}

func setValue(ptr, val interface{}) {
	reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(val))
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
