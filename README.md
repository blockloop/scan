# scnr 

[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/blockloop/scnr)
[![Travis](https://img.shields.io/travis/blockloop/scnr.svg?style=flat-square)](https://travis-ci.org/blockloop/scnr)
[![Coveralls github](https://img.shields.io/coveralls/github/blockloop/scnr.svg?style=flat-square)](https://coveralls.io/github/blockloop/scnr)

scnr provides the ability to use database/sql/rows to scan datasets directly to structs or slices. 
For the most comprehensive and up-to-date docs see the [godoc](https://godoc.org/github.com/blockloop/scnr)

## Examples

### Multiple Rows
```go
db, err := sql.Open("sqlite3", "database.sqlite")
rows, err := db.Query("SELECT * FROM persons")
var persons []Person
err := scnr.Slice(&persons, rows)

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
err := scnr.Slice(&names, rows)

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
err := scnr.One(&person, rows)

fmt.Printf("%#v", person)
// Person{ ID: 1, Name: "brett" }
```

### Scalar value

```go
row := db.Query("SELECT age FROM persons where name = 'brett' LIMIT 1")
var age int8
err := scnr.Scalar(&age, row)

fmt.Printf("%d", age)
// 100
```

## Why

While many other awesome db project support similar features (i.e. [sqlx](https://github.com/jmoiron/sqlx)) this provides
the ability to use other projects like [sq](https://github.com/Masterminds/squirrel) to write fluent sql statements and
pass the resulting `row` to `scnr` for simple scanning


## Benchmarks

I created some benchmarks in [scanner_bench_test.go](scanner_bench_test.go) to compare using `scnr` against
manually scanning directly to structs and/or appending to slices

```
→ go test -bench=. -benchtime=5s ./
goos: darwin
goarch: amd64
pkg: github.com/blockloop/scnr
BenchmarkScnrOneOneField-8               1000000              9956 ns/op
BenchmarkDirectScanOneField-8            1000000              9111 ns/op
BenchmarkScnrRowFiveFields-8              500000             21125 ns/op
BenchmarkDirectScanFiveFields-8           500000             16446 ns/op
BenchmarkScnrSliceOneField-8              500000             17365 ns/op
BenchmarkDirectScanManyOneField-8         500000             13136 ns/op
PASS
ok      github.com/blockloop/scnr       53.995s
```
