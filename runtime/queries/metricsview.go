package queries

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/xuri/excelize/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// resolveMeasures returns the selected measures
func resolveMeasures(mv *runtimev1.MetricsViewSpec, inlines []*runtimev1.InlineMeasure, selectedNames []string) ([]*runtimev1.MetricsViewSpec_MeasureV2, error) {
	// Build combined measures
	ms := make([]*runtimev1.MetricsViewSpec_MeasureV2, len(selectedNames))
	for i, n := range selectedNames {
		found := false
		// Search in the inlines (take precedence)
		for _, m := range inlines {
			if m.Name == n {
				ms[i] = &runtimev1.MetricsViewSpec_MeasureV2{
					Name:       m.Name,
					Expression: m.Expression,
				}
				found = true
				break
			}
		}
		if found {
			continue
		}
		// Search in the metrics view
		for _, m := range mv.Measures {
			if m.Name == n {
				ms[i] = m
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("measure does not exist: '%s'", n)
		}
	}

	return ms, nil
}

func metricsQuery(ctx context.Context, olap drivers.OLAPStore, priority int, sql string, args []any) ([]*runtimev1.MetricsViewColumn, []*structpb.Struct, error) {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer rows.Close()

	data, err := rowsToData(rows)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	return structTypeToMetricsViewColumn(rows.Schema), data, nil
}

func olapQuery(ctx context.Context, olap drivers.OLAPStore, priority int, sql string, args []any) (*runtimev1.StructType, []*structpb.Struct, error) {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer rows.Close()

	data, err := rowsToData(rows)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	return rows.Schema, data, nil
}

func rowsToData(rows *drivers.Result) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct, err := pbutil.ToStruct(rowMap, rows.Schema)
		if err != nil {
			return nil, err
		}

		data = append(data, rowStruct)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func structTypeToMetricsViewColumn(v *runtimev1.StructType) []*runtimev1.MetricsViewColumn {
	res := make([]*runtimev1.MetricsViewColumn, len(v.Fields))
	for i, f := range v.Fields {
		res[i] = &runtimev1.MetricsViewColumn{
			Name:     f.Name,
			Type:     f.Type.Code.String(),
			Nullable: f.Type.Nullable,
		}
	}
	return res
}

// buildFilterClauseForMetricsViewFilter builds a SQL string of conditions joined with AND.
// Unless the result is empty, it is prefixed with "AND".
// I.e. it has the format "AND (...) AND (...) ...".
func buildFilterClauseForMetricsViewFilter(mv *runtimev1.MetricsViewSpec, filter *runtimev1.MetricsViewFilter, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity) (string, []any, error) {
	var clauses []string
	var args []any

	if filter != nil && filter.Include != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(mv, filter.Include, false, dialect)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, clause)
		args = append(args, clauseArgs...)
	}

	if filter != nil && filter.Exclude != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(mv, filter.Exclude, true, dialect)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, clause)
		args = append(args, clauseArgs...)
	}

	if policy != nil && policy.RowFilter != "" {
		clauses = append(clauses, "AND "+policy.RowFilter)
	}

	return strings.Join(clauses, " "), args, nil
}

// buildFilterClauseForConditions returns a string with the format "AND (...) AND (...) ..."
func buildFilterClauseForConditions(mv *runtimev1.MetricsViewSpec, conds []*runtimev1.MetricsViewFilter_Cond, exclude bool, dialect drivers.Dialect) (string, []any, error) {
	var clauses []string
	var args []any

	for _, cond := range conds {
		condClause, condArgs, err := buildFilterClauseForCondition(mv, cond, exclude, dialect)
		if err != nil {
			return "", nil, err
		}
		if condClause == "" {
			continue
		}
		clauses = append(clauses, condClause)
		args = append(args, condArgs...)
	}

	return strings.Join(clauses, " "), args, nil
}

