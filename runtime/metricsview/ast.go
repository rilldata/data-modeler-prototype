package metricsview

import (
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid/druidsqldriver"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

// AST is the abstract syntax tree for a metrics SQL query.
type AST struct {
	// Root of the AST
	Root *SelectNode
	// List of CTEs to add to the query
	CTEs []*SelectNode

	// Cached internal state for building the AST
	underlyingTable     *string
	underlyingWhere     *ExprNode
	dimFields           []FieldNode
	comparisonDimFields []FieldNode
	unnests             []string
	nextIdentifier      int

	// Contextual info for building the AST
	metricsView *runtimev1.MetricsViewSpec
	security    *runtime.ResolvedSecurity
	query       *Query
	dialect     drivers.Dialect
	olapContext map[string]any
}

// SelectNode represents a query that computes measures by dimensions.
// The from/join clauses are not all compatible. The allowed combinations are:
//   - FromTable
//   - FromSelect and optionally SpineSelect and/or LeftJoinSelects
//   - FromSelect and optionally JoinComparisonSelect (for comparison CTE based optimization, this combination is used, both should be set and one of them will be used as CTE)
type SelectNode struct {
	Alias                string           // Alias for the node used by outer SELECTs to reference it.
	IsCTE                bool             // Whether this node is a Common Table Expression
	DimFields            []FieldNode      // Dimensions fields to select
	MeasureFields        []FieldNode      // Measure fields to select
	FromTable            *string          // Underlying table expression to select from (if set, FromSelect must not be set)
	FromSelect           *SelectNode      // Sub-select to select from (if set, FromTable must not be set)
	SpineSelect          *SelectNode      // Sub-select that returns a spine of dimensions. Currently it will be right-joined onto FromSelect.
	LeftJoinSelects      []*SelectNode    // Sub-selects to left join onto FromSelect, to enable "per-dimension" measures
	JoinComparisonSelect *SelectNode      // Sub-select to join onto FromSelect for comparison measures
	JoinComparisonType   JoinType         // Type of join to use for JoinComparisonSelect
	Unnests              []string         // Unnest expressions to add in the FROM clause
	Group                bool             // Whether the SELECT is grouped. If yes, it will group by all DimFields.
	Where                *ExprNode        // Expression for the WHERE clause
	TimeWhere            *ExprNode        // Expression for the time range to add to the WHERE clause
	Having               *ExprNode        // Expression for the HAVING clause. If HAVING is not allowed in the current context, it will added as a WHERE in a wrapping SELECT.
	OrderBy              []OrderFieldNode // Fields to order by
	Limit                *int64           // Limit for the query
	Offset               *int64           // Offset for the query
	FromCrossSelect      *SelectNode      // sub-select containing cross join to select from (only one of FromTable, FromSelect, FromCrossSelect can be set)
	CrossJoinSelects     []*SelectNode    // sub-selects containing cross joins of FromCrossSelect
	UnionAllSelects      []*SelectNode    // Selects to union all, if this is set all other selects are ignored
}

// FieldNode represents a column in a SELECT clause. It also carries metadata related to the dimension/measure it was derived from.
// The Name must always match a the name of a dimension/measure in the metrics view or a computed field specified in the request.
// This means that if two columns in different places in the AST have the same Name, they're guaranteed to resolve to the same value.
type FieldNode struct {
	Name        string
	DisplayName string
	Expr        string
	AutoUnnest  bool
	TreatNullAs string // only used for measures
}

// ExprNode represents an expression for a WHERE clause.
type ExprNode struct {
	Expr string
	Args []any
}

// and returns a new node that is the AND of the current node and the given expression.
func (n *ExprNode) and(expr string, args []any) *ExprNode {
	if expr == "" {
		return n
	}

	if n == nil || n.Expr == "" {
		return &ExprNode{
			Expr: expr,
			Args: args,
		}
	}

	return &ExprNode{
		Expr: fmt.Sprintf("(%s) AND (%s)", n.Expr, expr),
		Args: append(n.Args, args...),
	}
}

// OrderFieldNode represents a field in an ORDER BY clause.
type OrderFieldNode struct {
	Name string
	Desc bool
}

// JoinType represents types of SQL joins.
type JoinType string

const (
	JoinTypeUnspecified JoinType = ""
	JoinTypeFull        JoinType = "FULL OUTER"
	JoinTypeLeft        JoinType = "LEFT OUTER"
	JoinTypeRight       JoinType = "RIGHT OUTER"
)

// NewAST builds a new SQL AST based on a metrics query.
//
// Dynamic time ranges in the qry must be resolved to static start/end timestamps before calling this function.
// This is due to NewAST not being able (or intended) to resolve external time anchors such as watermarks.
//
// The qry's PivotOn must be empty. Pivot queries must be rewritten/handled upstream of NewAST.
func NewAST(mv *runtimev1.MetricsViewSpec, sec *runtime.ResolvedSecurity, qry *Query, dialect drivers.Dialect) (*AST, error) {
	// Validation
	if len(qry.PivotOn) > 0 {
		return nil, errors.New("cannot build AST for pivot queries")
	}
	if len(qry.Dimensions) == 0 && len(qry.Measures) == 0 {
		return nil, fmt.Errorf("must specify at least one dimension or measure")
	}

	// Init
	ast := &AST{
		metricsView: mv,
		security:    sec,
		query:       qry,
		dialect:     dialect,
	}

	// Determine the minimum time grain for the time dimension in the query.
	minGrain := TimeGrainUnspecified
	for _, qd := range ast.query.Dimensions {
		if qd.Compute == nil || qd.Compute.TimeFloor == nil {
			continue
		}
		if !strings.EqualFold(qd.Compute.TimeFloor.Dimension, ast.metricsView.TimeDimension) {
			continue
		}
		tg := qd.Compute.TimeFloor.Grain
		if tg == TimeGrainUnspecified {
			continue
		}
		if minGrain == TimeGrainUnspecified {
			minGrain = tg
			continue
		}
		if tg.ToTimeutil() < minGrain.ToTimeutil() {
			minGrain = tg
		}
	}

	// Build dimensions to apply against the underlying SELECT.
	// We cache these in the AST type because when resolving expressions and adding new JOINs, we need the ability to reference these.
	ast.dimFields = make([]FieldNode, 0, len(ast.query.Dimensions))
	ast.comparisonDimFields = make([]FieldNode, 0, len(ast.query.Dimensions))
	for _, qd := range ast.query.Dimensions {
		dim, err := ast.resolveDimension(qd, true)
		if err != nil {
			return nil, fmt.Errorf("invalid dimension %q: %w", qd.Name, err)
		}

		f := FieldNode{
			Name:        dim.Name,
			DisplayName: dim.DisplayName,
			Expr:        ast.dialect.MetricsViewDimensionExpression(dim),
		}

		if dim.Unnest {
			unnestAlias := ast.generateIdentifier()

			tblWithAlias, auto, err := ast.dialect.LateralUnnest(f.Expr, unnestAlias, f.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to unnest field %q: %w", f.Name, err)
			}

			if auto {
				f.Expr = ast.dialect.AutoUnnest(f.Expr)
				f.AutoUnnest = true
			} else {
				ast.unnests = append(ast.unnests, tblWithAlias)
				f.Expr = ast.sqlForMember(unnestAlias, f.Name)
			}
		}

		// If a comparison time range is provided and the time dimension is in DimFields,
		// we need to add the time interval between the base and comparison time ranges to the time dimension expression in the comparison sub-query.
		// This makes the time dimension values comparable across the base and comparison select, so they can be joined on.
		// Note that estimating the time interval between the two time ranges is best effort and may not always be possible.
		// Also note that the comparison time range currently always targets a.metricsView.TimeDimension, so we only apply the correction for that dimension.
		cf := f // Clone
		if ast.query.ComparisonTimeRange != nil && qd.Compute != nil && qd.Compute.TimeFloor != nil {
			if strings.EqualFold(qd.Compute.TimeFloor.Dimension, ast.metricsView.TimeDimension) {
				cf.Expr, err = ast.sqlForExpressionAdjustedByComparisonTimeRangeOffset(f.Expr, qd.Compute.TimeFloor.Grain, minGrain)
				if err != nil {
					return nil, err
				}
			}
		}

		ast.dimFields = append(ast.dimFields, f)
		ast.comparisonDimFields = append(ast.comparisonDimFields, cf)
	}

	// Build underlying SELECT
	tbl := ast.dialect.EscapeTable(mv.Database, mv.DatabaseSchema, mv.Table)
	where, err := ast.buildUnderlyingWhere()
	if err != nil {
		return nil, err
	}
	ast.underlyingTable = &tbl
	ast.underlyingWhere = where

	// Build initial root node (empty query against the base select)
	n, err := ast.buildBaseSelect(ast.generateIdentifier(), false)
	if err != nil {
		return nil, err
	}
	ast.Root = n

	// Incrementally add each output measure.
	// As each measure is added, the AST is transformed to accommodate it based on its type.
	for _, qm := range ast.query.Measures {
		m, err := ast.resolveMeasure(qm, true)
		if err != nil {
			return nil, fmt.Errorf("invalid measure %q: %w", qm.Name, err)
		}

		err = ast.addMeasureField(ast.Root, m)
		if err != nil {
			return nil, fmt.Errorf("can't query measure %q: %w", qm.Name, err)
		}
	}

	// Handle Having. If the root node is grouped, we add it as a HAVING clause, otherwise wrap it in a SELECT and add it as a WHERE clause.
	if ast.query.Having != nil {
		// We need to wrap in a new SELECT because a WHERE/HAVING clause cannot apply directly to a field with a window function.
		// This also enables us to template the field name instead of the field expression into the expression.
		ast.wrapSelect(ast.Root, ast.generateIdentifier())

		expr, args, err := ast.sqlForExpression(ast.query.Having, ast.Root, true, true)
		if err != nil {
			return nil, fmt.Errorf("failed to compile 'having': %w", err)
		}
		res := &ExprNode{
			Expr: expr,
			Args: args,
		}

		ast.Root.Where = res
	}

	// Incrementally add each sort criterion.
	for _, s := range ast.query.Sort {
		err = ast.addOrderField(ast.Root, s.Name, s.Desc)
		if err != nil {
			return nil, fmt.Errorf("can't sort by %q: %w", s.Name, err)
		}
	}

	// Add limit and offset
	ast.Root.Limit = ast.query.Limit
	ast.Root.Offset = ast.query.Offset

	return ast, nil
}

// resolveDimension returns a dimension spec for the given dimension query.
// If the dimension query specifies a computed dimension, it constructs a dimension spec to match it.
func (a *AST) resolveDimension(qd Dimension, visible bool) (*runtimev1.MetricsViewSpec_DimensionV2, error) {
	// Handle regular dimension
	if qd.Compute == nil {
		return a.lookupDimension(qd.Name, visible)
	}

	// Handle computed dimension. This means "compute.time_floor" must be configured.

	if qd.Compute.TimeFloor == nil {
		return nil, errors.New(`unsupported "compute"`)
	}

	if !qd.Compute.TimeFloor.Grain.Valid() {
		return nil, errors.New(`invalid "grain"`)
	}
	if qd.Compute.TimeFloor.Grain == TimeGrainUnspecified {
		return nil, errors.New(`"grain" must be specified for time floor`)
	}

	dim, err := a.lookupDimension(qd.Compute.TimeFloor.Dimension, visible)
	if err != nil {
		return nil, err
	}

	if qd.Name != qd.Compute.TimeFloor.Dimension {
		err := a.checkNameForComputedField(qd.Name)
		if err != nil {
			return nil, err
		}
	}

	grain := qd.Compute.TimeFloor.Grain.ToProto()
	expr, err := a.dialect.DateTruncExpr(dim, grain, a.query.TimeZone, int(a.metricsView.FirstDayOfWeek), int(a.metricsView.FirstMonthOfYear))
	if err != nil {
		return nil, fmt.Errorf(`failed to compute time floor: %w`, err)
	}

	displayName := dim.DisplayName
	if displayName == "" {
		displayName = qd.Name
	}
	displayName = fmt.Sprintf("%s (%s)", displayName, qd.Compute.TimeFloor.Grain)

	return &runtimev1.MetricsViewSpec_DimensionV2{
		Name:        qd.Name,
		Expression:  expr,
		DisplayName: displayName,
		Unnest:      dim.Unnest,
	}, nil
}

// resolveMeasure returns a measure spec for the given measure query.
// If the measure query specifies a computed measure, it constructs a measure spec to match it.
func (a *AST) resolveMeasure(qm Measure, visible bool) (*runtimev1.MetricsViewSpec_MeasureV2, error) {
	if qm.Compute == nil {
		return a.lookupMeasure(qm.Name, visible)
	}

	if err := qm.Compute.Validate(); err != nil {
		return nil, fmt.Errorf(`invalid "compute": %w`, err)
	}

	err := a.checkNameForComputedField(qm.Name)
	if err != nil {
		return nil, err
	}

	if qm.Compute.Count {
		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:        qm.Name,
			Expression:  "COUNT(*)",
			Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			DisplayName: "Count",
		}, nil
	}

	if qm.Compute.CountDistinct != nil {
		dim, err := a.lookupDimension(qm.Compute.CountDistinct.Dimension, visible)
		if err != nil {
			return nil, err
		}

		expr := dim.Expression
		if expr == "" {
			expr = a.dialect.EscapeIdentifier(dim.Column)
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:        qm.Name,
			Expression:  fmt.Sprintf("COUNT(DISTINCT %s)", expr),
			Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			DisplayName: fmt.Sprintf("Unique %s", dim.DisplayName),
		}, nil
	}

	if qm.Compute.ComparisonValue != nil {
		m, err := a.lookupMeasure(qm.Compute.ComparisonValue.Measure, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("comparison.%s", a.dialect.EscapeIdentifier(m.Name)),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{qm.Compute.ComparisonValue.Measure},
			DisplayName:        fmt.Sprintf("%s (prev)", m.DisplayName),
		}, nil
	}

	if qm.Compute.ComparisonDelta != nil {
		m, err := a.lookupMeasure(qm.Compute.ComparisonDelta.Measure, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("base.%s - comparison.%s", a.dialect.EscapeIdentifier(m.Name), a.dialect.EscapeIdentifier(m.Name)),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{qm.Compute.ComparisonDelta.Measure},
			DisplayName:        fmt.Sprintf("%s (Δ)", m.DisplayName),
		}, nil
	}

	if qm.Compute.ComparisonRatio != nil {
		m, err := a.lookupMeasure(qm.Compute.ComparisonRatio.Measure, visible)
		if err != nil {
			return nil, err
		}

		base := fmt.Sprintf("base.%s", a.dialect.EscapeIdentifier(m.Name))
		comp := fmt.Sprintf("comparison.%s", a.dialect.EscapeIdentifier(m.Name))
		expr := a.dialect.SafeDivideExpression(fmt.Sprintf("%s - %s", base, comp), comp)

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         expr,
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{qm.Compute.ComparisonRatio.Measure},
			DisplayName:        fmt.Sprintf("%s (Δ%%)", m.DisplayName),
		}, nil
	}

	if qm.Compute.PercentOfTotal != nil {
		if qm.Compute.PercentOfTotal.Total == nil {
			return nil, fmt.Errorf("totals not computed for %s", qm.Name)
		}

		m, err := a.lookupMeasure(qm.Compute.PercentOfTotal.Measure, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("%s/%#f", a.dialect.EscapeIdentifier(m.Name), *qm.Compute.PercentOfTotal.Total),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED,
			ReferencedMeasures: []string{qm.Compute.PercentOfTotal.Measure},
			DisplayName:        fmt.Sprintf("%s (Σ%%)", m.DisplayName),
		}, nil
	}

	if qm.Compute.URI != nil {
		dim, err := a.lookupDimension(qm.Compute.URI.Dimension, visible)
		if err != nil {
			return nil, err
		}

		uri := dim.Uri
		if uri == "" {
			return nil, fmt.Errorf("`uri` not set for the dimension %v", qm.Compute.URI.Dimension)
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:        qm.Name,
			Expression:  a.sqlForAnyInGroup(uri),
			Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			DisplayName: fmt.Sprintf("URI for %s", dim.DisplayName),
		}, nil
	}

	return nil, errors.New("unhandled compute operation")
}

