package integration_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/blockloop/scnr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func makeDBSchema(t *testing.T, schema string) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestScanOneScansSingleItem(t *testing.T) {
	type Item struct {
		Name string
		Age  int8
	}

	schema := `CREATE TABLE IF NOT EXISTS persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name, age) VALUES ('brett', 100)
	`
	db := makeDBSchema(t, schema)
	row, err := db.Query(`SELECT * FROM persons LIMIT 1`)
	require.NoError(t, err, "execute query")

	var item Item
	require.NoError(t, scnr.One(&item, row))
	assert.EqualValues(t, "brett", item.Name)
	assert.EqualValues(t, 100, item.Age)
}

func TestScanOneScansSingleItemWithTags(t *testing.T) {
	type Item struct {
		MyName string `db:"name"`
		MyAge  int8   `db:"age"`
	}

	schema := `CREATE TABLE IF NOT EXISTS persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name, age) VALUES ('brett', 100)
	`
	db := makeDBSchema(t, schema)
	row, err := db.Query(`SELECT * FROM persons LIMIT 1`)
	require.NoError(t, err, "execute query")

	var item Item
	require.NoError(t, scnr.One(&item, row))
	assert.EqualValues(t, "brett", item.MyName)
	assert.EqualValues(t, 100, item.MyAge)
}

func TestScanOneScansMultipleItems(t *testing.T) {
	type Item struct {
		Name string
		Age  int8
	}

	schema := `CREATE TABLE IF NOT EXISTS persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name, age) VALUES ('brett', 100), ('jones', 100)
	`
	db := makeDBSchema(t, schema)
	row, err := db.Query(`SELECT * FROM persons ORDER BY name ASC`)
	require.NoError(t, err, "execute query")

	var items []Item
	require.NoError(t, scnr.Slice(&items, row), "Slice")
	require.NotNil(t, items)
	assert.EqualValues(t, "brett", items[0].Name)
	assert.EqualValues(t, 100, items[0].Age)
	assert.EqualValues(t, "jones", items[1].Name)
	assert.EqualValues(t, 100, items[1].Age)
}

func TestScanOneScansMultipleItemsWithTags(t *testing.T) {
	type Item struct {
		MyName string `db:"name"`
		MyAge  int8   `db:"age"`
	}

	schema := `CREATE TABLE IF NOT EXISTS persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name, age) VALUES ('brett', 100), ('jones', 100);
	`
	db := makeDBSchema(t, schema)
	row, err := db.Query(`SELECT * FROM persons ORDER BY name ASC`)
	require.NoError(t, err, "execute query")

	var items []Item
	require.NoError(t, scnr.Slice(&items, row), "Slice")
	require.Len(t, items, 2)
	assert.EqualValues(t, "brett", items[0].MyName)
	assert.EqualValues(t, 100, items[0].MyAge)
	assert.EqualValues(t, "jones", items[1].MyName)
	assert.EqualValues(t, 100, items[1].MyAge)
}

func TestScanOneScansPrimitiveTypesStrings(t *testing.T) {
	schema := `CREATE TABLE IF NOT EXISTS persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name, age) VALUES ('brett', 100), ('jones', 100);
	`
	db := makeDBSchema(t, schema)
	row, err := db.Query(`SELECT name FROM persons ORDER BY name ASC`)
	require.NoError(t, err, "execute query")

	var items []string
	assert.NoError(t, scnr.Slice(&items, row))
	assert.EqualValues(t, []string{"brett", "jones"}, items)
}

func TestScanOneScansPrimitiveTypesInts(t *testing.T) {
	schema := `CREATE TABLE IF NOT EXISTS persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name, age) VALUES ('brett', 100), ('jones', 100);
	`
	db := makeDBSchema(t, schema)
	row, err := db.Query(`SELECT age FROM persons ORDER BY name ASC`)
	require.NoError(t, err, "execute query")

	var items []int
	assert.NoError(t, scnr.Slice(&items, row))
	assert.EqualValues(t, []int{100, 100}, items)
}

func TestScanOneScansPrimitiveTypesInterface(t *testing.T) {
	schema := `CREATE TABLE IF NOT EXISTS persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name, age) VALUES ('brett', 100), ('jones', 100);
	`
	db := makeDBSchema(t, schema)
	row, err := db.Query(`SELECT age FROM persons ORDER BY name ASC`)
	require.NoError(t, err, "execute query")

	var items []interface{}
	assert.NoError(t, scnr.Slice(&items, row))
	// int64 is what Scan uses by default for numbers
	assert.Equal(t, []interface{}{int64(100), int64(100)}, items)
}

func TestScanOneScansWhenMoreColumnsThanProperties(t *testing.T) {
	schema := `CREATE TABLE IF NOT EXISTS persons (
		name VARCHAR(120),
		age TINYINT
	);

	INSERT INTO PERSONS (name, age) VALUES ('brett', 100), ('jones', 100);
	`
	db := makeDBSchema(t, schema)
	row, err := db.Query(`SELECT * FROM persons ORDER BY name ASC`)
	require.NoError(t, err, "execute query")

	type Item struct {
		Name string `db:"name"`
	}

	var items []Item

	assert.NoError(t, scnr.Slice(&items, row))
	assert.EqualValues(t, []Item{
		{Name: "brett"},
		{Name: "jones"},
	}, items)
}