// buildFilterClauseForCondition returns a string with the format "AND (...)"
func buildFilterClauseForCondition(mv *runtimev1.MetricsViewSpec, cond *runtimev1.MetricsViewFilter_Cond, exclude bool, dialect drivers.Dialect) (string, []any, error) {
	var clauses []string
	var args []any

	// NOTE: Looking up for dimension like this will lead to O(nm).
	//       Ideal way would be to create a map, but we need to find a clean solution down the line
	dim, err := metricsViewDimension(mv, cond.Name)
	if err != nil {
		return "", nil, err
	}
	name := safeName(metricsViewDimensionColumn(dim))

	notKeyword := ""
	if exclude {
		notKeyword = "NOT"
	}

	// Tracks if we found NULL(s) in cond.In
	inHasNull := false

	// Build "dim [NOT] IN (?, ?, ...)" clause
	if len(cond.In) > 0 {
		// Add to args, skipping nulls
		for _, val := range cond.In {
			if _, ok := val.Kind.(*structpb.Value_NullValue); ok {
				inHasNull = true
				continue // Handled later using "dim IS [NOT] NULL" clause
			}
			arg, err := pbutil.FromValue(val)
			if err != nil {
				return "", nil, fmt.Errorf("filter error: %w", err)
			}
			args = append(args, arg)
		}

		// If there were non-null args, add a "dim [NOT] IN (...)" clause
		if len(args) > 0 {
			questionMarks := strings.Join(repeatString("?", len(args)), ",")
			var clause string
			// Build [NOT] list_has_any("dim", ARRAY[?, ?, ...])
			if dim.Unnest && dialect != drivers.DialectDruid {
				clause = fmt.Sprintf("%s list_has_any(%s, ARRAY[%s])", notKeyword, name, questionMarks)
			} else {
				clause = fmt.Sprintf("%s %s IN (%s)", name, notKeyword, questionMarks)
			}
			clauses = append(clauses, clause)
		}
	}

	// Build "dim [NOT] ILIKE ?"
	if len(cond.Like) > 0 {
		for _, val := range cond.Like {
			var clause string
			// Build [NOT] len(list_filter("dim", x -> x ILIKE ?)) > 0
			if dim.Unnest && dialect != drivers.DialectDruid {
				clause = fmt.Sprintf("%s len(list_filter(%s, x -> x %s ILIKE ?)) > 0", notKeyword, name, notKeyword)
			} else {
				if dialect == drivers.DialectDruid {
					// Druid does not support ILIKE
					clause = fmt.Sprintf("LOWER(%s) %s LIKE LOWER(?)", name, notKeyword)
				} else {
					clause = fmt.Sprintf("%s %s ILIKE ?", name, notKeyword)
				}
			}

			args = append(args, val)
			clauses = append(clauses, clause)
		}
	}

	// Add null check
	// NOTE: DuckDB doesn't handle NULL values in an "IN" expression. They must be checked with a "dim IS [NOT] NULL" clause.
	if inHasNull {
		clauses = append(clauses, fmt.Sprintf("%s IS %s NULL", name, notKeyword))
	}

	// If no checks were added, exit
	if len(clauses) == 0 {
		return "", nil, nil
	}

	// Join conditions
	var condJoiner string
	if exclude {
		condJoiner = " AND "
	} else {
		condJoiner = " OR "
	}
	condsClause := strings.Join(clauses, condJoiner)

	// When you have "dim NOT IN (a, b, ...)", then NULL values are always excluded, even if NULL is not in the list.
	// E.g. this returns zero rows: "select * from (select 1 as a union select null as a) where a not in (1)"
	// We need to explicitly include it.
	if exclude && !inHasNull && len(condsClause) > 0 {
		condsClause += fmt.Sprintf(" OR %s IS NULL", name)
	}

	// Done
	return fmt.Sprintf("AND (%s) ", condsClause), args, nil
}

type columnIdentifier struct {
	// expression to use instead of a name for dimension or expression
	// EG: measure expression : impressions => "impressions" (would be aliases in query)
	//     dimension column : publisher => "publisher"
	//     dimension expression : tld => "regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2)" (needed since tld might not be selected)
	expr   string
	unnest bool
}

func newIdentifier(name string) columnIdentifier {
	return columnIdentifier{safeName(name), false}
}

func dimensionAliases(mv *runtimev1.MetricsViewSpec) map[string]columnIdentifier {
	aliases := map[string]columnIdentifier{}
	for _, dim := range mv.Dimensions {
		aliases[dim.Name] = columnIdentifier{safeName(metricsViewDimensionColumn(dim)), dim.Unnest}
	}
	return aliases
}