// lookupDimension finds a dimension spec in the metrics view.
// If visible is true, it returns an error if the security policy does not grant access to the dimension.
func (a *AST) lookupDimension(name string, visible bool) (*runtimev1.MetricsViewSpec_DimensionV2, error) {
	if name == "" {
		return nil, errors.New("received empty dimension name")
	}

	if name == a.metricsView.TimeDimension {
		return &runtimev1.MetricsViewSpec_DimensionV2{
			Name:   name,
			Column: name,
		}, nil
	}

	if visible {
		if !a.security.CanAccessField(name) {
			return nil, runtime.ErrForbidden
		}
	}

	for _, dim := range a.metricsView.Dimensions {
		if dim.Name == name {
			return dim, nil
		}
	}

	return nil, fmt.Errorf("dimension %q not found", name)
}

// lookupMeasure finds a measure spec in the metrics view.
// If visible is true, it returns an error if the security policy does not grant access to the measure.
func (a *AST) lookupMeasure(name string, visible bool) (*runtimev1.MetricsViewSpec_MeasureV2, error) {
	if visible {
		if !a.security.CanAccessField(name) {
			return nil, runtime.ErrForbidden
		}
	}

	for _, m := range a.metricsView.Measures {
		if m.Name == name {
			return m, nil
		}
	}

	return nil, fmt.Errorf("measure %q not found", name)
}

