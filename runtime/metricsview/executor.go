package metricsview

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

const defaultExecutionTimeout = time.Minute * 3

// Executor is capable of executing queries and other operations against a metrics view.
type Executor struct {
	rt          *runtime.Runtime
	instanceID  string
	metricsView *runtimev1.MetricsViewSpec
	security    *runtime.ResolvedMetricsViewSecurity
	priority    int

	olap        drivers.OLAPStore
	olapRelease func()
	instanceCfg drivers.InstanceConfig

	watermark time.Time
}

// NewExecutor creates a new Executor for the provided metrics view.
func NewExecutor(ctx context.Context, rt *runtime.Runtime, instanceID string, mv *runtimev1.MetricsViewSpec, sec *runtime.ResolvedMetricsViewSecurity, priority int) (*Executor, error) {
	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connector for metrics view: %w", err)
	}

	instanceCfg, err := rt.InstanceConfig(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	return &Executor{
		rt:          rt,
		instanceID:  instanceID,
		metricsView: mv,
		security:    sec,
		priority:    priority,
		olap:        olap,
		olapRelease: release,
		instanceCfg: instanceCfg,
	}, nil
}

// Close releases the resources held by the Executor.
func (e *Executor) Close() {
	e.olapRelease()
}

// ValidateMetricsView validates the dimensions and measures in the executor's metrics view.
func (e *Executor) ValidateMetricsView(ctx context.Context) error {
	// TODO: Implement it
	panic("not implemented")
}

// ValidateQuery validates the provided query against the executor's metrics view.
func (e *Executor) ValidateQuery(qry *Query) error {
	// TODO: Implement it
	panic("not implemented")
}

// Watermark returns the current watermark of the metrics view.
// If the watermark resolves to null, it defaults to the current time.
func (e *Executor) Watermark(ctx context.Context) (time.Time, error) {
	return e.loadWatermark(ctx, nil)
}

// Schema returns a schema for the metrics view's dimensions and measures.
func (e *Executor) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	// Build a query that selects all dimensions and measures
	qry := &Query{}

	if e.metricsView.TimeDimension != "" {
		qry.Dimensions = append(qry.Dimensions, Dimension{
			Name: e.metricsView.TimeDimension,
			Compute: &DimensionCompute{
				TimeFloor: &DimensionComputeTimeFloor{
					Dimension: e.metricsView.TimeDimension,
					Grain:     TimeGrainDay,
				},
			},
		})
	}

	for _, d := range e.metricsView.Dimensions {
		qry.Dimensions = append(qry.Dimensions, Dimension{Name: d.Name})
	}

	for _, m := range e.metricsView.Measures {
		qry.Measures = append(qry.Measures, Measure{Name: m.Name})
	}

	// Setting both base and comparison time ranges in case there are time_comparison measures.
	if e.metricsView.TimeDimension != "" {
		now := time.Now()
		qry.TimeRange = &TimeRange{
			Start: now.Add(-time.Second),
			End:   now,
		}
		qry.ComparisonTimeRange = &TimeRange{
			Start: now.Add(-2 * time.Second),
			End:   now.Add(-time.Second),
		}
	}

	// Importantly, limit to 0 rows
	zero := int64(0)
	qry.Limit = &zero

	// Execute the query to get the schema
	ast, err := NewAST(e.metricsView, nil, qry, e.olap.Dialect())
	if err != nil {
		return nil, err
	}

	sql, args, err := ast.SQL()
	if err != nil {
		return nil, err
	}

	res, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return res.Schema, nil
}

// Query executes the provided query against the metrics view.
func (e *Executor) Query(ctx context.Context, qry *Query, executionTime *time.Time) (*drivers.Result, bool, error) {
	if e.security != nil && !e.security.Access {
		return nil, false, runtime.ErrForbidden
	}

	export := qry.Label // TODO: Always set to false once separate exports are implemented
	if err := e.rewriteQueryLimit(qry, export); err != nil {
		return nil, false, err
	}

	if err := e.rewriteQueryTimeRanges(ctx, qry, executionTime); err != nil {
		return nil, false, err
	}

	if err := e.rewriteQueryDruidExactify(ctx, qry); err != nil {
		return nil, false, err
	}

	ast, err := NewAST(e.metricsView, e.security, qry, e.olap.Dialect())
	if err != nil {
		return nil, false, err
	}

	if err := e.rewriteApproximateComparisons(ast); err != nil {
		return nil, false, err
	}

	if err := e.rewriteDruidJoins(ast); err != nil {
		return nil, false, err
	}

	sql, args, err := ast.SQL()
	if err != nil {
		return nil, false, err
	}

	res, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, false, err
	}

	limitCap := e.queryLimitCap(export)
	if limitCap > 0 {
		res.SetCap(limitCap)
	}

	// TODO: Get from OLAP instead of hardcoding
	cache := e.olap.Dialect() == drivers.DialectDuckDB

	return res, cache, nil
}