func buildExpression(expr *runtimev1.Expression, allowedIdentifiers map[string]columnIdentifier, dialect drivers.Dialect) (string, []any, error) {
	var emptyArg []any
	switch e := expr.Expression.(type) {
	case *runtimev1.Expression_Val:
		arg, err := pbutil.FromValue(e.Val)
		if err != nil {
			return "", emptyArg, err
		}
		return "?", []any{arg}, nil

	case *runtimev1.Expression_Ident:
		col, ok := allowedIdentifiers[e.Ident]
		if !ok {
			return "", emptyArg, fmt.Errorf("unknown column filter: %s", e.Ident)
		}
		return col.expr, emptyArg, nil

	case *runtimev1.Expression_Cond:
		return buildConditionExpression(e.Cond, allowedIdentifiers, dialect)
	}

	return "", emptyArg, nil
}

func buildConditionExpression(cond *runtimev1.Condition, allowedIdentifiers map[string]columnIdentifier, dialect drivers.Dialect) (string, []any, error) {
	switch cond.Op {
	case runtimev1.Operation_OPERATION_LIKE, runtimev1.Operation_OPERATION_NLIKE:
		return buildLikeExpression(cond, allowedIdentifiers, dialect)

	case runtimev1.Operation_OPERATION_IN, runtimev1.Operation_OPERATION_NIN:
		return buildInExpression(cond, allowedIdentifiers, dialect)

	case runtimev1.Operation_OPERATION_AND:
		return buildAndOrExpressions(cond, allowedIdentifiers, dialect, " AND ")

	case runtimev1.Operation_OPERATION_OR:
		return buildAndOrExpressions(cond, allowedIdentifiers, dialect, " OR ")

	default:
		leftExpr, args, err := buildExpression(cond.Exprs[0], allowedIdentifiers, dialect)
		if err != nil {
			return "", nil, err
		}

		rightExpr, subArgs, err := buildExpression(cond.Exprs[1], allowedIdentifiers, dialect)
		if err != nil {
			return "", nil, err
		}
		args = append(args, subArgs...)

		return fmt.Sprintf("(%s) %s (%s)", leftExpr, conditionExpressionOperation(cond.Op), rightExpr), args, nil
	}
}

func buildLikeExpression(cond *runtimev1.Condition, allowedIdentifiers map[string]columnIdentifier, dialect drivers.Dialect) (string, []any, error) {
	if len(cond.Exprs) != 2 {
		return "", nil, fmt.Errorf("like/not like expression should have exactly 2 sub expressions")
	}

	leftExpr, args, err := buildExpression(cond.Exprs[0], allowedIdentifiers, dialect)
	if err != nil {
		return "", nil, err
	}

	rightExpr, subArgs, err := buildExpression(cond.Exprs[1], allowedIdentifiers, dialect)
	if err != nil {
		return "", nil, err
	}
	args = append(args, subArgs...)

	notKeyword := ""
	if cond.Op == runtimev1.Operation_OPERATION_NLIKE {
		notKeyword = "NOT"
	}

	// identify if immediate identifier has unnest
	unnest := false
	ident, isIdent := cond.Exprs[0].Expression.(*runtimev1.Expression_Ident)
	if isIdent {
		i := allowedIdentifiers[ident.Ident]
		unnest = i.unnest
	}

	var clause string
	// Build [NOT] len(list_filter("dim", x -> x ILIKE ?)) > 0
	if unnest && dialect != drivers.DialectDruid {
		clause = fmt.Sprintf("%s len(list_filter(%s, x -> x %s ILIKE %s)) > 0", notKeyword, leftExpr, notKeyword, rightExpr)
	} else {
		if dialect == drivers.DialectDruid {
			// Druid does not support ILIKE
			clause = fmt.Sprintf("LOWER(%s) %s LIKE LOWER(%s)", leftExpr, notKeyword, rightExpr)
		} else {
			clause = fmt.Sprintf("%s %s ILIKE %s", leftExpr, notKeyword, rightExpr)
		}
	}

	// When you have "dim NOT ILIKE '...'", then NULL values are always excluded.
	// We need to explicitly include it.
	if cond.Op == runtimev1.Operation_OPERATION_NLIKE {
		clause += fmt.Sprintf(" OR %s IS NULL", leftExpr)
	}

	return clause, args, nil
}