// checkNameForComputedField checks that the name for a computed field does not collide with an existing dimension or measure name.
// (This is necessary because even if the other name is not used in the query, it might be referenced by a derived measure.)
func (a *AST) checkNameForComputedField(name string) error {
	if name == "" {
		return errors.New("name for computed field is empty")
	}

	if name == a.metricsView.TimeDimension {
		return errors.New("name for computed field collides with the time dimension name")
	}

	for _, d := range a.metricsView.Dimensions {
		if d.Name == name {
			return errors.New("name for computed field collides with an existing dimension name")
		}
	}

	for _, m := range a.metricsView.Measures {
		if m.Name == name {
			return errors.New("name for computed field collides with an existing measure name")
		}
	}

	return nil
}

// checkRequiredDimensionsPresentInQuery checks that all the dimensions required by the measure are present in the query.
// It checks against the query's dimensions because:
//  1. This enables correctly checking against the underlying time dimension name for dimensions with time floor applied.
//  2. The query's dimensions are projected into every sub-query, so it's not necessary to check against the current sub-query's Dimensions.
func (a *AST) checkRequiredDimensionsPresentInQuery(m *runtimev1.MetricsViewSpec_MeasureV2) error {
	for _, rd := range m.RequiredDimensions {
		var found bool
		for _, qd := range a.query.Dimensions {
			// Handle computed dimension
			if qd.Compute != nil {
				if qd.Compute.TimeFloor == nil {
					continue
				}

				if rd.Name != qd.Compute.TimeFloor.Dimension {
					continue
				}

				isGrainSpecified := rd.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
				isGrainDifferent := TimeGrainFromProto(rd.TimeGrain) != qd.Compute.TimeFloor.Grain
				if isGrainSpecified && isGrainDifferent {
					continue
				}

				found = true
				break
			}

			// Check for time dimension with time floor applied
			if rd.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				if qd.Compute == nil || qd.Compute.TimeFloor == nil {
					continue
				}

				if TimeGrainFromProto(rd.TimeGrain) != qd.Compute.TimeFloor.Grain {
					continue
				}

				if rd.Name == qd.Compute.TimeFloor.Dimension {
					found = true
					break
				}
			}

			// Checking for regular dimension
			if qd.Compute != nil {
				continue
			}
			if rd.Name == qd.Name {
				found = true
				break
			}
		}

		if !found {
			if rd.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				return fmt.Errorf("missing required dimension %q at %q granularity", rd.Name, TimeGrainFromProto(rd.TimeGrain))
			}
			return fmt.Errorf("missing required dimension %q", rd.Name)
		}
	}

	return nil
}

