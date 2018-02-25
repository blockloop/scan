# scan 

[![GoDoc](https://godoc.org/github.com/blockloop/scan?status.svg)](https://godoc.org/github.com/blockloop/scan)
[![Travis](https://img.shields.io/travis/blockloop/scan.svg)](https://travis-ci.org/blockloop/scan)
[![Coveralls github](https://img.shields.io/coveralls/github/blockloop/scan.svg)](https://coveralls.io/github/blockloop/scan)
[![Report Card](https://goreportcard.com/badge/github.com/blockloop/scan)](https://goreportcard.com/report/github.com/blockloop/scan)

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
### Multiple rows of primitive type

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

### Columns

`Columns` scans a struct and returns a string slice of the assumed column names based on the `db` tag or the struct field name respectively. To avoid assumptions, use `ColumnsStrict` which will _only_ return the fields tagged with the `db` tag. Both `Columns` and `ColumnsStrict` are variadic. They both accept a string slice of column names to exclude from the list. It is recommended that you cache this slice.

```go
package main

type User struct {
        ID        int64
        Name      string
        Age       int
        BirthDate string `db:"bday"`
        Zipcode   string `db:"-"`
        Store     struct {
                ID int
                // ...
        }
}

var nobody = new(User)
var userInsertCols = scan.Columns(nobody, "ID")
// []string{ "Name", "Age", "bday" }

var userSelectCols = scan.Columns(nobody)
// []string{ "ID", "Name", "Age", "bday" }
```

### Values

`Values` scans a struct and returns the values associated with the provided columns. Values uses a `sync.Map` to cache fields of structs to greatly improve the performance of scanning types. The first time a struct is scanned it's **exported** fields locations are cached. When later retrieving values from the same struct it should be much faster. See [Benchmarks](#Benchmarks) below.

```go
user := &User{
        ID: 1,
        Name: "Brett",
        Age: 100,
}

vals := scan.Values([]string{"ID", "Name"}, user)
// []interface{}{ 1, "Brett" }
```

I find that the usefulness of both Values and Columns lies within using a library such as [sq][].

```go
sq.Insert(userCols).
        Into("users").
        Values(scan.Values(userCols, &user))
```


## Why

While many other awesome db project support similar features (i.e. [sqlx](https://github.com/jmoiron/sqlx)) this provides the ability to use other projects like [sq][] to write fluent sql statements and pass the resulting `row` to `scan` for simple scanning


## Benchmarks

I created some benchmarks in [bench_scanner_test.go](bench_scanner_test.go) to compare using `scan` against manually scanning directly to structs and/or appending to slices. The results aren't staggering as you can see. Roughly 850ns for one field structs and 4.6μs for five field structs.

```
> go test -benchtime=10s -bench=.
goos: darwin
goarch: amd64
pkg: github.com/blockloop/scan
BenchmarkValuesLargeStruct-8             2000000              8088 ns/op
BenchmarkValuesLargeStructCached-8      10000000              1564 ns/op
BenchmarkScanRowOneField-8               1000000             12496 ns/op
BenchmarkDirectScanOneField-8            2000000              9448 ns/op
BenchmarkScanRowFiveFields-8              500000             23949 ns/op
BenchmarkDirectScanFiveFields-8          1000000             16993 ns/op
BenchmarkScanRowsOneField-8              1000000             16872 ns/op
BenchmarkDirectScanManyOneField-8        1000000             14030 ns/op
PASS
ok      github.com/blockloop/scan       125.432s
```


[sq]: https://github.com/Masterminds/squirrel	"Squirrel"