func buildInExpression(cond *runtimev1.Condition, allowedIdentifiers map[string]columnIdentifier, dialect drivers.Dialect) (string, []any, error) {
	if len(cond.Exprs) <= 1 {
		return "", nil, fmt.Errorf("in/not in expression should have atleast 2 sub expressions")
	}

	leftExpr, args, err := buildExpression(cond.Exprs[0], allowedIdentifiers, dialect)
	if err != nil {
		return "", nil, err
	}

	notKeyword := ""
	exclude := cond.Op == runtimev1.Operation_OPERATION_NIN
	if exclude {
		notKeyword = "NOT"
	}

	inHasNull := false
	var valClauses []string
	// Add to args, skipping nulls
	for _, subExpr := range cond.Exprs[1:] {
		if v, isVal := subExpr.Expression.(*runtimev1.Expression_Val); isVal {
			if _, isNull := v.Val.Kind.(*structpb.Value_NullValue); isNull {
				inHasNull = true
				continue // Handled later using "dim IS [NOT] NULL" clause
			}
		}
		inVal, subArgs, err := buildExpression(subExpr, allowedIdentifiers, dialect)
		if err != nil {
			return "", nil, err
		}
		args = append(args, subArgs...)
		valClauses = append(valClauses, inVal)
	}

	// identify if immediate identifier has unnest
	// TODO: do we need to do a deeper check?
	unnest := false
	ident, isIndent := cond.Exprs[0].Expression.(*runtimev1.Expression_Ident)
	if isIndent {
		i := allowedIdentifiers[ident.Ident]
		unnest = i.unnest
	}

	clauses := make([]string, 0)

	// If there were non-null args, add a "dim [NOT] IN (...)" clause
	if len(valClauses) > 0 {
		questionMarks := strings.Join(valClauses, ",")
		var clause string
		// Build [NOT] list_has_any("dim", ARRAY[?, ?, ...])
		if unnest && dialect != drivers.DialectDruid {
			clause = fmt.Sprintf("%s list_has_any(%s, ARRAY[%s])", notKeyword, leftExpr, questionMarks)
		} else {
			clause = fmt.Sprintf("%s %s IN (%s)", leftExpr, notKeyword, questionMarks)
		}
		clauses = append(clauses, clause)
	}

	if inHasNull {
		// Add null check
		// NOTE: DuckDB doesn't handle NULL values in an "IN" expression. They must be checked with a "dim IS [NOT] NULL" clause.
		clauses = append(clauses, fmt.Sprintf("%s IS %s NULL", leftExpr, notKeyword))
	}
	var condsClause string
	if exclude {
		condsClause = strings.Join(clauses, " AND ")
	} else {
		condsClause = strings.Join(clauses, " OR ")
	}
	if exclude && !inHasNull && len(clauses) > 0 {
		// When you have "dim NOT IN (a, b, ...)", then NULL values are always excluded, even if NULL is not in the list.
		// E.g. this returns zero rows: "select * from (select 1 as a union select null as a) where a not in (1)"
		// We need to explicitly include it.
		condsClause += fmt.Sprintf(" OR %s IS NULL", leftExpr)
	}

	return condsClause, args, nil
}

func buildAndOrExpressions(cond *runtimev1.Condition, allowedIdentifiers map[string]columnIdentifier, dialect drivers.Dialect, joiner string) (string, []any, error) {
	clauses := make([]string, 0)
	var args []any
	for _, expr := range cond.Exprs {
		clause, subArgs, err := buildExpression(expr, allowedIdentifiers, dialect)
		if err != nil {
			return "", nil, err
		}
		args = append(args, subArgs...)
		clauses = append(clauses, fmt.Sprintf("(%s)", clause))
	}
	return strings.Join(clauses, joiner), args, nil
}

func conditionExpressionOperation(oprn runtimev1.Operation) string {
	switch oprn {
	case runtimev1.Operation_OPERATION_EQ:
		return "="
	case runtimev1.Operation_OPERATION_NEQ:
		return "!="
	case runtimev1.Operation_OPERATION_LT:
		return "<"
	case runtimev1.Operation_OPERATION_LTE:
		return "<="
	case runtimev1.Operation_OPERATION_GT:
		return ">"
	case runtimev1.Operation_OPERATION_GTE:
		return ">="
	}
	panic(fmt.Sprintf("unknown condition operation: %v", oprn))
}