// buildUnderlyingWhere constructs the base WHERE clause for the query.
func (a *AST) buildUnderlyingWhere() (*ExprNode, error) {
	var res *ExprNode

	expr, args, err := a.sqlForExpression(a.query.Where, nil, false, true)
	if err != nil {
		return nil, fmt.Errorf("failed to compile 'where': %w", err)
	}
	res = res.and(expr, args)

	if qf := a.security.QueryFilter(); qf != nil {
		e := NewExpressionFromProto(qf)
		expr, args, err = a.sqlForExpression(e, nil, false, false)
		if err != nil {
			return nil, fmt.Errorf("failed to compile the security policy's query filter: %w", err)
		}
		res = res.and(expr, args)
	}

	if rf := a.security.RowFilter(); rf != "" {
		res = res.and(rf, nil)
	}

	return res, nil
}

// buildBaseSelect constructs a base SELECT node against the underlying table.
func (a *AST) buildBaseSelect(alias string, comparison bool) (*SelectNode, error) {
	n := &SelectNode{
		Alias:     alias,
		DimFields: a.dimFields,
		Unnests:   a.unnests,
		Group:     true,
		FromTable: a.underlyingTable,
		Where:     a.underlyingWhere,
	}

	tr := a.query.TimeRange
	if comparison {
		n.DimFields = a.comparisonDimFields
		tr = a.query.ComparisonTimeRange
	}

	a.addTimeRange(n, tr)

	// If there is a spine, we wrap the base SELECT in a new SELECT that we add the spine to.
	// We do not join the spine directly to the FromTable because the join would be evaluated before the GROUP BY,
	// which would impact the measure aggregations (e.g. counts per group would be wrong).
	if a.query.Spine != nil {
		sn, err := a.buildSpineSelect(a.generateIdentifier(), a.query.Spine, tr)
		if err != nil {
			return nil, err
		}
		if sn == nil {
			return n, nil
		}

		a.wrapSelect(n, a.generateIdentifier())
		n.SpineSelect = sn

		// Update the dimension fields to derive from the SpineSelect instead of the FromSelect
		// (since by definition, some dimension values in the spine might not be present in FromSelect).
		for i, f := range n.DimFields {
			f.Expr = a.sqlForMember(sn.Alias, f.Name)
			n.DimFields[i] = f
		}
	}

	return n, nil
}

