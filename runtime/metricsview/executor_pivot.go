package metricsview

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

// rewriteQueryForPivot rewrites a query for pivoting if qry.PivotOn is not empty.
// It rewrites queries with PivotOn fields to a simpler underlying query,
// and returns a pivotAST that represents a PIVOT query against the results of the underlying query.
func (e *Executor) rewriteQueryForPivot(qry *Query) (*pivotAST, bool, error) {
	// Skip if we're not pivoting
	if len(qry.PivotOn) == 0 {
		return nil, false, nil
	}

	// Check pivot fields are dims (not measures)
	for _, f := range qry.PivotOn {
		var found bool
		for _, d := range qry.Dimensions {
			if d.Name == f {
				found = true
				break
			}
		}
		if !found {
			return nil, false, fmt.Errorf("pivot field %q not found in dimensions", f)
		}
	}

	// Check sort fields are non-pivoted dims
	for _, s := range qry.Sort {
		var found bool
		for _, d := range qry.Dimensions {
			if d.Name == s.Name {
				found = true
				break
			}
		}
		if !found {
			return nil, false, fmt.Errorf("sort field %q is not a dimension (pivot queries can only sort on non-pivoted dimensions)", s.Name)
		}

		for _, f := range qry.PivotOn {
			if f == s.Name {
				return nil, false, fmt.Errorf("sort field %q is a pivot field (pivot queries can only sort on non-pivoted dimensions)", s.Name)
			}
		}
	}

	// Determine dialect for the PIVOT (in practice, this currently always becomes DuckDB because it's the only OLAP that supports pivoting)
	dialect := e.olap.Dialect()
	if !dialect.CanPivot() {
		dialect = drivers.DialectDuckDB
	}

	// Build a pivotAST based on fields to apply during and after the pivot (instead of in the underlying query)
	ast := &pivotAST{
		keep:    nil, // Populated below
		on:      qry.PivotOn,
		using:   nil, // Populated below
		orderBy: nil, // Populated below
		limit:   qry.Limit,
		offset:  qry.Offset,
		label:   qry.Label,
		dialect: dialect,
	}
	for _, d := range qry.Dimensions {
		var found bool
		for _, f := range qry.PivotOn {
			if f == d.Name {
				found = true
				break
			}
		}
		if !found {
			ast.keep = append(ast.keep, d.Name)
		}
	}
	for _, m := range qry.Measures {
		ast.using = append(ast.using, m.Name)
	}
	for _, f := range qry.Sort {
		ast.orderBy = append(ast.orderBy, OrderFieldNode(f))
	}

	// Remove parameters from the underlying query that are now handled in the pivot AST
	qry.PivotOn = nil
	qry.Sort = nil
	qry.Limit = nil
	qry.Offset = nil
	qry.Label = false

	// If we have a cell limit, apply a row limit just above it to the underlying query.
	// This prevents the DB from scanning too much data before we can detect that the query will exceed the cell limit.
	if e.instanceCfg.PivotCellLimit != 0 {
		cols := int64(len(qry.Dimensions) + len(qry.Measures))
		ast.underlyingCellCap = e.instanceCfg.PivotCellLimit
		ast.underlyingRowCap = e.instanceCfg.PivotCellLimit / cols

		tmp := ast.underlyingRowCap + 1
		qry.Limit = &tmp
	}

	return ast, true, nil
}