func convertFilterToExpression(filter *runtimev1.MetricsViewFilter) *runtimev1.Expression {
	var exprs []*runtimev1.Expression

	if len(filter.Include) > 0 {
		var includeExprs []*runtimev1.Expression
		for _, cond := range filter.Include {
			domExpr := convertDimensionFilterToExpression(cond, false)
			if domExpr != nil {
				includeExprs = append(includeExprs, domExpr)
			}
		}
		exprs = append(exprs, FilterOrClause(includeExprs))
	}

	if len(filter.Exclude) > 0 {
		for _, cond := range filter.Exclude {
			domExpr := convertDimensionFilterToExpression(cond, true)
			if domExpr != nil {
				exprs = append(exprs, domExpr)
			}
		}
	}

	if len(exprs) == 1 {
		return exprs[0]
	} else if len(exprs) > 1 {
		return FilterAndClause(exprs)
	}
	return nil
}

func convertDimensionFilterToExpression(cond *runtimev1.MetricsViewFilter_Cond, exclude bool) *runtimev1.Expression {
	var inExpr *runtimev1.Expression
	if len(cond.In) > 0 {
		var inExprs []*runtimev1.Expression
		for _, inVal := range cond.In {
			inExprs = append(inExprs, FilterValue(inVal))
		}
		if exclude {
			inExpr = FilterNotInClause(FilterColumn(cond.Name), inExprs)
		} else {
			inExpr = FilterInClause(FilterColumn(cond.Name), inExprs)
		}
	}

	var likeExpr *runtimev1.Expression
	if len(cond.Like) == 1 {
		if exclude {
			likeExpr = FilterNotLikeClause(FilterColumn(cond.Name), FilterValue(structpb.NewStringValue(cond.Like[0])))
		} else {
			likeExpr = FilterLikeClause(FilterColumn(cond.Name), FilterValue(structpb.NewStringValue(cond.Like[0])))
		}
	} else if len(cond.Like) > 1 {
		var likeExprs []*runtimev1.Expression
		for _, l := range cond.Like {
			col := FilterColumn(cond.Name)
			val := FilterValue(structpb.NewStringValue(l))
			if exclude {
				likeExprs = append(likeExprs, FilterNotLikeClause(col, val))
			} else {
				likeExprs = append(likeExprs, FilterLikeClause(col, val))
			}
		}
		if exclude {
			likeExpr = FilterAndClause(likeExprs)
		} else {
			likeExpr = FilterOrClause(likeExprs)
		}
	}

	if inExpr != nil && likeExpr != nil {
		if exclude {
			return FilterAndClause([]*runtimev1.Expression{inExpr, likeExpr})
		}
		return FilterOrClause([]*runtimev1.Expression{inExpr, likeExpr})
	} else if inExpr != nil {
		return inExpr
	} else if likeExpr != nil {
		return likeExpr
	}

	return nil
}

func repeatString(val string, n int) []string {
	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = val
	}
	return res
}

func convertToString(pbvalue *structpb.Value) (string, error) {
	switch pbvalue.GetKind().(type) {
	case *structpb.Value_StructValue:
		bts, err := protojson.Marshal(pbvalue)
		if err != nil {
			return "", err
		}

		return string(bts), nil
	case *structpb.Value_NullValue:
		return "", nil
	default:
		return fmt.Sprintf("%v", pbvalue.AsInterface()), nil
	}
}

func convertToXLSXValue(pbvalue *structpb.Value) (interface{}, error) {
	switch pbvalue.GetKind().(type) {
	case *structpb.Value_StructValue:
		bts, err := protojson.Marshal(pbvalue)
		if err != nil {
			return "", err
		}

		return string(bts), nil
	case *structpb.Value_NullValue:
		return "", nil
	default:
		return pbvalue.AsInterface(), nil
	}
}

func metricsViewDimensionToSafeColumn(mv *runtimev1.MetricsViewSpec, dimName string) (string, error) {
	dimName = strings.ToLower(dimName)
	dimension, err := metricsViewDimension(mv, dimName)
	if err != nil {
		return "", err
	}
	return safeName(metricsViewDimensionColumn(dimension)), nil
}

func metricsViewDimension(mv *runtimev1.MetricsViewSpec, dimName string) (*runtimev1.MetricsViewSpec_DimensionV2, error) {
	for _, dimension := range mv.Dimensions {
		if strings.EqualFold(dimension.Name, dimName) {
			return dimension, nil
		}
	}
	return nil, fmt.Errorf("dimension %s not found", dimName)
}

