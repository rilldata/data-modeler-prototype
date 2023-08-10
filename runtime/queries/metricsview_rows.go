package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewRows struct {
	MetricsViewName string                       `json:"metrics_view_name,omitempty"`
	TimeStart       *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp       `json:"time_end,omitempty"`
	TimeGranularity runtimev1.TimeGrain          `json:"time_granularity,omitempty"`
	Filter          *runtimev1.MetricsViewFilter `json:"filter,omitempty"`
	Sort            []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Limit           *int64                       `json:"limit,omitempty"`
	Offset          int64                        `json:"offset,omitempty"`
	TimeZone        string                       `json:"time_zone,omitempty"`

	Result *runtimev1.MetricsViewRowsResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewRows{}

func (q *MetricsViewRows) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewRows:%s", string(r))
}

func (q *MetricsViewRows) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewRows) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewRows) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewRowsResponse)
	if !ok {
		return fmt.Errorf("MetricsViewRows: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewRows) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
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

	timeRollupColumnName, err := q.resolveTimeRollupColumnName(ctx, rt, instanceID, priority, mv)
	if err != nil {
		return err
	}

	ql, args, err := q.buildMetricsRowsSQL(mv, olap.Dialect(), timeRollupColumnName)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	meta, data, err := metricsQuery(ctx, olap, priority, ql, args)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewRowsResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewRows) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		return err
	}

	mv, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	filename := strings.ReplaceAll(mv.Model, `"`, `_`)
	if q.TimeStart != nil || q.TimeEnd != nil || q.Filter != nil && (len(q.Filter.Include) > 0 || len(q.Filter.Exclude) > 0) {
		filename += "_filtered"
	}

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(filename)
		if err != nil {
			return err
		}
	}

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return writeCSV(q.Result.Meta, q.Result.Data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return writeXLSX(q.Result.Meta, q.Result.Data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return writeParquet(q.Result.Meta, q.Result.Data, w)
	}

	return nil
}

// resolveTimeRollupColumnName infers a column name for time rollup values.
// The rollup column name takes the format "{time dimension name}__{granularity}[optional number]".
// The optional number is appended in case of collision with an existing column name.
// It returns an empty string for cases where no time rollup should be calculated (such as when q.TimeGranularity is not set).
func (q *MetricsViewRows) resolveTimeRollupColumnName(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int, mv *runtimev1.MetricsView) (string, error) {
	// Skip if no time info is available
	if mv.TimeDimension == "" || q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		return "", nil
	}

	entry, err := rt.GetCatalogEntry(ctx, instanceID, mv.Model)
	if err != nil {
		return "", err
	}
	model := entry.GetModel()
	if model == nil {
		return "", fmt.Errorf("model %q not found for metrics view %q", mv.Model, mv.Name)
	}

	// Create name stem
	stem := fmt.Sprintf("%s__%s", mv.TimeDimension, strings.ToLower(convertToDateTruncSpecifier(q.TimeGranularity)))

	// Try new candidate names until we find an available one (capping the search at 10 names)
	for i := 0; i < 10; i++ {
		candidate := stem
		if i != 0 {
			candidate += strconv.Itoa(i)
		}

		// Do a case-insensitive search for the candidate name
		found := false
		for _, col := range model.Schema.Fields {
			if strings.EqualFold(candidate, col.Name) {
				found = true
				break
			}
		}
		if !found {
			return candidate, nil
		}
	}

	// Very unlikely case where no available candidate name was found.
	// By returning the empty string, the downstream logic will skip computing the rollup.
	return "", nil
}

func (q *MetricsViewRows) buildMetricsRowsSQL(mv *runtimev1.MetricsView, dialect drivers.Dialect, timeRollupColumnName string) (string, []any, error) {
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
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(mv, q.Filter, dialect)
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

	var limitClause string
	if q.Limit != nil {
		if *q.Limit == 0 {
			*q.Limit = 100
		}
		limitClause = fmt.Sprintf("LIMIT %d", *q.Limit)
	}

	selectColumns := []string{"*"}

	if timeRollupColumnName != "" {
		if mv.TimeDimension == "" || q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			panic("timeRollupColumnName is set, but time dimension info is missing")
		}

		timezone := "UTC"
		if q.TimeZone != "" {
			timezone = q.TimeZone
		}
		args = append([]any{timezone, timezone}, args...)
		rollup := fmt.Sprintf("timezone(?, date_trunc('%s', timezone(?, %s::TIMESTAMPTZ))) AS %s", convertToDateTruncSpecifier(q.TimeGranularity), safeName(mv.TimeDimension), safeName(timeRollupColumnName))

		// Prepend the rollup column
		selectColumns = append([]string{rollup}, selectColumns...)
	}

	sql := fmt.Sprintf("SELECT %s FROM %q WHERE %s %s %s OFFSET %d",
		strings.Join(selectColumns, ","),
		mv.Model,
		whereClause,
		orderClause,
		limitClause,
		q.Offset,
	)

	return sql, args, nil
}
