# scan 

[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/blockloop/scan)
[![Travis](https://img.shields.io/travis/blockloop/scan.svg?style=flat-square)](https://travis-ci.org/blockloop/scan)
[![Coveralls github](https://img.shields.io/coveralls/github/blockloop/scan.svg?style=flat-square)](https://coveralls.io/github/blockloop/scan)

scan provides the ability to use database/sql/rows to scan datasets directly to structs or slices. 
For the most comprehensive and up-to-date docs see the [godoc](https://godoc.org/github.com/blockloop/scan)

## Examples

### Multiple Rows
```go
db, err := sql.Open("sqlite3", "database.sqlite")
rows, err := db.Query("SELECT * FROM persons")
var persons []Person
err := scan.Rows(&persons, rows)

fmt.Printf("%#v", persons)
// []Person{
//    {ID: 1, Name: "brett"},
//    {ID: 2, Name: "fred"},
//    {ID: 3, Name: "stacy"},
// }
```
### Multiple rows of primative type

```go
rows, err := db.Query("SELECT name FROM persons")
var names []string
err := scan.Rows(&names, rows)

fmt.Printf("%#v", names)
// []string{
//    "brett",
//    "fred",
//    "stacy",
// }
```

### Single row

```go
rows, err := db.Query("SELECT * FROM persons where name = 'brett' LIMIT 1")
var person Person
err := scan.Row(&person, rows)

fmt.Printf("%#v", person)
// Person{ ID: 1, Name: "brett" }
```

### Scalar value

```go
rows, err := db.Query("SELECT age FROM persons where name = 'brett' LIMIT 1")
var age int8
err := scan.Row(&age, row)

fmt.Printf("%d", age)
// 100
```

## Why

While many other awesome db project support similar features (i.e. [sqlx](https://github.com/jmoiron/sqlx)) this provides
the ability to use other projects like [sq](https://github.com/Masterminds/squirrel) to write fluent sql statements and
pass the resulting `row` to `scan` for simple scanning


## Benchmarks

I created some benchmarks in [scanner_bench_test.go](scanner_bench_test.go) to compare using `scan` against
manually scanning directly to structs and/or appending to slices

```
→ go test -bench=. -benchtime=5s ./
goos: darwin
goarch: amd64
pkg: github.com/blockloop/scan
BenchmarkScanRowOneField-8               1000000              9956 ns/op
BenchmarkDirectScanOneField-8            1000000              9111 ns/op
BenchmarkScanRowFiveFields-8              500000             21125 ns/op
BenchmarkDirectScanFiveFields-8           500000             16446 ns/op
BenchmarkScanRowsOneField-8               500000             17365 ns/op
BenchmarkDirectScanManyOneField-8         500000             13136 ns/op
PASS
ok      github.com/blockloop/scan       53.995s
```
