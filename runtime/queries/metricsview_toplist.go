package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewToplist struct {
	MetricsViewName string                       `json:"metrics_view_name,omitempty"`
	DimensionName   string                       `json:"dimension_name,omitempty"`
	MeasureNames    []string                     `json:"measure_names,omitempty"`
	InlineMeasures  []*runtimev1.InlineMeasure   `json:"inline_measures,omitempty"`
	TimeStart       *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp       `json:"time_end,omitempty"`
	Limit           int64                        `json:"limit,omitempty"`
	Offset          int64                        `json:"offset,omitempty"`
	Sort            []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Filter          *runtimev1.MetricsViewFilter `json:"filter,omitempty"`

	Result *runtimev1.MetricsViewToplistResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewToplist{}

func (q *MetricsViewToplist) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewToplist:%s", string(r))
}

func (q *MetricsViewToplist) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewToplist) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewToplist) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewToplistResponse)
	if !ok {
		return fmt.Errorf("MetricsViewToplist: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewToplist) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	mv, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	if mv.TimeDimension == "" && (q.TimeStart != nil || q.TimeEnd != nil) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	// Build query
	sql, args, err := q.buildMetricsTopListSQL(mv, olap.Dialect())
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	// Execute
	meta, data, err := metricsQuery(ctx, olap, priority, sql, args)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewToplistResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewToplist) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int, format runtimev1.ExportFormat, writer io.Writer) error {
	err := q.Resolve(ctx, rt, instanceID, priority)
	if err != nil {
		return err
	}

	switch format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return writeCSV(q.Result.Meta, q.Result.Data, writer)
	}

	return nil
}

func (q *MetricsViewToplist) buildMetricsTopListSQL(mv *runtimev1.MetricsView, dialect drivers.Dialect) (string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return "", nil, err
	}

	dimName := safeName(q.DimensionName)
	selectCols := []string{dimName}
	for _, m := range ms {
		expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
		selectCols = append(selectCols, expr)
	}

	whereClause := "1=1"
	args := []any{}
	if mv.TimeDimension != "" {
		if q.TimeStart != nil {
			whereClause += fmt.Sprintf(" AND %s >= ?", safeName(mv.TimeDimension))
			args = append(args, q.TimeStart.AsTime())
		}
		if q.TimeEnd != nil {
			whereClause += fmt.Sprintf(" AND %s < ?", safeName(mv.TimeDimension))
			args = append(args, q.TimeEnd.AsTime())
		}
	}

	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(q.Filter, dialect)
		if err != nil {
			return "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
	}

	sortingCriteria := make([]string, 0, len(q.Sort))
	for _, s := range q.Sort {
		sortCriterion := safeName(s.Name)
		if !s.Ascending {
			sortCriterion += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			sortCriterion += " NULLS LAST"
		}
		sortingCriteria = append(sortingCriteria, sortCriterion)
	}
	orderClause := ""
	if len(sortingCriteria) > 0 {
		orderClause = "ORDER BY " + strings.Join(sortingCriteria, ", ")
	}

	if q.Limit == 0 {
		q.Limit = 100
	}

	sql := fmt.Sprintf("SELECT %s FROM %q WHERE %s GROUP BY %s %s LIMIT %d OFFSET %d",
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
		dimName,
		orderClause,
		q.Limit,
		q.Offset,
	)

	return sql, args, nil
}
