package clickhouse

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() drivers.InformationSchema {
	return informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context) ([]*drivers.Table, error) {
	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	// Clickhouse does not have a concept of schemas. Both table_catalog and table_schema refer to the database where table is located.
	// Given the usual way of querying table in clickhouse is `SELECT * FROM table_name` or `SELECT * FROM database.table_name`.
	// We map clickhouse database to `database schema` and table_name to `table name`.
	q := `
		SELECT 
			T.database AS SCHEMA,
			T.database = currentDatabase() AS is_default_schema,
			T.name AS NAME,
			if(lower(T.engine) like '%view%', 'VIEW', 'TABLE') AS TABLE_TYPE,
			C.name AS COLUMNS,
			C.type AS COLUMN_TYPE,
			C.position AS ORDINAL_POSITION
		FROM system.tables T
		JOIN system.columns C ON T.database = C.database AND T.name = C.table
		WHERE lower(T.database) NOT IN ('information_schema', 'system') 
		ORDER BY SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`

	rows, err := conn.QueryxContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func (i informationSchema) Lookup(ctx context.Context, db, schema, name string) (*drivers.Table, error) {
	conn, release, err := i.c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	var q string
	var args []any
	q = `
		SELECT 
			T.database AS SCHEMA,
			T.database = currentDatabase() AS is_default_schema,
			T.name AS NAME,
			if(lower(T.engine) like '%view%', 'VIEW', 'TABLE') AS TABLE_TYPE,
			C.name AS COLUMNS,
			C.type AS COLUMN_TYPE,
			C.position AS ORDINAL_POSITION
		FROM system.tables T
		JOIN system.columns C ON T.database = C.database AND T.name = C.table
		WHERE T.database = coalesce(?, currentDatabase()) AND T.name = ?
		ORDER BY SCHEMA, NAME, TABLE_TYPE, ORDINAL_POSITION
	`
	if schema == "" {
		args = append(args, nil, name)
	} else {
		args = append(args, schema, name)
	}

	rows, err := conn.QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables, err := i.scanTables(rows)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (i informationSchema) scanTables(rows *sqlx.Rows) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for rows.Next() {
		var databaseSchema string
		var isDefaultSchema bool
		var name string
		var tableType string
		var columnName string
		var columnType string
		var oridinalPosition int

		err := rows.Scan(&databaseSchema, &isDefaultSchema, &name, &tableType, &columnName, &columnType, &oridinalPosition)
		if err != nil {
			return nil, err
		}

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.DatabaseSchema == databaseSchema && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.Table{
				DatabaseSchema:          databaseSchema,
				IsDefaultDatabaseSchema: isDefaultSchema,
				Name:                    name,
				View:                    tableType == "VIEW",
				Schema:                  &runtimev1.StructType{},
			}
			res = append(res, t)
		}

		// parse column type
		colType, err := databaseTypeToPB(columnType, false)
		if err != nil {
			if !errors.Is(err, errUnsupportedType) {
				return nil, err
			}
			if t.UnsupportedCols == nil {
				t.UnsupportedCols = make(map[string]string)
			}
			t.UnsupportedCols[columnName] = columnType
			continue
		}

		// append column
		t.Schema.Fields = append(t.Schema.Fields, &runtimev1.StructType_Field{
			Name: columnName,
			Type: colType,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