// buildSpineSelect constructs a SELECT node for the given spine of dimension values.
func (a *AST) buildSpineSelect(alias string, spine *Spine, tr *TimeRange) (*SelectNode, error) {
	if spine == nil {
		return nil, nil
	}

	if spine.Where != nil {
		expr, args, err := a.sqlForExpression(spine.Where.Expression, nil, false, true)
		if err != nil {
			return nil, fmt.Errorf("failed to compile 'spine.where': %w", err)
		}

		n := &SelectNode{
			Alias:     alias,
			DimFields: a.dimFields,
			Unnests:   a.unnests,
			Group:     true,
			FromTable: a.underlyingTable,
		}
		n.Where = n.Where.and(expr, args)
		a.addTimeRange(n, tr)

		return n, nil
	}

	if spine.TimeRange != nil {
		if a.dialect == drivers.DialectDruid {
			if a.olapContext == nil {
				a.olapContext = make(map[string]any)
			}
			a.olapContext[druidsqldriver.SkipEmptyBucketsContextKey] = false
			return nil, nil
		}

		// if spine generates more than 1000 values then return an error
		bins := timeutil.ApproximateBins(spine.TimeRange.Start, spine.TimeRange.End, spine.TimeRange.Grain.ToTimeutil())
		if bins > 1000 {
			return nil, errors.New("time spine generates more than 1000 values")
		}

		var rangeSelect *SelectNode
		if a.dialect == drivers.DialectDuckDB {
			from := fmt.Sprintf("range('%s'::TIMESTAMP, '%s'::TIMESTAMP, INTERVAL '1 %s')", spine.TimeRange.Start.Format(time.RFC3339), spine.TimeRange.End.Format(time.RFC3339), spine.TimeRange.Grain)
			rangeSelect = &SelectNode{
				Alias: alias,
				DimFields: []FieldNode{
					{
						Name:        spine.TimeRange.Alias,
						DisplayName: spine.TimeRange.Alias,
						Expr:        "range",
					},
				},
				FromTable: &from,
			}
		} else if a.dialect == drivers.DialectClickHouse {
			rangeSelect = &SelectNode{
				Alias: alias,
			}

			// ClickHouse does not support range() function, so we need to generate a series of timestamps
			for t := spine.TimeRange.Start; t.Before(spine.TimeRange.End); t = timeutil.AddTime(t, spine.TimeRange.Grain.ToTimeutil(), 1) {
				rangeSelect.UnionAllSelects = append(rangeSelect.UnionAllSelects, &SelectNode{
					Alias: "", // alias is not needed for union all selects
					DimFields: []FieldNode{
						{
							Name:        spine.TimeRange.Alias,
							DisplayName: spine.TimeRange.Alias,
							Expr:        fmt.Sprintf("'%s'::DateTime64", t.Format("2006-01-02T15:04:05.000")),
						},
					},
				})
			}
		} else {
			return nil, errors.New("unsupported dialect")
		}
		if len(a.dimFields) == 1 {
			return rangeSelect, nil
		}

		var newDims []FieldNode
		for _, f := range a.dimFields {
			if f.Name != spine.TimeRange.Alias {
				newDims = append(newDims, f)
			}
		}

		rangeSelect.Alias = a.generateIdentifier()

		dimSelect := &SelectNode{
			Alias:     a.generateIdentifier(),
			DimFields: newDims,
			FromTable: a.underlyingTable,
			Where:     a.underlyingWhere,
			Group:     true,
		}

		a.addTimeRange(dimSelect, tr)

		// there are other dimensions in the query, so cross join the spine time range with the other dimensions
		crossSelect := &SelectNode{
			Alias:            alias,
			CrossJoinSelects: []*SelectNode{dimSelect, rangeSelect},
		}

		a.wrapSelect(crossSelect, "")

		return crossSelect, nil
	}

	return nil, errors.New("unhandled spine type")
}

// addTimeRange adds a time range to the given SelectNode's WHERE clause.
func (a *AST) addTimeRange(n *SelectNode, tr *TimeRange) {
	if tr == nil || tr.IsZero() || a.metricsView.TimeDimension == "" {
		return
	}

	// Since resolving time ranges may require contextual info (like watermarks), the upstream caller is responsible for resolving them.
	if tr.Start.IsZero() && tr.End.IsZero() {
		panic("ast received a non-empty, unresolved time range")
	}

	expr, args := a.sqlForTimeRange(a.metricsView.TimeDimension, tr.Start, tr.End)
	n.TimeWhere = &ExprNode{
		Expr: expr,
		Args: args,
	}
}

// addMeasureField adds a measure field to the given SelectNode.
// Depending on the measure type, it may rewrite the SelectNode to accommodate the measure.
func (a *AST) addMeasureField(n *SelectNode, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	// Skip if the measure has already been added.
	// This can happen if the measure was already added as a referenced measure of a derived measure.
	if hasMeasure(n, m.Name) {
		return nil
	}

	// Check that the measure's required dimensions are satisfied
	err := a.checkRequiredDimensionsPresentInQuery(m)
	if err != nil {
		return err
	}

	switch m.Type {
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_UNSPECIFIED, runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE:
		_, err = a.addSimpleMeasure(n, m)
		return err
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED:
		_, err = a.addDerivedMeasure(n, m)
		return err
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON:
		_, err = a.addTimeComparisonMeasure(n, m)
		return err
	default:
		panic("unhandled measure type")
	}
}

// addSimpleMeasure adds a measure of type simple to the given SelectNode.
// When called, we know the measure is not present in the SelectNode, but it might be present in a sub-select.
func (a *AST) addSimpleMeasure(n *SelectNode, m *runtimev1.MetricsViewSpec_MeasureV2) (*FieldNode, error) {
	// Base case: it targets the underlying table.
	// Add the measure directly to the SELECT list.
	if n.FromTable != nil {
		expr, err := a.sqlForMeasure(m, n)
		if err != nil {
			return nil, err
		}

		n.MeasureFields = append(n.MeasureFields, FieldNode{
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Expr:        expr,
			TreatNullAs: m.TreatNullsAs,
		})

		return &n.MeasureFields[len(n.MeasureFields)-1], nil
	}

	// Recursive case: it targets another SelectNode.
	// We recurse on the sub-select and add a pass-through field to the current node.

	if !hasMeasure(n.FromSelect, m.Name) { // Don't recurse if already in scope in sub-query
		f, err := a.addSimpleMeasure(n.FromSelect, m)
		if err != nil {
			return nil, err
		}
		f.TreatNullAs = "" // only have to set this at the top level
	}

	expr := a.sqlForMember(n.FromSelect.Alias, m.Name)
	if n.Group {
		expr = a.sqlForAnyInGroup(expr)
	}

	n.MeasureFields = append(n.MeasureFields, FieldNode{
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Expr:        expr,
		TreatNullAs: m.TreatNullsAs,
	})

	return &n.MeasureFields[len(n.MeasureFields)-1], nil
}

