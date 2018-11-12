package scan_test

import (
	"testing"

	"github.com/blockloop/scan"
	"github.com/stretchr/testify/assert"
)

func BenchmarkScanRowOneField(b *testing.B) {
	var item struct {
		First string
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := fakeRowsWithColumns(b, 1, "First")
		if err := scan.Row(&item, rows); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkScanRowFiveFields(b *testing.B) {
	var item struct {
		First  string `db:"first"`
		Age    int8   `db:"age"`
		Active bool   `db:"active"`
		City   string `db:"city"`
		State  string `db:"state"`
	}
	cols, err := scan.Columns(&item)
	assert.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := fakeRowsWithColumns(b, 1, cols...)
		if err := scan.Row(&item, rows); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkScanTenRowsOneField(b *testing.B) {
	type item struct {
		First string `db:"First"`
	}
	var items []item
	cols, err := scan.Columns(&item{})
	assert.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := fakeRowsWithColumns(b, 10, cols...)
		if err := scan.Rows(&items, rows); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkScanTenRowsTenFields(b *testing.B) {
	type item struct {
		One   string `db:"one"`
		Two   string `db:"two"`
		Three int8   `db:"three"`
		Four  bool   `db:"four"`
		Five  string `db:"five"`
		Six   string `db:"six"`
		Seven string `db:"seven"`
		Eight string `db:"eight"`
		Nine  string `db:"nine"`
		Ten   string `db:"ten"`
	}
	var items []item
	cols, err := scan.Columns(&item{})
	assert.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows := fakeRowsWithColumns(b, 10, cols...)
		if err := scan.Rows(&items, rows); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
