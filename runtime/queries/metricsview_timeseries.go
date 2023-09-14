package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewTimeSeries struct {
	MetricsViewName    string                               `json:"metrics_view_name,omitempty"`
	MeasureNames       []string                             `json:"measure_names,omitempty"`
	InlineMeasures     []*runtimev1.InlineMeasure           `json:"inline_measures,omitempty"`
	TimeStart          *timestamppb.Timestamp               `json:"time_start,omitempty"`
	TimeEnd            *timestamppb.Timestamp               `json:"time_end,omitempty"`
	Limit              int64                                `json:"limit,omitempty"`
	Offset             int64                                `json:"offset,omitempty"`
	Sort               []*runtimev1.MetricsViewSort         `json:"sort,omitempty"`
	Filter             *runtimev1.MetricsViewFilter         `json:"filter,omitempty"`
	TimeGranularity    runtimev1.TimeGrain                  `json:"time_granularity,omitempty"`
	TimeZone           string                               `json:"time_zone,omitempty"`
	MetricsView        *runtimev1.MetricsView               `json:"-"`
	ResolvedMVSecurity *runtime.ResolvedMetricsViewSecurity `json:"security"`

	Result *runtimev1.MetricsViewTimeSeriesResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewTimeSeries{}

func (q *MetricsViewTimeSeries) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewTimeSeries:%s", r)
}

func (q *MetricsViewTimeSeries) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewTimeSeries) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewTimeSeries) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewTimeSeriesResponse)
	if !ok {
		return fmt.Errorf("MetricsViewTimeSeries: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewTimeSeries) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	if q.MetricsView.TimeDimension == "" {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	obj, err := rt.GetCatalogEntry(ctx, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	mv := obj.GetMetricsView()

	sql, tsAlias, args, err := q.buildMetricsTimeseriesSQL(olap, mv, q.ResolvedMVSecurity)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	// Omit the time value from the result schema
	schema := rows.Schema
	if schema != nil {
		for i, f := range schema.Fields {
			if f.Name == tsAlias {
				schema.Fields = slices.Delete(schema.Fields, i, i+1)
				break
			}
		}
	}

	var start time.Time
	var zeroTime time.Time
	var data []*runtimev1.TimeSeriesValue
	var nullRecords *structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return err
		}

		var t time.Time
		switch v := rowMap[tsAlias].(type) {
		case time.Time:
			t = v
		default:
			panic(fmt.Sprintf("unexpected type for timestamp column: %T", v))
		}
		delete(rowMap, tsAlias)

		records, err := pbutil.ToStruct(rowMap, schema)
		if err != nil {
			return err
		}

		tz := time.UTC
		if q.TimeZone != "" {
			tz, err = time.LoadLocation(q.TimeZone)
			if err != nil {
				return err
			}
		}
		if nullRecords == nil {
			nullRecords = generateNullRecords(records)
		}
		if start.Before(t) {
			if zeroTime.Equal(start) {
				if q.TimeStart != nil {
					start = truncateTime(q.TimeStart.AsTime(), q.TimeGranularity, tz, 1, 1)
					data = addNulls(data, nullRecords, start, t, q.TimeGranularity)
				}
			} else {
				data = addNulls(data, nullRecords, start, t, q.TimeGranularity)
			}
		}

		data = append(data, &runtimev1.TimeSeriesValue{
			Ts:      timestamppb.New(t),
			Records: records,
		})
		start = addTo(t, q.TimeGranularity)
	}
	if q.TimeEnd != nil && nullRecords != nil {
		data = addNulls(data, nullRecords, start, q.TimeEnd.AsTime(), q.TimeGranularity)
	}

	meta := structTypeToMetricsViewColumn(rows.Schema)

	q.Result = &runtimev1.MetricsViewTimeSeriesResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func truncateTime(start time.Time, tg runtimev1.TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	switch tg {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return start.Truncate(time.Millisecond)
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return start.Truncate(time.Second)
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return start.Truncate(time.Minute)
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), 0, 0, 0, tz)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, tz)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		start = start.In(tz)
		weekday := int(start.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		if firstDay < 1 {
			firstDay = 1
		}
		if firstDay > 7 {
			firstDay = 7
		}

		daysToSubtract := -(weekday - firstDay)
		if weekday < firstDay {
			daysToSubtract = -7 + daysToSubtract
		}
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, tz)
		start = start.AddDate(0, 0, daysToSubtract)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, tz)
		start = start.In(time.UTC)
		return start
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		monthsToSubtract := 1 - int(start.Month())%3 // todo first month of year
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, tz)
		start = start.AddDate(0, monthsToSubtract, 0)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		start = start.In(tz)
		year := start.Year()
		if int(start.Month()) < firstMonth {
			year = start.Year() - 1
		}

		start = time.Date(year, time.Month(firstMonth), 1, 0, 0, 0, 0, tz)
		return start.In(time.UTC)
	}

	return start
}