// addDerivedMeasure adds a measure of type derived to the given SelectNode.
// When called, we know the measure is not present in the SelectNode, but it might be present in a sub-select.
func (a *AST) addDerivedMeasure(n *SelectNode, m *runtimev1.MetricsViewSpec_MeasureV2) (*FieldNode, error) {
	// Handle derived measures with "per" dimensions separately.
	if len(m.PerDimensions) > 0 {
		return a.addDerivedMeasureWithPer(n, m)
	}

	// If the current node has a comparison join, push calculation of the derived measure into its FromSelect and add a pass-through field in the current node.
	// This avoids a potential ambiguity issue because the derived measure expression does not use "base.name" and "comparison.name" to identify referenced measures,
	// so we need to ensure the referenced names exist only in ONE sub-query.
	if n.JoinComparisonSelect != nil {
		if !hasMeasure(n.FromSelect, m.Name) {
			f, err := a.addDerivedMeasure(n.FromSelect, m)
			if err != nil {
				return nil, err
			}
			f.TreatNullAs = "" // only have to set this at the top level
		}

		expr := a.sqlForMember(n.FromSelect.Alias, m.Name)
		if n.Group {
			expr = a.sqlForAnyInGroup(expr)
		}

		n.MeasureFields = append(n.MeasureFields, FieldNode{
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Expr:        expr,
			TreatNullAs: m.TreatNullsAs,
		})

		return &n.MeasureFields[len(n.MeasureFields)-1], nil
	}
	// Now we know it's NOT a node with a comparison join.

	// If the current node has a spine join, we wrap it in a new SELECT that we add the derived measure to.
	// Even though the spine join won't add any ambiguous measure names, it will make dimension names ambiguous.
	// So in case the derived measure references a dimension by name (unlikely, but possible), we need to ensure the dimension is only in scope in one sub-query.
	if n.SpineSelect != nil {
		a.wrapSelect(n, a.generateIdentifier())
	}
	// Now we know it's ALSO NOT a node with a spine join.

	// If the referenced measures are not already in scope in the sub-selects, add them.
	// Since the node doesn't have a comparison join, addReferencedMeasuresToScope guarantees the referenced measures are ONLY in scope in ONE sub-query, which prevents ambiguous references in the measure expression.
	err := a.addReferencedMeasuresToScope(n, m.ReferencedMeasures)
	if err != nil {
		return nil, err
	}

	// Add the derived measure expression to the current node.
	expr, err := a.sqlForMeasure(m, n)
	if err != nil {
		return nil, err
	}
	if n.Group {
		// TODO: There's a risk of expr containing a window, which can't be wrapped by ANY_VALUE. Need to fix it by wrapping with a non-grouped SELECT. Doesn't matter until we implement addDerivedMeasureWithPer.
		expr = a.sqlForAnyInGroup(expr)
	}

	n.MeasureFields = append(n.MeasureFields, FieldNode{
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Expr:        expr,
		TreatNullAs: m.TreatNullsAs,
	})

	return &n.MeasureFields[len(n.MeasureFields)-1], nil
}

// addDerivedMeasureWithPer adds a measure of type derived with "per" dimensions to the given SelectNode.
// When called, we know the measure is not present in the SelectNode, but it might be present in a sub-select.
func (a *AST) addDerivedMeasureWithPer(_ *SelectNode, _ *runtimev1.MetricsViewSpec_MeasureV2) (*FieldNode, error) {
	return nil, errors.New(`support for "per" not implemented`)
}

// addTimeComparisonMeasure adds a measure of type time comparison to the given SelectNode.
// When called, we know the measure is not present in the SelectNode, but it might be present in a sub-select.
func (a *AST) addTimeComparisonMeasure(n *SelectNode, m *runtimev1.MetricsViewSpec_MeasureV2) (*FieldNode, error) {
	// If the node doesn't have a comparison join, we wrap it in a new SELECT that we add the comparison join to.
	// We use the hardcoded aliases "base" and "comparison" for the two SELECTs (which must be used in the comparison measure expression).
	if n.JoinComparisonSelect == nil {
		if a.query.ComparisonTimeRange == nil {
			return nil, errors.New("comparison time range not provided")
		}

		a.wrapSelect(n, "base")

		csn, err := a.buildBaseSelect("comparison", true)
		if err != nil {
			return nil, err
		}
		n.JoinComparisonSelect = csn

		n.JoinComparisonType = JoinTypeFull

		for i, f := range n.DimFields {
			f.Expr = fmt.Sprintf("COALESCE(%s, %s)", f.Expr, a.sqlForMember("comparison", f.Name))
			n.DimFields[i] = f // Because it's not a value, not a pointer
		}
	}

	// Add the referenced measures to the base and comparison SELECTs.
	err := a.addReferencedMeasuresToScope(n, m.ReferencedMeasures)
	if err != nil {
		return nil, err
	}

	// Add the comparison measure expression to the current node.
	expr, err := a.sqlForMeasure(m, n)
	if err != nil {
		return nil, err
	}
	if n.Group {
		// TODO: There's a risk of expr containing a window, which can't be wrapped by ANY_VALUE. Need to fix it by wrapping with a non-grouped SELECT. Doesn't matter until we implement addDerivedMeasureWithPer.
		// TODO: Can a node with a comparison ever have Group==true?
		expr = a.sqlForAnyInGroup(expr)
	}

	n.MeasureFields = append(n.MeasureFields, FieldNode{
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Expr:        expr,
		TreatNullAs: m.TreatNullsAs,
	})

	return &n.MeasureFields[len(n.MeasureFields)-1], nil
}

// addOrderField adds a sort field to the given SelectNode.
func (a *AST) addOrderField(n *SelectNode, name string, desc bool) error {
	// We currently only allow sorting by selected dimensions and measures.
	if !hasName(n, name) {
		return errors.New("name not present in context")
	}

	n.OrderBy = append(n.OrderBy, OrderFieldNode{
		Name: name,
		Desc: desc,
	})

	return nil
}

