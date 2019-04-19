# Scan

[![GoDoc](https://godoc.org/github.com/blockloop/scan?status.svg)](https://godoc.org/github.com/blockloop/scan)
[![Travis](https://img.shields.io/travis/blockloop/scan.svg)](https://travis-ci.org/blockloop/scan)
[![Coveralls github](https://img.shields.io/coveralls/github/blockloop/scan.svg)](https://coveralls.io/github/blockloop/scan)
[![Report Card](https://goreportcard.com/badge/github.com/blockloop/scan)](https://goreportcard.com/report/github.com/blockloop/scan)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=blockloop/scan)](https://dependabot.com)


Scan provides the ability to use database/sql/rows to scan datasets directly to structs or slices. 
For the most comprehensive and up-to-date docs see the [godoc](https://godoc.org/github.com/blockloop/scan)

## Examples

### Multiple Rows
```go
db, _ := sql.Open("sqlite3", "database.sqlite")
rows, _ := db.Query("SELECT * FROM persons")

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
rows, _ := db.Query("SELECT name FROM persons")
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
rows, _ := db.Query("SELECT * FROM persons where name = 'brett' LIMIT 1")
var person Person
err := scan.Row(&person, rows)

fmt.Printf("%#v", person)
// Person{ ID: 1, Name: "brett" }
```

### Scalar value

```go
rows, _ := db.Query("SELECT age FROM persons where name = 'brett' LIMIT 1")
var age int8
err := scan.Row(&age, row)

fmt.Printf("%d", age)
// 100
```

### Strict Scanning

Both `Rows` and `Row` have strict alternatives to allow scanning to structs _strictly_ based on their `db` tag.
To avoid unwanted behavior you can use `RowsStrict` or `RowStrict` to scan without using field names.
Any fields not tagged with the `db` tag will be ignored even if columns are found that match the field names.

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
sq.Insert(userCols...).
        Into("users").
        Values(scan.Values(userCols, &user)...)
```

## Configuration

AutoClose: Automatically call `rows.Close()` after scan completes (default true)

## Why

While many other projects support similar features (i.e. [sqlx](https://github.com/jmoiron/sqlx)) scan allows you to use any database lib such as the stdlib or [squirrel][sq] to write fluent SQL statements and pass the resulting `rows` to `scan` for scanning.

## Benchmarks 

```
λ go test -bench=. -benchtime=10s ./...
goos: linux
goarch: amd64
pkg: github.com/blockloop/scan
BenchmarkColumnsLargeStruct-8           50000000               272 ns/op
BenchmarkValuesLargeStruct-8             2000000              8611 ns/op
BenchmarkScanRowOneField-8               2000000              8528 ns/op
BenchmarkScanRowFiveFields-8             1000000             12234 ns/op
BenchmarkScanTenRowsOneField-8           1000000             16802 ns/op
BenchmarkScanTenRowsTenFields-8           100000            104587 ns/op
PASS
ok      github.com/blockloop/scan       116.055s
```


[sq]: https://github.com/Masterminds/squirrel	"Squirrel"
