package rduckdb

import (
	"context"
	"fmt"
)

type Table struct {
	Database       string `db:"database"`
	Name           string `db:"name"`
	View           bool   `db:"view"`
	ColumnNames    []any  `db:"column_names"`
	ColumnTypes    []any  `db:"column_types"`
	ColumnNullable []any  `db:"column_nullable"`
}

func (d *db) Schema(ctx context.Context, like string, matchCase bool) ([]*Table, error) {
	connx, release, err := d.AcquireReadConnection(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = release()
	}()

	var likeClause, likeOp string
	var args []any
	if matchCase {
		likeOp = "like"
	} else {
		likeOp = "ilike"
	}
	if like != "" {
		likeClause = fmt.Sprintf(" and t.table_name %s ?", likeOp)
		args = []any{like}
	}

	q := fmt.Sprintf(`
		SELECT
			coalesce(t.table_catalog, current_database()) AS "database",
			t.table_name AS "name",
			t.table_type = 'VIEW' AS "view", 
			array_agg(c.column_name ORDER BY c.ordinal_position) AS "column_names",
			array_agg(c.data_type ORDER BY c.ordinal_position) AS "column_types",
			array_agg(c.is_nullable = 'YES' ORDER BY c.ordinal_position) AS "column_nullable"
		FROM information_schema.tables t
		JOIN information_schema.columns c 
			ON t.table_schema = c.table_schema 
			AND t.table_name = c.table_name
		WHERE database = current_database() 
			AND t.table_schema = current_schema()
			%s
		GROUP BY 1, 2, 3
		ORDER BY 1, 2, 3
	`, likeClause)

	var res []*Table
	err = connx.SelectContext(ctx, &res, q, args...)
	if err != nil {
		return nil, err
	}

	// due to external table storage the information_schema always returns table type as view
	// so we look at catalog to determine if it is a view or table
	// NOTE : there is a chance of inconsistency since tables can get updated between these two calls
	tables := d.catalog.listTables()
	catalog := make(map[string]*tableMeta)
	for _, table := range tables {
		catalog[table.Name] = table
	}
	for _, t := range res {
		table, ok := catalog[t.Name]
		if ok {
			t.View = table.Type == "VIEW"
		}
	}

	return res, nil
}