// addReferencedMeasuresToScope adds the referenced measures to the node's scope (skipping any that are already in scope).
// If the node has a comparison join, it adds the measures to both the base (FromSelect) and comparison (JoinComparisonSelect) nodes.
// Otherwise, it guarantees that the referenced measures are only in scope in one of the node's sub-select (so no need to worry about ambiguous references).
//
// Note that it does not add the measures to the current node's SELECT list, only to one of its sub-selects such that it's in scope for querying.
func (a *AST) addReferencedMeasuresToScope(n *SelectNode, referencedMeasures []string) error {
	if len(referencedMeasures) == 0 {
		return nil
	}

	// If the node targets the underlying table, we need to add a new level of nesting.
	// This ensures n.FromSelect is set, so we can bring the referenced measures into scope.
	if n.FromTable != nil {
		a.wrapSelect(n, a.generateIdentifier())
	}

	for _, rm := range referencedMeasures {
		// Note we pass visible==false because the measure won't be projected into the current node's SELECT list, only brought into scope for derived measures.
		m, err := a.lookupMeasure(rm, false)
		if err != nil {
			return err
		}

		// Add to the base SELECT. addMeasureField skips it if it's already present.
		err = a.addMeasureField(n.FromSelect, m)
		if err != nil {
			return err
		}

		// Add to the comparison SELECT if it exists.
		if n.JoinComparisonSelect != nil {
			err = a.addMeasureField(n.JoinComparisonSelect, m)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// wrapSelect rewrites the given node with a wrapping SELECT that includes the same dimensions and measures as the original node.
// The innerAlias is used as the alias of the inner SELECT in the new outer SELECT.
// Example: wrapSelect("SELECT a, count(*) as b FROM c", "t") -> "SELECT t.a, t.b FROM (SELECT a, count(*) as b FROM c) t".
func (a *AST) wrapSelect(s *SelectNode, innerAlias string) {
	cpy := *s
	cpy.Alias = innerAlias

	s.DimFields = make([]FieldNode, 0, len(cpy.DimFields))
	for _, f := range cpy.DimFields {
		s.DimFields = append(s.DimFields, FieldNode{
			Name:        f.Name,
			DisplayName: f.DisplayName,
			Expr:        a.sqlForMember(cpy.Alias, f.Name),
		})
	}

	for _, cjs := range cpy.CrossJoinSelects {
		for _, f := range cjs.DimFields {
			s.DimFields = append(s.DimFields, FieldNode{
				Name:        f.Name,
				DisplayName: f.DisplayName,
				Expr:        a.sqlForMember(cpy.Alias, f.Name),
			})
		}

		if len(cjs.UnionAllSelects) > 0 {
			// All dimensions will be same across UNION ALL SELECTS so we can just pick the first one
			for _, f := range cjs.UnionAllSelects[0].DimFields {
				s.DimFields = append(s.DimFields, FieldNode{
					Name:        f.Name,
					DisplayName: f.DisplayName,
					Expr:        a.sqlForMember(cpy.Alias, f.Name),
				})
			}
		}
	}

	s.MeasureFields = make([]FieldNode, 0, len(cpy.MeasureFields))
	for _, f := range cpy.MeasureFields {
		s.MeasureFields = append(s.MeasureFields, FieldNode{
			Name:        f.Name,
			DisplayName: f.DisplayName,
			Expr:        a.sqlForMember(cpy.Alias, f.Name),
			TreatNullAs: "",
		})
	}

	s.FromTable = nil
	if len(cpy.CrossJoinSelects) > 0 {
		s.FromCrossSelect = &cpy
	} else {
		s.FromSelect = &cpy
	}
	s.SpineSelect = nil
	s.LeftJoinSelects = nil
	s.JoinComparisonSelect = nil
	s.JoinComparisonType = JoinTypeUnspecified
	s.Unnests = nil
	s.Group = false
	s.Where = nil
	s.TimeWhere = nil
	s.Having = nil

	s.OrderBy = make([]OrderFieldNode, 0, len(cpy.OrderBy))
	s.OrderBy = append(s.OrderBy, cpy.OrderBy...) // Fresh copy

	s.Limit = nil
	s.Offset = nil
	s.CrossJoinSelects = nil
}

// findFieldForDimension finds the field in the SelectNode that corresponds to the dimension selector.
// It takes computed dimensions into account, comparing against the underlying dimension name instead of the query alias.
func (a *AST) findFieldForDimension(n *SelectNode, dim *runtimev1.MetricsViewSpec_DimensionSelector) (FieldNode, bool) {
	for _, f := range n.DimFields {
		// If name matches, we're done
		if dim.Name == f.Name {
			return f, true
		}

		// Find original query dimension for the field
		var fqd Dimension
		for _, qd := range a.query.Dimensions {
			if f.Name == qd.Name {
				fqd = qd
				break
			}
		}

		// If it's a computed dimension, check against the underlying dimension name (and time grain if specified)
		if fqd.Compute != nil && fqd.Compute.TimeFloor != nil {
			if dim.Name != fqd.Compute.TimeFloor.Dimension {
				continue
			}

			if dim.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED && dim.TimeGrain != fqd.Compute.TimeFloor.Grain.ToProto() {
				continue
			}

			return f, true
		}
	}

	return FieldNode{}, false
}

// generateIdentifier generates a unique table identifier for use in the AST.
func (a *AST) generateIdentifier() string {
	tmp := fmt.Sprintf("t%d", a.nextIdentifier)
	a.nextIdentifier++
	return tmp
}

// sqlForTimeRange builds a SQL expression and query args for filtering by a time range.
func (a *AST) sqlForTimeRange(timeCol string, start, end time.Time) (string, []any) {
	var where string
	var args []any
	if !start.IsZero() && !end.IsZero() {
		col := a.dialect.EscapeIdentifier(timeCol)
		where = fmt.Sprintf("%s >= ? AND %s < ?", col, col)
		args = []any{start, end}
	} else if !start.IsZero() {
		where = fmt.Sprintf("%s >= ?", a.dialect.EscapeIdentifier(timeCol))
		args = []any{start}
	} else if !end.IsZero() {
		where = fmt.Sprintf("%s < ?", a.dialect.EscapeIdentifier(timeCol))
		args = []any{end}
	} else {
		return "", nil
	}
	return where, args
}

// sqlForMeasure builds a SQL expression for a measure, including its window if present.
// It uses the provided n to resolve dimensions expressions for window partitions.
func (a *AST) sqlForMeasure(m *runtimev1.MetricsViewSpec_MeasureV2, n *SelectNode) (string, error) {
	// If not applying a window, just return the measure expression.
	if m.Window == nil {
		return m.Expression, nil
	}

	// If partitioning is not enabled, ordering and framing doesn't matter
	if !m.Window.Partition {
		return fmt.Sprintf("%s OVER ()", m.Expression), nil
	}

	// If partitioning is enabled, we partition by all dimensions that we don't order by.

	// Gather order by fields
	orderFields := make([]FieldNode, 0, len(m.Window.OrderBy))
	orderDesc := make([]bool, 0, len(m.Window.OrderBy))
	for _, d := range m.Window.OrderBy {
		f, ok := a.findFieldForDimension(n, d)
		if !ok {
			// In practice, this should never happen because the OrderBy dimensions are in RequiredDimensions and have been checked before this point.
			return "", fmt.Errorf("dimension %q required by window measure %q not found in query", d.Name, m.Name)
		}
		orderFields = append(orderFields, f)
		orderDesc = append(orderDesc, d.Desc)
	}

	// Gather partition fields
	var partitionFields []FieldNode
	for _, f := range n.DimFields {
		found := false
		for _, of := range orderFields {
			if f.Name == of.Name {
				found = true
				break
			}
		}
		if !found {
			partitionFields = append(partitionFields, f)
		}
	}

	// Build the window expression
	b := &strings.Builder{}
	b.WriteString(m.Expression)
	b.WriteString(" OVER (")
	if len(partitionFields) > 0 {
		b.WriteString("PARTITION BY ")
		for i, f := range partitionFields {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteString(f.Expr)
		}
	}
	if len(orderFields) > 0 {
		if len(partitionFields) > 0 {
			b.WriteString(" ")
		}
		b.WriteString("ORDER BY ")
		for i, f := range orderFields {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteString(f.Expr)
			if orderDesc[i] {
				b.WriteString(" DESC")
			}
		}
	}
	if m.Window.FrameExpression != "" {
		b.WriteString(" ")
		b.WriteString(m.Window.FrameExpression)
	}
	b.WriteString(")")

	return b.String(), nil
}

// sqlForMember builds a SQL expression for a column in a table.
// It does not escape the tbl identifier because we currently only use it for internally generated aliases.
func (a *AST) sqlForMember(tbl, name string) string {
	if tbl == "" {
		return a.dialect.EscapeIdentifier(name)
	}
	return fmt.Sprintf("%s.%s", tbl, a.dialect.EscapeIdentifier(name))
}

// sqlForAnyInGroup returns a SQL expression for passing through a field in a GROUP BY.
func (a *AST) sqlForAnyInGroup(expr string) string {
	return fmt.Sprintf("ANY_VALUE(%s)", expr)
}

// sqlForExpression returns the provided time expression adjusted by the fixed time offset between the current query's base and comparison time ranges.
// The timestamp column (ie a.metricsView.TimeDimension) is expected to be the base timestamp for `expr` (in case of multiple metrics view time dimensions defined).
func (a *AST) sqlForExpressionAdjustedByComparisonTimeRangeOffset(expr string, g, mg TimeGrain) (string, error) {
	if a.query.TimeRange == nil || a.query.TimeRange.Start.IsZero() || a.query.ComparisonTimeRange == nil || a.query.ComparisonTimeRange.Start.IsZero() {
		return "", errors.New("must specify an explicit start time for both the base and comparison time range when comparing by a time dimension")
	}

	start1 := a.query.TimeRange.Start
	start2 := a.query.ComparisonTimeRange.Start

	var dateDiff string
	if g == TimeGrainUnspecified {
		g = TimeGrainMillisecond // todo millis won't work for druid
		res, err := a.dialect.DateDiff(g.ToProto(), start1, start2)
		if err != nil {
			return "", err
		}
		dateDiff = res
	} else if g == mg {
		res, err := a.dialect.DateDiff(g.ToProto(), start1, start2)
		if err != nil {
			return "", err
		}
		dateDiff = res
	} else {
		// larger time grain values can change as well
		res, err := a.dialect.DateDiff(mg.ToProto(), start1, start2)
		if err != nil {
			return "", err
		}
		dateDiff = res

		// DATE_TRUNC('year', t - INTERVAL (DATE_DIFF(start, end)) day)
		tc := a.dialect.EscapeIdentifier(a.metricsView.TimeDimension)
		expr := fmt.Sprintf("(%s - INTERVAL (%s) %s)", tc, dateDiff, a.dialect.ConvertToDateTruncSpecifier(mg.ToProto()))
		dim := &runtimev1.MetricsViewSpec_DimensionV2{
			Expression: expr,
		}
		expr, err = a.dialect.DateTruncExpr(dim, g.ToProto(), a.query.TimeZone, int(a.metricsView.FirstDayOfWeek), int(a.metricsView.FirstMonthOfYear))
		if err != nil {
			return "", fmt.Errorf(`failed to compute time floor: %w`, err)
		}
		return expr, nil
	}

	return fmt.Sprintf("(%s - INTERVAL (%s) %s)", expr, dateDiff, a.dialect.ConvertToDateTruncSpecifier(g.ToProto())), nil
}

// convertToCTE util func that sets IsCTE and only adds to a.CTEs if IsCTE was false
func (a *AST) convertToCTE(n *SelectNode) {
	if n.IsCTE {
		return
	}

	n.IsCTE = true
	a.CTEs = append(a.CTEs, n)
}

// hasName checks if the given name is present as either a dimension or measure field in the node.
// It relies on field names always resolving to the same value regardless of where in the AST they're referenced.
// I.e. a name always corresponds to a dimension/measure name in the metrics view or as a computed field in the query.
// NOTE: Even if it returns false, the measure may still be present in a sub-select (which can happen if it was added as a referenced measure of a derived measure).
func hasName(n *SelectNode, name string) bool {
	for _, f := range n.DimFields {
		if f.Name == name {
			return true
		}
	}
	for _, f := range n.MeasureFields {
		if f.Name == name {
			return true
		}
	}
	return false
}

// hasMeasure checks if the given measure name is already present in the given SelectNode.
// See hasName for details about name checks.
func hasMeasure(n *SelectNode, name string) bool {
	for _, f := range n.MeasureFields {
		if f.Name == name {
			return true
		}
	}
	return false
}
