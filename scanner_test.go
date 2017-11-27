package scnr_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/blockloop/scnr"
	"github.com/blockloop/scnr/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRowsConvertsColumnNamesToTitleText(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first"}, nil)
	rs.On("Next").Return(true).Times(1)
	rs.On("Next").Return(false)
	rs.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 1)
		v := args.Get(0)
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf("Brett Jones"))
	}).Return(nil)

	var item struct {
		First string
	}

	require.NoError(t, scnr.Row(&item, rs))
	assert.Equal(t, "Brett Jones", item.First)
}

func TestRowsUsesTagName(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first_and_last_name"}, nil)
	rs.On("Next").Return(true).Times(1)
	rs.On("Next").Return(false)
	rs.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 1)
		v := args.Get(0)
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf("Brett Jones"))
	}).Return(nil)

	var item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}

	require.NoError(t, scnr.Row(&item, rs))
	assert.Equal(t, "Brett Jones", item.FirstAndLastName)
}

func TestRowsIgnoresUnsetableColumns(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first_and_last_name"}, nil)
	rs.On("Next").Return(true).Times(1)
	rs.On("Next").Return(false)
	rs.On("Scan", mock.Anything).Return(nil)

	var item struct {
		// private, unsetable
		firstAndLastName string `db:"first_and_last_name"`
	}

	require.NoError(t, scnr.Row(&item, rs))
	rs.AssertNotCalled(t, "Scan", mock.Anything)
}

func TestErrorsWhenScanErrors(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first_and_last_name"}, nil)
	rs.On("Next").Return(true).Times(1)
	rs.On("Next").Return(false)
	scanErr := errors.New("broken")
	rs.On("Scan", mock.Anything).Return(scanErr)

	var item struct {
		FirstAndLastName string `db:"first_and_last_name"`
	}

	err := scnr.Row(&item, rs)
	assert.Equal(t, scanErr, err)
	rs.AssertCalled(t, "Scan", mock.Anything)
}

func TestRowsPanicsWhenNotGivenAPointer(t *testing.T) {
	assert.Panics(t, func() {
		scnr.Rows("hello", nil)
	})
}

func TestRowsPanicsWhenNotGivenAPointerToSlice(t *testing.T) {
	var item struct{}
	assert.Panics(t, func() {
		scnr.Rows(&item, nil)
	})
}

func TestErrorsWhenColumnsReturnsError(t *testing.T) {
	columnsErr := errors.New("broken")
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return(nil, columnsErr)

	var items []struct {
		Name string
		Age  int
	}
	err := scnr.Rows(&items, rs)
	assert.Equal(t, columnsErr, err)
	rs.AssertCalled(t, "Columns")
}

func TestDoesNothingWhenNextIsFalse(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"col_int"}, nil)
	rs.On("Next").Return(false)

	var items []struct {
		Name string
		Age  int
	}
	err := scnr.Rows(&items, rs)
	assert.NoError(t, err)
	assert.Nil(t, items)
}

func TestIgnoresColumnsThatDoNotHaveFields(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first", "last", "age"}, nil)
	rs.On("Next").Return(true).Times(2)
	rs.On("Next").Return(false)

	rs.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 2)
		reflect.ValueOf(args.Get(0)).Elem().Set(reflect.ValueOf("Brett"))
		reflect.ValueOf(args.Get(1)).Elem().Set(reflect.ValueOf("Jones"))
	}).Return(nil).Times(1)

	rs.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 2)
		reflect.ValueOf(args.Get(0)).Elem().Set(reflect.ValueOf("Fred"))
		reflect.ValueOf(args.Get(1)).Elem().Set(reflect.ValueOf("Jones"))
	}).Return(nil).Times(1)

	var items []struct {
		First string
		Last  string
	}

	require.NoError(t, scnr.Rows(&items, rs))
	require.Len(t, items, 2)
	assert.Equal(t, "Brett", items[0].First)
	assert.Equal(t, "Jones", items[0].Last)
	assert.Equal(t, "Fred", items[1].First)
	assert.Equal(t, "Jones", items[1].Last)
}

func TestIgnoresFieldsThatDoNotHaveColumns(t *testing.T) {
	rs := &mocks.RowsScanner{}
	rs.On("Columns").Return([]string{"first", "age"}, nil)
	rs.On("Next").Return(true).Times(2)
	rs.On("Next").Return(false)

	rs.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 2)
		reflect.ValueOf(args.Get(0)).Elem().Set(reflect.ValueOf("Brett"))
		reflect.ValueOf(args.Get(1)).Elem().Set(reflect.ValueOf(int8(100)))
	}).Return(nil).Times(1)

	rs.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		assert.Len(t, args, 2)
		reflect.ValueOf(args.Get(0)).Elem().Set(reflect.ValueOf("Fred"))
		reflect.ValueOf(args.Get(1)).Elem().Set(reflect.ValueOf(int8(100)))
	}).Return(nil).Times(1)

	var items []struct {
		First string
		Last  string
		Age   int8
	}

	require.NoError(t, scnr.Rows(&items, rs))
	require.Len(t, items, 2)
	assert.EqualValues(t, "Brett", items[0].First)
	assert.EqualValues(t, "", items[0].Last)
	assert.EqualValues(t, 100, items[0].Age)

	assert.EqualValues(t, "Fred", items[1].First)
	assert.EqualValues(t, "", items[1].Last)
	assert.EqualValues(t, 100, items[1].Age)
}
