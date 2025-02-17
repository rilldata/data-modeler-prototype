package duckdb

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
)

type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() drivers.InformationSchema {
	return &informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context, ilike string) ([]*drivers.Table, error) {
	// TODO: this bypasses the acquireMetaConn call in the original implementation. Fix this.
	db, release, err := i.c.acquireDB()
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	rows, err := db.Schema(ctx, ilike, false)
	if err != nil {
		return nil, i.c.checkErr(err)
	}

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func (i informationSchema) Lookup(ctx context.Context, _, _, name string) (*drivers.Table, error) {
	// TODO: this bypasses the acquireMetaConn call in the original implementation. Fix this.
	db, release, err := i.c.acquireDB()
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	rows, err := db.Schema(ctx, name, true)
	if err != nil {
		return nil, i.c.checkErr(err)
	}

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (i informationSchema) scanTables(rows []*rduckdb.Table) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for _, row := range rows {
		t := &drivers.Table{
			Database: row.Database,
			// the database schema changes with every ingestion in duckdb(Refer rduckdb pkg for more info)
			// we pin the read connection to the latest schema and set schema as `main` to give impression that everything is in the same schema
			// This also means that fully qualified names should not be used anywhere
			DatabaseSchema:          "main",
			IsDefaultDatabase:       true,
			IsDefaultDatabaseSchema: true,
			Name:                    row.Name,
			View:                    row.View,
			Schema:                  &runtimev1.StructType{},
		}

		// should NEVER happen, but just to be safe
		if len(row.ColumnNames) != len(row.ColumnTypes) {
			panic(fmt.Errorf("duckdb: column slices have different length"))
		}

		for idx, colName := range row.ColumnNames {
			databaseType := row.ColumnTypes[idx].(string)
			nullable := row.ColumnNullable[idx].(bool)
			colType, err := databaseTypeToPB(databaseType, nullable)
			if err != nil {
				return nil, err
			}

			t.Schema.Fields = append(t.Schema.Fields, &runtimev1.StructType_Field{
				Name: colName.(string),
				Type: colType,
			})
		}

		res = append(res, t)
	}

	return res, nil
}

func databaseTypeToPB(dbt string, nullable bool) (*runtimev1.Type, error) {
	t := &runtimev1.Type{Nullable: nullable}
	match := true
	switch dbt {
	case "INVALID":
		return nil, fmt.Errorf("encountered invalid duckdb type")
	case "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "TINYINT":
		t.Code = runtimev1.Type_CODE_INT8
	case "SMALLINT":
		t.Code = runtimev1.Type_CODE_INT16
	case "INTEGER":
		t.Code = runtimev1.Type_CODE_INT32
	case "BIGINT":
		t.Code = runtimev1.Type_CODE_INT64
	case "UTINYINT":
		t.Code = runtimev1.Type_CODE_UINT8
	case "USMALLINT":
		t.Code = runtimev1.Type_CODE_UINT16
	case "UINTEGER":
		t.Code = runtimev1.Type_CODE_UINT32
	case "UBIGINT":
		t.Code = runtimev1.Type_CODE_UINT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIMESTAMPTZ", "TIMESTAMP WITH TIME ZONE":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "DATE":
		t.Code = runtimev1.Type_CODE_DATE
	case "TIME":
		t.Code = runtimev1.Type_CODE_TIME
	case "TIME WITH TIME ZONE":
		t.Code = runtimev1.Type_CODE_TIME
	case "INTERVAL":
		t.Code = runtimev1.Type_CODE_INTERVAL
	case "HUGEINT":
		t.Code = runtimev1.Type_CODE_INT128
	case "VARCHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "BLOB":
		t.Code = runtimev1.Type_CODE_BYTES
	case "TIMESTAMP_S":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIMESTAMP_MS":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "TIMESTAMP_NS":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "ENUM":
		t.Code = runtimev1.Type_CODE_STRING // TODO - Consider how to handle enums
	case "UUID":
		t.Code = runtimev1.Type_CODE_UUID
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "CHAR":
		t.Code = runtimev1.Type_CODE_STRING
	case "NULL":
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	default:
		match = false
	}

	if match {
		return t, nil
	}

	// Handle complex types

	// Handle arrays. They have the format "type[]"
	if strings.HasSuffix(dbt, "[]") {
		at, err := databaseTypeToPB(dbt[0:len(dbt)-2], true)
		if err != nil {
			return nil, err
		}

		t.Code = runtimev1.Type_CODE_ARRAY
		t.ArrayElementType = at
		return t, nil
	}

	// All other complex types have details in parentheses after the type name.
	base, args, ok := splitBaseAndArgs(dbt)
	if !ok {
		return nil, fmt.Errorf("encountered unsupported duckdb type '%s'", dbt)
	}

	switch base {
	// Example: "DECIMAL(10,20)"
	case "DECIMAL":
		t.Code = runtimev1.Type_CODE_DECIMAL
	// Example: `STRUCT("a" INT, "b" INT)`
	case "STRUCT":
		t.Code = runtimev1.Type_CODE_STRUCT
		t.StructType = &runtimev1.StructType{}

		fieldStrs := splitCommasUnlessQuotedOrNestedInParens(args)
		for _, fieldStr := range fieldStrs {
			// Each field has format `name TYPE` or `"name" TYPE`
			fieldName, fieldTypeStr, ok := splitStructFieldStr(fieldStr)
			if !ok {
				return nil, fmt.Errorf("encountered unsupported duckdb type '%s'", dbt)
			}

			// Convert to type
			fieldType, err := databaseTypeToPB(fieldTypeStr, true)
			if err != nil {
				return nil, err
			}

			// Add to fields
			t.StructType.Fields = append(t.StructType.Fields, &runtimev1.StructType_Field{
				Name: fieldName,
				Type: fieldType,
			})
		}
	// Example: "MAP(VARCHAR, INT)"
	case "MAP":
		fieldStrs := splitCommasUnlessQuotedOrNestedInParens(args)
		if len(fieldStrs) != 2 {
			return nil, fmt.Errorf("encountered unsupported duckdb type '%s'", dbt)
		}

		keyType, err := databaseTypeToPB(fieldStrs[0], true)
		if err != nil {
			return nil, err
		}

		valType, err := databaseTypeToPB(fieldStrs[1], true)
		if err != nil {
			return nil, err
		}

		t.Code = runtimev1.Type_CODE_MAP
		t.MapType = &runtimev1.MapType{
			KeyType:   keyType,
			ValueType: valType,
		}
	case "ENUM":
		t.Code = runtimev1.Type_CODE_STRING // representing enums as strings for now
	default:
		return nil, fmt.Errorf("encountered unsupported duckdb type '%s'", dbt)
	}

	// Done
	return t, nil
}

// Splits a type with args in parentheses, for example:
//
//	`STRUCT("a" INT, "b" INT)` -> (`STRUCT`, `"a" INT, "b" INT`, true)
func splitBaseAndArgs(s string) (string, string, bool) {
	// Split on opening parenthesis
	base, rest, found := strings.Cut(s, "(")
	if !found {
		return "", "", false
	}

	// Remove closing parenthesis
	rest = rest[0 : len(rest)-1]

	return base, rest, true
}

// Splits a comma-separated list, but ignores commas inside strings or nested in parentheses.
// (NOTE: DuckDB escapes strings by replacing `"` with `""`. Example: hello "world" -> "hello ""world""".)
//
// Examples:
//
//	`10,20` -> [`10`, `20`]
//	`VARCHAR, INT` -> [`VARCHAR`, `INT`]
//	`"foo "",""" INT, "bar" STRUCT("a" INT, "b" INT)` -> [`"foo "",""" INT`, `"bar" STRUCT("a" INT, "b" INT)`]
func splitCommasUnlessQuotedOrNestedInParens(s string) []string {
	// Result slice
	splits := []string{}
	// Starting idx of current split
	fromIdx := 0
	// True if quote level is unmatched (this is sufficient for escaped quotes since they will immediately flip again)
	quoted := false
	// Nesting level
	nestCount := 0

	// Consume input character-by-character
	for idx, char := range s {
		// Toggle quoted
		if char == '"' {
			quoted = !quoted
			continue
		}
		// If quoted, don't parse for nesting or commas
		if quoted {
			continue
		}
		// Increase nesting on opening paren
		if char == '(' {
			nestCount++
			continue
		}
		// Decrease nesting on closing paren
		if char == ')' {
			nestCount--
			continue
		}
		// If nested, don't parse for commas
		if nestCount != 0 {
			continue
		}
		// If not nested and there's a comma, add split to result
		if char == ',' {
			splits = append(splits, s[fromIdx:idx])
			fromIdx = idx + 1
			continue
		}
		// If not nested, and there's a space at the start of the split, skip it
		if fromIdx == idx && char == ' ' {
			fromIdx++
			continue
		}
	}

	// Add last split to result and return
	splits = append(splits, s[fromIdx:])
	return splits
}

// splitStructFieldStr splits a single struct name/type pair.
// It expects fieldStr to have the format `name TYPE` or `"name" TYPE`.
// If the name string is quoted and contains escaped quotes `""`, they'll be replaced by `"`.
// For example: splitStructFieldStr(`"hello "" world" VARCHAR`) -> (`hello " world`, `VARCHAR`, true).
func splitStructFieldStr(fieldStr string) (string, string, bool) {
	// If the string DOES NOT start with a `"`, we can just split on the first space.
	if fieldStr == "" || fieldStr[0] != '"' {
		return strings.Cut(fieldStr, " ")
	}

	// Find end of quoted string (skipping `""` since they're escaped quotes)
	idx := 1
	found := false
	for !found && idx < len(fieldStr) {
		// Continue if not a quote
		if fieldStr[idx] != '"' {
			idx++
			continue
		}

		// Skip two ahead if it's two quotes in a row (i.e. an escaped quote)
		if len(fieldStr) > idx+1 && fieldStr[idx+1] == '"' {
			idx += 2
			continue
		}

		// It's the last quote of the string. We're done.
		idx++
		found = true
	}

	// If not found, format was unexpected
	if !found {
		return "", "", false
	}

	// Remove surrounding `"` and replace escaped quotes `""` with `"`
	nameStr := strings.ReplaceAll(fieldStr[1:idx-1], `""`, `"`)

	// The rest of the string is the type, minus the initial space
	typeStr := strings.TrimLeft(fieldStr[idx:], " ")

	return nameStr, typeStr, true
}
