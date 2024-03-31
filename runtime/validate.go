package runtime

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

type ValidateMetricsViewResult struct {
	TimeDimensionErr error
	DimensionErrs    []IndexErr
	MeasureErrs      []IndexErr
	OtherErrs        []error
}

type IndexErr struct {
	Idx int
	Err error
}

func (r *ValidateMetricsViewResult) IsZero() bool {
	return r.TimeDimensionErr == nil && len(r.DimensionErrs) == 0 && len(r.MeasureErrs) == 0 && len(r.OtherErrs) == 0
}

// Error returns a single error containing all validation errors.
// If there are no errors, it returns nil.
func (r *ValidateMetricsViewResult) Error() error {
	var errs []error
	errs = append(errs, r.TimeDimensionErr)
	for _, e := range r.DimensionErrs {
		errs = append(errs, e.Err)
	}
	for _, e := range r.MeasureErrs {
		errs = append(errs, e.Err)
	}
	errs = append(errs, r.OtherErrs...)

	// NOTE: errors.Join returns nil if all input errs are nil.
	return errors.Join(errs...)
}

// ValidateMetricsView validates a metrics view spec.
// NOTE: If we need validation for more resources, we should consider moving it to the queries (or a dedicated validation package).
func (r *Runtime) ValidateMetricsView(ctx context.Context, instanceID string, mv *runtimev1.MetricsViewSpec) (*ValidateMetricsViewResult, error) {
	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	olap, release, err := ctrl.AcquireOLAP(ctx, mv.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	// Create the result
	res := &ValidateMetricsViewResult{}

	// Check underlying table exists
	t, err := olap.InformationSchema().Lookup(ctx, mv.Table)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			res.OtherErrs = append(res.OtherErrs, fmt.Errorf("table %q does not exist", mv.Table))
			return res, nil
		}
		return nil, fmt.Errorf("could not find table %q: %w", mv.Table, err)
	}

	fields := make(map[string]*runtimev1.StructType_Field, len(t.Schema.Fields))
	for _, f := range t.Schema.Fields {
		fields[strings.ToLower(f.Name)] = f
	}

	// Check time dimension exists
	if mv.TimeDimension != "" {
		f, ok := fields[strings.ToLower(mv.TimeDimension)]
		if !ok {
			res.TimeDimensionErr = fmt.Errorf("timeseries %q is not a column in table %q", mv.TimeDimension, mv.Table)
		} else if f.Type.Code != runtimev1.Type_CODE_TIMESTAMP && f.Type.Code != runtimev1.Type_CODE_DATE {
			res.TimeDimensionErr = fmt.Errorf("timeseries %q is not a TIMESTAMP column", mv.TimeDimension)
		}
	}

	// Check dimension columns exist
	for idx, d := range mv.Dimensions {
		err = validateDimensionExist(d, fields)
		if err != nil {
			res.DimensionErrs = append(res.DimensionErrs, IndexErr{
				Idx: idx,
				Err: err,
			})
		}
	}

	err = validateAllDimensionAndMeasureExpr(ctx, olap, t, mv)
	if err != nil {
		// Some measure or dimension failed validation, check individuals to provide useful errors
		// Check dimension expressions are valid
		for idx, d := range mv.Dimensions {
			err = validateDimensionExpr(ctx, olap, t, d)
			if err != nil {
				res.DimensionErrs = append(res.DimensionErrs, IndexErr{
					Idx: idx,
					Err: err,
				})
			}
		}
		// Check measure expressions are valid
		for idx, m := range mv.Measures {
			err := validateMeasure(ctx, olap, t, m)
			if err != nil {
				res.MeasureErrs = append(res.MeasureErrs, IndexErr{
					Idx: idx,
					Err: fmt.Errorf("invalid expression for measure %q: %w", m.Name, err),
				})
			}
		}
	}

	if mv.DefaultTheme != "" {
		_, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindTheme, Name: mv.DefaultTheme}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrNotFound) {
				res.OtherErrs = append(res.OtherErrs, fmt.Errorf("theme %q does not exist", mv.DefaultTheme))
			}
			return nil, fmt.Errorf("could not find theme %q: %w", mv.DefaultTheme, err)
		}
	}

	return res, nil
}

func validateDimensionExist(d *runtimev1.MetricsViewSpec_DimensionV2, fields map[string]*runtimev1.StructType_Field) error {
	if d.Column != "" {
		if _, isColumn := fields[strings.ToLower(d.Column)]; !isColumn {
			return fmt.Errorf("failed to validate dimension %q: column %q not found in table", d.Name, d.Column)
		}
	}
	return nil
}

func validateAllDimensionAndMeasureExpr(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, mv *runtimev1.MetricsViewSpec) error {
	var dimExprs []string
	var groupIndexes []string
	for idx, d := range mv.Dimensions {
		dimExprs = append(dimExprs, extractDimExpr(d))
		groupIndexes = append(groupIndexes, strconv.Itoa(idx+1))
	}
	var metricExprs []string
	for _, m := range mv.Measures {
		metricExprs = append(metricExprs, m.Expression) // note the = instead of :=
	}
	var query string
	if len(dimExprs) == 0 && len(metricExprs) == 0 {
		// No metric and dimension, nothing to check
		return nil
	}
	if len(dimExprs) == 0 {
		// Only metrics
		query = fmt.Sprintf("SELECT 1, %s FROM %s GROUP BY 1", strings.Join(metricExprs, ","), safeSQLName(t.Name))
	} else if len(metricExprs) == 0 {
		// No metrics
		query = fmt.Sprintf("SELECT (%s) FROM %s GROUP BY %s", strings.Join(dimExprs, "),("), safeSQLName(t.Name), strings.Join(groupIndexes, ","))
	} else {
		query = fmt.Sprintf("SELECT (%s), %s FROM %s GROUP BY %s", strings.Join(dimExprs, "),("), strings.Join(metricExprs, ","), safeSQLName(t.Name), strings.Join(groupIndexes, ","))
	}
	err := olap.Exec(ctx, &drivers.Statement{
		Query:  query,
		DryRun: true,
	})
	if err != nil {
		return fmt.Errorf("failed to validate dims and metrics: %w", err)
	}
	return nil
}

func validateDimensionExpr(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, d *runtimev1.MetricsViewSpec_DimensionV2) error {
	err := olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT (%s) FROM %s GROUP BY 1", extractDimExpr(d), safeSQLName(t.Name)),
		DryRun: true,
	})
	if err != nil {
		return fmt.Errorf("failed to validate expression for dimension %q: %w", d.Name, err)
	}
	return nil
}

func validateMeasure(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	err := olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT 1, %s FROM %s GROUP BY 1", m.Expression, safeSQLName(t.Name)),
		DryRun: true,
	})
	return err
}

func safeSQLName(name string) string {
	if name == "" {
		return name
	}
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(name, "\"", "\"\""))
}

func extractDimExpr(d *runtimev1.MetricsViewSpec_DimensionV2) string {
	if d.Column != "" {
		return "\"" + d.Column + "\""
	}
	return d.Expression
}