func TestScanSliceScansAllColumnTypes(t *testing.T) {
	db := makeDBSchema(t, allTypesSchema)

	var items []rowItem
	rows, err := db.Query(`SELECT * FROM all_types LIMIT 1`)
	require.NoError(t, err)
	err = scnr.Slice(&items, rows)
	require.NoError(t, err)

	assert.EqualValues(t, 2147483640, items[0].ColInt)
	assert.EqualValues(t, 2147483641, items[0].ColInteger)
	assert.EqualValues(t, 126, items[0].ColTinyint)
	assert.EqualValues(t, 127, items[0].ColSmallint)
	assert.EqualValues(t, 2147483642, items[0].ColMediumint)
	assert.EqualValues(t, 9223372036854775800, items[0].ColBigint)
	assert.EqualValues(t, 9223372036854775801, items[0].ColUnsigned)
	assert.EqualValues(t, 2147483643, items[0].ColInt2)
	assert.EqualValues(t, 127, items[0].ColInt8)
	assert.EqualValues(t, "a", items[0].ColCharacter)
	assert.EqualValues(t, "ab", items[0].ColVarchar)
	assert.EqualValues(t, "abc", items[0].ColVarying)
	assert.EqualValues(t, "abcd", items[0].ColNchar)
	assert.EqualValues(t, "abcde", items[0].ColNative)
	assert.EqualValues(t, "abcdef", items[0].ColNvarchar)
	assert.EqualValues(t, "abcdefgh", items[0].ColText)
	assert.EqualValues(t, "abcdefghi", items[0].ColClob)
	assert.EqualValues(t, "abcdefghij", items[0].ColBlob)
	assert.EqualValues(t, "3.1", items[0].ColReal)
	assert.EqualValues(t, 3.14, items[0].ColDouble)
	assert.EqualValues(t, 3.141, items[0].ColFloat)
	assert.EqualValues(t, 3141, items[0].ColNumeric)
	assert.EqualValues(t, 3.1415, items[0].ColDecimal)
	assert.Equal(t, true, items[0].ColBoolean)
	assert.Equal(t, "2017-11-27", items[0].ColDate.Format("2006-01-02"))
	assert.Equal(t, "2017-11-27 17:59", items[0].ColDatetime.Format("2006-01-02 15:04"))

}

var allTypesSchema = `
CREATE TABLE IF NOT EXISTS all_types ( col_int INT, col_integer INTEGER, col_tinyint TINYINT, col_smallint SMALLINT, col_mediumint MEDIUMINT, col_bigint BIGINT, col_unsigned UNSIGNED BIG INT, col_int2 INT2, col_int8 INT8, col_character CHARACTER(20), col_varchar VARCHAR(255), col_varying VARYING CHARACTER(255), col_nchar NCHAR(55), col_native NATIVE CHARACTER(70), col_nvarchar NVARCHAR(100), col_text TEXT, col_clob CLOB, col_blob BLOB, col_real REAL, col_double DOUBLE, col_float FLOAT, col_numeric NUMERIC, col_decimal DECIMAL(10,5), col_boolean BOOLEAN, col_date DATE, col_datetime DATETIME);
INSERT INTO all_types ( col_int, col_integer, col_tinyint, col_smallint, col_mediumint, col_bigint, col_unsigned, col_int2, col_int8, col_character, col_varchar, col_varying, col_nchar, col_native, col_nvarchar, col_text, col_clob, col_blob, col_real, col_double, col_float, col_numeric, col_decimal, col_boolean, col_date, col_datetime)
VALUES ( 2147483640, 2147483641, 126, 127, 2147483642, 9223372036854775800, 9223372036854775801, 2147483643, 127, 'a', 'ab', 'abc', 'abcd', 'abcde', 'abcdef', 'abcdefgh', 'abcdefghi', 'abcdefghij', '3.1', 3.14, 3.141, 3141, 3.1415, 1, '2017-11-27', '2017-11-27 17:59:48');
`

type rowItem struct {
	ColInt       int       `db:"col_int"`
	ColInteger   int       `db:"col_integer"`
	ColTinyint   int8      `db:"col_tinyint"`
	ColSmallint  int16     `db:"col_smallint"`
	ColMediumint int32     `db:"col_mediumint"`
	ColBigint    int64     `db:"col_bigint"`
	ColUnsigned  uint64    `db:"col_unsigned"`
	ColInt2      int       `db:"col_int2"`
	ColInt8      int       `db:"col_int8"`
	ColCharacter string    `db:"col_character"`
	ColVarchar   string    `db:"col_varchar"`
	ColVarying   string    `db:"col_varying"`
	ColNchar     string    `db:"col_nchar"`
	ColNative    string    `db:"col_native"`
	ColNvarchar  string    `db:"col_nvarchar"`
	ColText      string    `db:"col_text"`
	ColClob      string    `db:"col_clob"`
	ColBlob      string    `db:"col_blob"`
	ColReal      string    `db:"col_real"`
	ColDouble    float64   `db:"col_double"`
	ColFloat     float64   `db:"col_float"`
	ColNumeric   int       `db:"col_numeric"`
	ColDecimal   float32   `db:"col_decimal"`
	ColBoolean   bool      `db:"col_boolean"`
	ColDate      time.Time `db:"col_date"`
	ColDatetime  time.Time `db:"col_datetime"`
}