// executePivotExport executes a PIVOT query prepared using rewriteQueryForPivot, and exports the result to a file in the given format.
func (e *Executor) executePivotExport(ctx context.Context, ast *AST, pivot *pivotAST, format string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultPivotExportTimeout)
	defer cancel()

	// Build underlying SQL
	underlyingSQL, args, err := ast.SQL()
	if err != nil {
		return "", err
	}

	// If the metrics view's connector doesn't support pivoting, we export the underlying (non-pivoted) data to a Parquet file, and handover to DuckDB to do the pivot.
	pivotConnector := e.metricsView.Connector
	if !e.olap.Dialect().CanPivot() {
		// Export non-pivoted data to a temporary Parquet file
		path, err := e.executeExport(ctx, "parquet", e.metricsView.Connector, map[string]any{
			"sql":  underlyingSQL,
			"args": args,
		})
		if err != nil {
			return "", fmt.Errorf("failed to execute pre-pivot export: %w", err)
		}
		defer os.Remove(path)

		// Hard-code DuckDB as the connector that executes the pivot
		pivotConnector = "duckdb"
		underlyingSQL = fmt.Sprintf("SELECT * FROM '%s'", path)
		args = nil

		// Check for consistency with rewriteQueryForPivot
		if pivot.dialect != drivers.DialectDuckDB {
			return "", fmt.Errorf("cannot execute pivot: the pivot AST fell back to dialect %q, not DuckDB", pivot.dialect.String())
		}
	}

	// Unfortunately, DuckDB does not support passing args to a PIVOT query.
	// So we stage the underlying data in a temporary table and run the PIVOT against that table instead.
	olap, release, err := e.rt.OLAP(ctx, e.instanceID, pivotConnector)
	if err != nil {
		return "", fmt.Errorf("failed to acquire OLAP for serving pivot: %w", err)
	}
	defer release()
	var path string
	err = olap.WithConnection(ctx, e.priority, false, false, func(wrappedCtx context.Context, ensuredCtx context.Context, conn *sql.Conn) error {
		// Stage the underlying data in a temporary table
		alias, err := randomString("t", 8)
		if err != nil {
			return fmt.Errorf("failed to generate random alias: %w", err)
		}
		err = olap.Exec(wrappedCtx, &drivers.Statement{
			Query: fmt.Sprintf("CREATE TEMPORARY TABLE %s AS (%s)", alias, underlyingSQL),
			Args:  args,
		})
		if err != nil {
			return fmt.Errorf("failed to stage underlying data for pivot: %w", err)
		}

		// Defer cleanup of the temporary table
		defer func() {
			err = olap.Exec(ensuredCtx, &drivers.Statement{
				Query: fmt.Sprintf("DROP TABLE %s", alias),
			})
			if err != nil {
				l, err2 := e.rt.InstanceLogger(ctx, e.instanceID)
				if err2 == nil {
					l.Error("duckdb: failed to cleanup temporary table for pivot export", zap.Error(err))
				}
			}
		}()

		// Build the PIVOT query
		pivotSQL, err := pivot.SQL(ast, alias, true)
		if err != nil {
			return err
		}

		// Execute the pivot export
		path, err = e.executeExport(wrappedCtx, format, pivotConnector, map[string]any{
			"sql": pivotSQL,
		})
		if err != nil {
			return fmt.Errorf("failed to execute pivot export: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

// pivotAST represents config for generating a PIVOT query.
type pivotAST struct {
	keep    []string
	on      []string
	using   []string
	orderBy []OrderFieldNode
	limit   *int64
	offset  *int64

	label             bool
	dialect           drivers.Dialect
	underlyingCellCap int64
	underlyingRowCap  int64
}

// SQL generates a query that outputs a pivoted table based on the pivot config and data in the underlying query.
// The underlyingAlias must be an alias for a table that holds the data produced by underlyingAST.SQL().
func (a *pivotAST) SQL(underlyingAST *AST, underlyingAlias string, checkCap bool) (string, error) {
	if !a.dialect.CanPivot() {
		return "", fmt.Errorf("pivot queries not supported for dialect %q", a.dialect.String())
	}

	b := &strings.Builder{}

	// Circumstances make it easiest to enforce the pivot cell limit in the PIVOT query itself. To do that, we need to be creative.
	// We leverage CTEs and DuckDB's ERROR() function to enforce the limit.
	// This is pretty DuckDB-specific, but that's also currently the only OLAP we use that supports pivoting.
	// The query looks something like:
	//
	//   WITH t AS (SELECT * FROM <underlyingAlias> WHERE IF(EXISTS (SELECT COUNT(*) AS count FROM <underlyingAlias> HAVING count > <limit>), ERROR('pivot query exceeds limit'), TRUE))
	//   PIVOT t2 ON ...
	//
	if checkCap {
		tmpAlias, err := randomString("t", 8)
		if err != nil {
			return "", fmt.Errorf("failed to generate random alias: %w", err)
		}

		b.WriteString("WITH ")
		b.WriteString(tmpAlias)
		b.WriteString(" AS (SELECT * FROM ")
		b.WriteString(underlyingAlias)
		b.WriteString(" WHERE IF(EXISTS (SELECT COUNT(*) AS count FROM ")
		b.WriteString(underlyingAlias)
		b.WriteString(" HAVING count > ")
		b.WriteString(strconv.FormatInt(a.underlyingRowCap, 10))
		b.WriteString("), ERROR('pivot query exceeds limit of ")
		b.WriteString(strconv.FormatInt(a.underlyingCellCap, 10))
		b.WriteString(" cells'), TRUE)) ")

		underlyingAlias = tmpAlias
	}

	// If we need to label some fields (in practice, this will be non-pivoted dims during exports),
	// we emit a query like: SELECT d1 AS "L1", d2 AS "L2", * EXCLUDE (d1, d2) FROM (PIVOT ...)
	wrapWithLabels := a.label && len(a.keep) > 0
	if wrapWithLabels {
		b.WriteString("SELECT ")
		for _, fn := range a.keep {
			f, ok := findField(fn, underlyingAST.Root.DimFields)
			if !ok {
				return "", fmt.Errorf("pivot keep dimension %q not found in underlying query", fn)
			}

			b.WriteString(a.dialect.EscapeIdentifier(f.Name))
			if f.Label != "" {
				b.WriteString(" AS ")
				b.WriteString(a.dialect.EscapeIdentifier(f.Label))
			}
			b.WriteString(", ")
		}

		b.WriteString("* EXCLUDE (")
		for i, fn := range a.keep {
			f, ok := findField(fn, underlyingAST.Root.DimFields)
			if !ok {
				return "", fmt.Errorf("pivot keep dimension %q not found in underlying query", fn)
			}

			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a.dialect.EscapeIdentifier(f.Name))
		}

		b.WriteString(") FROM (")
	}

	// Build a PIVOT query like: PIVOT (<underlyingSQL>) ON <dimensions> USING <measures> ORDER BY <sort> LIMIT <limit> OFFSET <offset>
	b.WriteString("PIVOT ")
	b.WriteString(underlyingAlias)
	b.WriteString(" ON ")
	for i, fn := range a.on {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(a.dialect.EscapeIdentifier(fn))
	}

	if len(a.using) > 0 {
		b.WriteString(" USING ")
		for i, fn := range a.using {
			f, ok := findField(fn, underlyingAST.Root.MeasureFields)
			if !ok {
				return "", fmt.Errorf("pivot using measure %q not found in underlying query", fn)
			}

			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString("ANY_VALUE(")
			b.WriteString(a.dialect.EscapeIdentifier(fn))
			b.WriteString(")")
			b.WriteString(" AS ")
			if a.label && f.Label != "" {
				b.WriteString(a.dialect.EscapeIdentifier(f.Label))
			} else {
				b.WriteString(a.dialect.EscapeIdentifier(f.Name))
			}
		}
	}

	if len(a.orderBy) > 0 {
		b.WriteString(" ORDER BY ")
		for i, f := range a.orderBy {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a.dialect.OrderByExpression(f.Name, f.Desc))
		}
	}

	if a.limit != nil {
		b.WriteString(" LIMIT ")
		b.WriteString(strconv.FormatInt(*a.limit, 10))
	}

	if a.offset != nil {
		b.WriteString(" OFFSET ")
		b.WriteString(strconv.FormatInt(*a.offset, 10))
	}

	if wrapWithLabels {
		b.WriteString(")")
	}

	return b.String(), nil
}

func randomString(prefix string, n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}

func findField(n string, fs []FieldNode) (FieldNode, bool) {
	for _, f := range fs {
		if f.Name == n {
			return f, true
		}
	}
	return FieldNode{}, false
}