func generateNullRecords(s *structpb.Struct) *structpb.Struct {
	nullStruct := structpb.Struct{Fields: make(map[string]*structpb.Value, len(s.Fields))}
	for k := range s.Fields {
		nullStruct.Fields[k] = structpb.NewNullValue()
	}
	return &nullStruct
}

func addNulls(data []*runtimev1.TimeSeriesValue, nullRecords *structpb.Struct, start, end time.Time, tg runtimev1.TimeGrain) []*runtimev1.TimeSeriesValue {
	for start.Before(end) {
		data = append(data, &runtimev1.TimeSeriesValue{
			Ts:      timestamppb.New(start),
			Records: nullRecords,
		})
		start = addTo(start, tg)
	}
	return data
}

func addTo(start time.Time, tg runtimev1.TimeGrain) time.Time {
	switch tg {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return start.Add(time.Millisecond)
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return start.Add(time.Second)
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return start.Add(time.Minute)
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return start.Add(time.Hour)
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return start.AddDate(0, 0, 1)
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return start.AddDate(0, 0, 7)
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return start.AddDate(0, 1, 0)
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return start.AddDate(0, 3, 0)
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return start.AddDate(1, 0, 0)
	}

	return start
}

func (q *MetricsViewTimeSeries) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		return err
	}

	obj, err := rt.GetCatalogEntry(ctx, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	mv := obj.GetMetricsView()

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(q.generateFilename(mv))
		if err != nil {
			return err
		}
	}

	tmp := make([]*structpb.Struct, 0, len(q.Result.Data))
	meta := append([]*runtimev1.MetricsViewColumn{{
		Name: mv.TimeDimension,
	}}, q.Result.Meta...)
	for _, dt := range q.Result.Data {
		dt.Records.Fields[mv.TimeDimension] = structpb.NewStringValue(dt.Ts.AsTime().Format(time.RFC3339Nano))
		tmp = append(tmp, dt.Records)
	}

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return writeCSV(meta, tmp, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return writeXLSX(meta, tmp, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return writeParquet(meta, tmp, w)
	}

	return nil
}

func (q *MetricsViewTimeSeries) generateFilename(mv *runtimev1.MetricsView) string {
	filename := strings.ReplaceAll(mv.Model, `"`, `_`)
	if q.TimeStart != nil || q.TimeEnd != nil || q.Filter != nil && (len(q.Filter.Include) > 0 || len(q.Filter.Exclude) > 0) {
		filename += "_filtered"
	}
	return filename
}

func (q *MetricsViewTimeSeries) buildMetricsTimeseriesSQL(olap drivers.OLAPStore, mv *runtimev1.MetricsView, policy *runtime.ResolvedMetricsViewSecurity) (string, string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return "", "", nil, err
	}

	selectCols := []string{}
	for _, m := range ms {
		expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
		selectCols = append(selectCols, expr)
	}

	whereClause := "1=1"
	args := []any{}
	if q.TimeStart != nil {
		whereClause += fmt.Sprintf(" AND %s >= ?", safeName(mv.TimeDimension))
		args = append(args, q.TimeStart.AsTime())
	}
	if q.TimeEnd != nil {
		whereClause += fmt.Sprintf(" AND %s < ?", safeName(mv.TimeDimension))
		args = append(args, q.TimeEnd.AsTime())
	}

	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(mv, q.Filter, drivers.DialectDruid, policy)
		if err != nil {
			return "", "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
	}

	tsAlias := tempName("_ts_")
	timezone := "UTC"
	if q.TimeZone != "" {
		timezone = q.TimeZone
	}
	args = append([]any{timezone, timezone}, args...)

	var sql string
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		sql = q.buildDuckDBSQL(args, mv, tsAlias, selectCols, whereClause)
	case drivers.DialectDruid:
		sql = q.buildDruidSQL(args, mv, tsAlias, selectCols, whereClause)
	default:
		return "", "", nil, fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	return sql, tsAlias, args, nil
}

func (q *MetricsViewTimeSeries) buildDruidSQL(args []any, mv *runtimev1.MetricsView, tsAlias string, selectCols []string, whereClause string) string {
	tsSpecifier := convertToDruidTimeFloorSpecifier(q.TimeGranularity)

	sql := fmt.Sprintf(
		`SELECT time_floor(%s, '%s', null, CAST(? AS VARCHAR)) AS %s, %s FROM %q WHERE %s GROUP BY 1 ORDER BY 1`,
		safeName(mv.TimeDimension),
		tsSpecifier,
		tsAlias,
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
	)

	return sql
}

func (q *MetricsViewTimeSeries) buildDuckDBSQL(args []any, mv *runtimev1.MetricsView, tsAlias string, selectCols []string, whereClause string) string {
	dateTruncSpecifier := convertToDateTruncSpecifier(q.TimeGranularity)
	sql := fmt.Sprintf(
		`SELECT timezone(?, date_trunc('%s', timezone(?, %s::TIMESTAMPTZ))) as %s, %s FROM %q WHERE %s GROUP BY 1 ORDER BY 1`,
		dateTruncSpecifier,
		safeName(mv.TimeDimension),
		tsAlias,
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
	)

	return sql
}
