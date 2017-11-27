# scnr 

[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/blockloop/scnr)
[![Travis](https://img.shields.io/travis/blockloop/scnr.svg?style=flat-square)](https://travis-ci.org/blockloop/scnr)
[![Coveralls github](https://img.shields.io/coveralls/github/blockloop/scnr.svg?style=flat-square)](https://coveralls.io/github/blockloop/scnr)

scnr provides the ability to use database/sql/rows to scan datasets directly to structs or slices. 
For the most comprehensive and up-to-date docs see the [godoc](https://godoc.org/github.com/blockloop/scnr)

## Example

```go
/// multiple rows
///

db, err := sql.Open("sqlite3", ":memory:")
// handle err

rows, err := db.Query("SELECT * FROM persons where name = 'brett'")
// handle err

var persons []Person

err := scnr.Rows(&persons, rows)
// handle err

fmt.Printf("%#v", persons)

/// Single row
///

rows, err := db.Query("SELECT * FROM persons where name = 'brett' LIMIT 1")
// handle err

var person Person

err := scnr.Row(&person, rows)
// handle err

fmt.Printf("%#v", person)

```

## Why

While many other awesome db project support similar features (i.e. [sqlx](https://github.com/jmoiron/sqlx)) this provides
the ability to use other projects like [sq](https://github.com/Masterminds/squirrel) to write fluent sql statements and
pass the resulting `row` to `scnr` for simple scanning

## Scalar

scnr does not have an option to scan scalar values because this is a one liner for the builtin row already provided by go

```go
row := db.QueryRow("SELECT age FROM persons where name = 'brett' LIMIT 1")
// should be one row with one column 'age'
var age int8
row.Scan(&age)
```
