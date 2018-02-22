package scan_test

import (
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/blockloop/scan"
	"github.com/blockloop/scan/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRowsConvertsColumnNamesToTitleText(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{"first"}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(true).Times(1)
	rs.EXPECT().Next().Return(false)
	rs.EXPECT().Scan(gomock.Any()).Do(func(v interface{}) {
		set(v, "Brett Jones")
	}).Return(nil)

	rs.EXPECT().Err().Return(nil)

	var item struct {
		First string
	}

	require.NoError(t, scan.Row(&item, rs))
	assert.Equal(t, "Brett Jones", item.First)
}

func TestRowsUsesTagName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{"first_and_last_name"}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(true).Times(1)
	rs.EXPECT().Next().Return(false)
	rs.EXPECT().Scan(gomock.Any()).Do(func(v interface{}) {
		set(v, "Brett Jones")
	}).Return(nil)
	rs.EXPECT().Err().Return(nil)

	var item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}

	require.NoError(t, scan.Row(&item, rs))
	assert.Equal(t, "Brett Jones", item.FirstAndLastName)
}

func TestRowsIgnoresUnsetableColumns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{"first_and_last_name"}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(true).Times(1)
	rs.EXPECT().Next().Return(false)
	rs.EXPECT().Scan(gomock.Any()).Return(nil)
	rs.EXPECT().Err().Return(nil)

	var item struct {
		// private, unsetable
		firstAndLastName string `db:"first_and_last_name"`
	}

	require.NoError(t, scan.Row(&item, rs))
}

func TestErrorsWhenScanErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{"first_and_last_name"}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(true).Times(1)
	scanErr := errors.New("broken")
	rs.EXPECT().Scan(gomock.Any()).Return(scanErr).Times(1)

	var item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}

	err := scan.Row(&item, rs)
	assert.Equal(t, scanErr, err)
}

func TestRowsPanicsWhenNotGivenAPointer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Close().Return(nil)

	assert.Panics(t, func() {
		scan.Rows("hello", rs)
	})
}

func TestRowsPanicsWhenNotGivenAPointerToSlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Close().Return(nil)

	var item struct{}
	assert.Panics(t, func() {
		scan.Rows(&item, rs)
	})
}

func TestErrorsWhenColumnsReturnsError(t *testing.T) {
	columnsErr := errors.New("broken")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return(nil, columnsErr)
	rs.EXPECT().Close().Return(nil)

	var items []struct {
		Name string
		Age  int
	}
	err := scan.Rows(&items, rs)
	assert.Equal(t, columnsErr, err)
}

func TestDoesNothingWhenNoColumns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(true).Times(1)

	var items []struct {
		Name string
		Age  int
	}
	err := scan.Rows(&items, rs)
	assert.NoError(t, err)
	assert.Nil(t, items)
}

func TestDoesNothingWhenNextIsFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{"col_int"}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(false)
	rs.EXPECT().Err().Return(nil)

	var items []struct {
		Name string
		Age  int
	}
	err := scan.Rows(&items, rs)
	assert.NoError(t, err)
	assert.Nil(t, items)
}

func TestIgnoresColumnsThatDoNotHaveFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{"first", "last", "age"}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(true).Times(2)
	rs.EXPECT().Next().Return(false)
	rs.EXPECT().Err().Return(nil)

	rs.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(v1, v2, v3 interface{}) {
		set(v1, "Brett")
		set(v2, "Jones")
	}).Return(nil).Times(1)

	rs.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(v1, v2, v3 interface{}) {
		set(v1, "Fred")
		set(v2, "Jones")
	}).Return(nil).Times(1)

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
}

func TestIgnoresFieldsThatDoNotHaveColumns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{"first", "age"}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(true).Times(2)
	rs.EXPECT().Next().Return(false)
	rs.EXPECT().Err().Return(nil)

	rs.EXPECT().Scan(gomock.Any(), gomock.Any()).Do(func(v1, v2 interface{}) {
		set(v1, "Brett")
		set(v2, int8(100))
	}).Return(nil).Times(1)

	rs.EXPECT().Scan(gomock.Any(), gomock.Any()).Do(func(v1, v2 interface{}) {
		set(v1, "Fred")
		set(v2, int8(100))
	}).Return(nil).Times(1)

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
}

func TestRowScansToPrimitiveType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Close().Return(nil)
	rs.EXPECT().Next().Return(true).Times(1)
	rs.EXPECT().Next().Return(false)
	rs.EXPECT().Columns().Return([]string{"doesn't matter"}, nil)
	rs.EXPECT().Scan(gomock.Any()).Do(func(args ...interface{}) {
		assert.Len(t, args, 1)
	}).Return(nil).Times(1)
	rs.EXPECT().Err().Return(nil)

	var name string
	assert.NoError(t, scan.Row(&name, rs))

}

func TestReturnsScannerError(t *testing.T) {
	scanErr := errors.New("broken")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Err().Return(scanErr)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(false)
	rs.EXPECT().Columns().Return([]string{"name"}, nil)

	var persons []struct {
		Name string
	}

	err := scan.Rows(&persons, rs)
	assert.EqualValues(t, scanErr, err)
}

func TestScansPrimitiveSlices(t *testing.T) {
	table := [][]interface{}{
		{1, 2, 3},
		{"brett", "fred", "geoff"},
		{true, false, true},
		{1.0, 1.1, 1.2},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, items := range table {
		queue := newSimpleQueue(items)

		rs := mocks.NewMockRowsScanner(ctrl)
		rs.EXPECT().Columns().Return([]string{"whatever"}, nil)
		rs.EXPECT().Close().Return(nil).Times(1)
		rs.EXPECT().Next().Return(true).Times(len(items))
		rs.EXPECT().Next().Return(false).AnyTimes()
		rs.EXPECT().Err().Return(nil).AnyTimes()

		rs.EXPECT().Scan(gomock.Any()).Do(func(v interface{}) {
			val, ok := queue.Pop()
			require.True(t, ok, "pop value")
			set(v, val)
		}).Return(nil).AnyTimes()

		var scanned []interface{}

		require.NoError(t, scan.Rows(&scanned, rs))
		assert.EqualValues(t, items, scanned)
	}
}

func TestErrorsWhenMoreThanOneColumnForPrimitiveSlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{"fname", "lname"}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(true).Times(1)

	var fnames []string

	err := scan.Rows(&fnames, rs)
	assert.EqualValues(t, scan.ErrTooManyColumns, err)
}

func TestErrorsWhenScanRowToSlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)

	var persons []struct {
		ID int
	}

	err := scan.Row(&persons, rs)
	assert.EqualValues(t, scan.ErrSliceForRow, err)
}

func TestRowReturnsErrNoRowsWhenQueryHasNoRows(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)
	rs.EXPECT().Columns().Return([]string{"first", "last"}, nil)
	rs.EXPECT().Close().Return(nil).Times(1)
	rs.EXPECT().Next().Return(false)
	rs.EXPECT().Err().Return(nil)

	var item struct {
		First string
	}

	assert.EqualValues(t, sql.ErrNoRows, scan.Row(&item, rs))
}

func TestRowPanicsWhenItemIsNotAPointer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	rs := mocks.NewMockRowsScanner(ctrl)

	var item struct {
		First string
	}

	assert.Panics(t, func() {
		scan.Row(item, rs)
	})
}

func set(ptr, val interface{}) {
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