func metricsViewDimensionColumn(dimension *runtimev1.MetricsViewSpec_DimensionV2) string {
	if dimension.Column != "" {
		return dimension.Column
	}
	// backwards compatibility for older projects that have not run reconcile on this dashboard
	// in that case `column` will not be present
	return dimension.Name
}

func metricsViewMeasureExpression(mv *runtimev1.MetricsViewSpec, measureName string) (string, error) {
	for _, measure := range mv.Measures {
		if strings.EqualFold(measure.Name, measureName) {
			return measure.Expression, nil
		}
	}
	return "", fmt.Errorf("measure %s not found", measureName)
}

func writeCSV(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, writer io.Writer) error {
	w := csv.NewWriter(writer)

	record := make([]string, 0, len(meta))
	for _, field := range meta {
		record = append(record, field.Name)
	}
	if err := w.Write(record); err != nil {
		return err
	}
	record = record[:0]

	for _, structs := range data {
		for _, field := range meta {
			pbvalue := structs.Fields[field.Name]
			str, err := convertToString(pbvalue)
			if err != nil {
				return err
			}

			record = append(record, str)
		}

		if err := w.Write(record); err != nil {
			return err
		}

		record = record[:0]
	}

	w.Flush()

	return nil
}

func writeXLSX(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, writer io.Writer) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}

	headers := make([]interface{}, 0, len(meta))
	for _, v := range meta {
		headers = append(headers, v.Name)
	}

	if err := sw.SetRow("A1", headers, excelize.RowOpts{Height: 45, Hidden: false}); err != nil {
		return err
	}

	row := make([]interface{}, 0, len(meta))
	for i, structs := range data {
		for _, field := range meta {
			pbvalue := structs.Fields[field.Name]
			value, err := convertToXLSXValue(pbvalue)
			if err != nil {
				return err
			}

			row = append(row, value)
		}

		cell, err := excelize.CoordinatesToCellName(1, i+2) // 1-based, and +1 for headers
		if err != nil {
			return err
		}

		if err := sw.SetRow(cell, row); err != nil {
			return err
		}

		row = row[:0]
	}

	if err := sw.Flush(); err != nil {
		return err
	}

	err = f.Write(writer)

	return err
}

func writeParquet(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, ioWriter io.Writer) error {
	fields := make([]arrow.Field, 0, len(meta))
	for _, f := range meta {
		arrowField := arrow.Field{}
		arrowField.Name = f.Name
		typeCode := runtimev1.Type_Code(runtimev1.Type_Code_value[f.Type])
		switch typeCode {
		case runtimev1.Type_CODE_BOOL:
			arrowField.Type = arrow.FixedWidthTypes.Boolean
		case runtimev1.Type_CODE_INT8:
			arrowField.Type = arrow.PrimitiveTypes.Int8
		case runtimev1.Type_CODE_INT16:
			arrowField.Type = arrow.PrimitiveTypes.Int16
		case runtimev1.Type_CODE_INT32:
			arrowField.Type = arrow.PrimitiveTypes.Int32
		case runtimev1.Type_CODE_INT64:
			arrowField.Type = arrow.PrimitiveTypes.Int64
		case runtimev1.Type_CODE_INT128:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_UINT8:
			arrowField.Type = arrow.PrimitiveTypes.Uint8
		case runtimev1.Type_CODE_UINT16:
			arrowField.Type = arrow.PrimitiveTypes.Uint16
		case runtimev1.Type_CODE_UINT32:
			arrowField.Type = arrow.PrimitiveTypes.Uint32
		case runtimev1.Type_CODE_UINT64:
			arrowField.Type = arrow.PrimitiveTypes.Uint64
		case runtimev1.Type_CODE_DECIMAL:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_FLOAT32:
			arrowField.Type = arrow.PrimitiveTypes.Float32
		case runtimev1.Type_CODE_FLOAT64:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_STRUCT, runtimev1.Type_CODE_UUID, runtimev1.Type_CODE_ARRAY, runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_MAP:
			arrowField.Type = arrow.BinaryTypes.String
		case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIME:
			arrowField.Type = arrow.FixedWidthTypes.Timestamp_us
		case runtimev1.Type_CODE_BYTES:
			arrowField.Type = arrow.BinaryTypes.Binary
		}
		fields = append(fields, arrowField)
	}
	schema := arrow.NewSchema(fields, nil)

	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	recordBuilder := array.NewRecordBuilder(mem, schema)
	defer recordBuilder.Release()
	for _, s := range data {
		for idx, t := range meta {
			v := s.Fields[t.Name]
			typeCode := runtimev1.Type_Code(runtimev1.Type_Code_value[t.Type])
			switch typeCode {
			case runtimev1.Type_CODE_BOOL:
				recordBuilder.Field(idx).(*array.BooleanBuilder).Append(v.GetBoolValue())
			case runtimev1.Type_CODE_INT8:
				recordBuilder.Field(idx).(*array.Int8Builder).Append(int8(v.GetNumberValue()))
			case runtimev1.Type_CODE_INT16:
				recordBuilder.Field(idx).(*array.Int16Builder).Append(int16(v.GetNumberValue()))
			case runtimev1.Type_CODE_INT32:
				recordBuilder.Field(idx).(*array.Int32Builder).Append(int32(v.GetNumberValue()))
			case runtimev1.Type_CODE_INT64:
				recordBuilder.Field(idx).(*array.Int64Builder).Append(int64(v.GetNumberValue()))
			case runtimev1.Type_CODE_UINT8:
				recordBuilder.Field(idx).(*array.Uint8Builder).Append(uint8(v.GetNumberValue()))
			case runtimev1.Type_CODE_UINT16:
				recordBuilder.Field(idx).(*array.Uint16Builder).Append(uint16(v.GetNumberValue()))
			case runtimev1.Type_CODE_UINT32:
				recordBuilder.Field(idx).(*array.Uint32Builder).Append(uint32(v.GetNumberValue()))
			case runtimev1.Type_CODE_UINT64:
				recordBuilder.Field(idx).(*array.Uint64Builder).Append(uint64(v.GetNumberValue()))
			case runtimev1.Type_CODE_INT128:
				recordBuilder.Field(idx).(*array.Float64Builder).Append(v.GetNumberValue())
			case runtimev1.Type_CODE_FLOAT32:
				recordBuilder.Field(idx).(*array.Float32Builder).Append(float32(v.GetNumberValue()))
			case runtimev1.Type_CODE_FLOAT64, runtimev1.Type_CODE_DECIMAL:
				recordBuilder.Field(idx).(*array.Float64Builder).Append(v.GetNumberValue())
			case runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_UUID:
				recordBuilder.Field(idx).(*array.StringBuilder).Append(v.GetStringValue())
			case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIME:
				tmp, err := arrow.TimestampFromString(v.GetStringValue(), arrow.Microsecond)
				if err != nil {
					return err
				}

				recordBuilder.Field(idx).(*array.TimestampBuilder).Append(tmp)
			case runtimev1.Type_CODE_ARRAY, runtimev1.Type_CODE_MAP, runtimev1.Type_CODE_STRUCT:
				bts, err := protojson.Marshal(v)
				if err != nil {
					return err
				}

				recordBuilder.Field(idx).(*array.StringBuilder).Append(string(bts))
			}
		}
	}

	parquetwriter, err := pqarrow.NewFileWriter(schema, ioWriter, nil, pqarrow.ArrowWriterProperties{})
	if err != nil {
		return err
	}

	defer parquetwriter.Close()

	rec := recordBuilder.NewRecord()
	err = parquetwriter.Write(rec)
	return err
}

func duckDBCopyExport(ctx context.Context, w io.Writer, opts *runtime.ExportOptions, sql string, args []any, filename string, olap drivers.OLAPStore, exportFormat runtimev1.ExportFormat) error {
	var extension string
	switch exportFormat {
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		extension = "parquet"
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		extension = "csv"
	}

	tmpPath := fmt.Sprintf("export_%s.%s", uuid.New().String(), extension)
	tmpPath = filepath.Join(os.TempDir(), tmpPath)
	defer os.Remove(tmpPath)

	sql = fmt.Sprintf("COPY (%s) TO '%s'", sql, tmpPath)
	if extension == "csv" {
		sql += " (FORMAT CSV, HEADER)"
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         opts.Priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(filename)
		if err != nil {
			return err
		}
	}

	f, err := os.Open(tmpPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

func (q *MetricsViewRows) generateFilename(mv *runtimev1.MetricsViewSpec) string {
	filename := strings.ReplaceAll(mv.Table, `"`, `_`)
	if q.TimeStart != nil || q.TimeEnd != nil || q.Where != nil {
		filename += "_filtered"
	}
	return filename
}
