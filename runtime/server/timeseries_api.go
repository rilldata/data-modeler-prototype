package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// normaliseMeasures is called before this method so measure.SqlName will be non empty
func getExpressionColumnsFromMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += measure.Expression + " as " + measure.SqlName
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

// normaliseMeasures is called before this method so measure.SqlName will be non empty
func getCoalesceStatementsMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += fmt.Sprintf(`COALESCE(series.%s, 0) as %s`, measure.SqlName, measure.SqlName)
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

func EscapeSingleQuotes(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func getFilterFromDimensionValuesFilter(
	dimensionValues []*runtimev1.MetricsViewDimensionValue,
	prefix string,
	dimensionJoiner string,
) string {
	if len(dimensionValues) == 0 {
		return ""
	}
	var result string
	conditions := make([]string, 3)
	for i, dv := range dimensionValues {
		escapedName := EscapeSingleQuotes(dv.Name)
		var nulls bool
		var notNulls bool
		for _, iv := range dv.In {
			if iv.GetKind() == nil {
				nulls = true
			} else {
				notNulls = true
			}
		}
		conditions = conditions[:0]
		if notNulls {
			var inClause = escapedName + " " + prefix + " IN ("
			for j, iv := range dv.In {
				if iv.GetKind() != nil {
					inClause += "'" + EscapeSingleQuotes(iv.GetStringValue()) + "'"
				}
				if j < len(dv.In)-1 {
					inClause += ", "
				}
			}
			inClause += ")"
			conditions = append(conditions, inClause)
		}
		if nulls {
			var nullClause = escapedName + " IS " + prefix + " NULL"
			conditions = append(conditions, nullClause)
		}
		if len(dv.Like) > 0 {
			var likeClause string
			for j, lv := range dv.Like {
				if lv.GetKind() == nil {
					continue
				}
				likeClause += escapedName + " " + prefix + " ILIKE '" + EscapeSingleQuotes(lv.GetStringValue()) + "'"
				if j < len(dv.Like)-1 {
					likeClause += " AND "
				}
			}
			conditions = append(conditions, likeClause)
		}
		result += strings.Join(conditions, " "+dimensionJoiner+" ")
		if i < len(dimensionValues)-1 {
			result += " AND "
		}
	}

	return result
}

func getFilterFromMetricsViewFilters(filters *runtimev1.MetricsViewRequestFilter) string {
	includeFilters := getFilterFromDimensionValuesFilter(filters.Include, "", "OR")
	excludeFilters := getFilterFromDimensionValuesFilter(filters.Exclude, "NOT", "AND")
	if includeFilters != "" && excludeFilters != "" {
		return includeFilters + " AND " + excludeFilters
	} else if includeFilters != "" {
		return includeFilters
	} else if excludeFilters != "" {
		return excludeFilters
	} else {
		return ""
	}
}

func getFallbackMeasureName(index int, sqlName string) string {
	if sqlName == "" {
		s := fmt.Sprintf("measure_%d", index)
		return s
	} else {
		return sqlName
	}
}

var countName = "count"

func normaliseMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure {
	if len(measures) == 0 {
		return []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    countName,
				Id:         "",
			},
		}
	}
	for i, measure := range measures {
		measure.SqlName = getFallbackMeasureName(i, measure.SqlName)
	}
	return measures
}

// Metrics/Timeseries APIs
func (s *Server) EstimateRollupInterval(ctx context.Context, request *runtimev1.EstimateRollupIntervalRequest) (*runtimev1.EstimateRollupIntervalResponse, error) {
	tableName := EscapeDoubleQuotes(request.TableName)
	escapedColumnName := EscapeDoubleQuotes(request.ColumnName)
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: `SELECT
        	max(` + escapedColumnName + `) - min(` + escapedColumnName + `) as r,
        	max(` + escapedColumnName + `) as max_value,
        	min(` + escapedColumnName + `) as min_value,
        	count(*) as count
        	from ` + tableName,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rows.Next()
	var r duckdb.Interval
	var max, min time.Time
	var count int64
	err = rows.Scan(&r, &max, &min, &count)
	if err != nil {
		return nil, err
	}

	const (
		MICROS_SECOND = 1000 * 1000
		MICROS_MINUTE = 1000 * 1000 * 60
		MICROS_HOUR   = 1000 * 1000 * 60 * 60
		MICROS_DAY    = 1000 * 1000 * 60 * 60 * 24
	)

	var rollupInterval runtimev1.TimeGrain
	if r.Days == 0 && r.Micros <= MICROS_MINUTE {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND
	} else if r.Days == 0 && r.Micros > MICROS_MINUTE && r.Micros <= MICROS_HOUR {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_SECOND
	} else if r.Days == 0 && r.Micros <= MICROS_DAY {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_MINUTE
	} else if r.Days <= 7 {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_HOUR
	} else if r.Days <= 365*20 {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_DAY
	} else if r.Days <= 365*500 {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_MONTH
	} else {
		rollupInterval = runtimev1.TimeGrain_TIME_GRAIN_YEAR
	}

	return &runtimev1.EstimateRollupIntervalResponse{
		Interval: rollupInterval,
		Start:    timestamppb.New(min),
		End:      timestamppb.New(max),
	}, nil
}

func (s *Server) normaliseTimeRange(ctx context.Context, request *runtimev1.GenerateTimeSeriesRequest) (*runtimev1.TimeSeriesTimeRange, error) {
	rtr := request.TimeRange
	if rtr == nil {
		rtr = &runtimev1.TimeSeriesTimeRange{}
	}
	var result runtimev1.TimeSeriesTimeRange
	if rtr.Interval == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		r, err := s.EstimateRollupInterval(ctx, &runtimev1.EstimateRollupIntervalRequest{
			InstanceId: request.InstanceId,
			TableName:  request.TableName,
			ColumnName: request.TimestampColumnName,
		})
		if err != nil {
			return nil, err
		}
		result = runtimev1.TimeSeriesTimeRange{
			Interval: r.Interval,
			Start:    r.Start,
			End:      r.End,
		}
	} else if rtr.Start == nil || rtr.End == nil {
		tr, err := s.GetTimeRangeSummary(ctx, &runtimev1.GetTimeRangeSummaryRequest{
			InstanceId: request.InstanceId,
			TableName:  request.TableName,
			ColumnName: request.TimestampColumnName,
		})
		if err != nil {
			return nil, err
		}
		result = runtimev1.TimeSeriesTimeRange{
			Interval: rtr.Interval,
			Start:    tr.TimeRangeSummary.Min,
			End:      tr.TimeRangeSummary.Max,
		}
	}
	if rtr.Start != nil {
		result.Start = rtr.Start
	}
	if rtr.End != nil {
		result.End = rtr.End
	}
	if rtr.Interval != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		result.Interval = rtr.Interval
	}
	return &result, nil
}

const IsoFormat string = "2006-01-02T15:04:05.000Z"

func sMap(k string, v float64) map[string]float64 {
	m := make(map[string]float64, 1)
	m[k] = v
	return m
}

/**
 * Contains an as-of-this-commit unpublished algorithm for an M4-like line density reduction.
 * This will take in an n-length time series and produce a pixels * 4 reduction of the time series
 * that preserves the shape and trends.
 *
 * This algorithm expects the source table to have a timestamp column and some kind of value column,
 * meaning it expects the data to essentially already be aggregated.
 *
 * It's important to note that this implemention is NOT the original M4 aggregation method, but a method
 * that has the same basic understanding but is much faster.
 *
 * Nonetheless, we mostly use this to reduce a many-thousands-point-long time series to about 120 * 4 pixels.
 * Importantly, this function runs very fast. For more information about the original M4 method,
 * see http://www.vldb.org/pvldb/vol7/p797-jugel.pdf
 */
func (s *Server) createTimestampRollupReduction( // metadata: DatabaseMetadata,
	ctx context.Context,
	instanceId string,
	table string,
	timestampColumn string,
	valueColumn string,
	pixels int,
) ([]*runtimev1.TimeSeriesValue, error) {
	escapedTimestampColumn := EscapeDoubleQuotes(timestampColumn)
	cardinality, err := s.GetTableCardinality(ctx, &runtimev1.GetTableCardinalityRequest{
		InstanceId: instanceId,
		TableName:  table,
	})
	if err != nil {
		return nil, err
	}

	if cardinality.Cardinality < int64(pixels*4) {
		rows, err := s.query(ctx, instanceId, &drivers.Statement{
			Query: `SELECT ` + escapedTimestampColumn + ` as ts, "` + valueColumn + `" as count FROM "` + table + `"`,
		})
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		results := make([]*runtimev1.TimeSeriesValue, 0, (pixels+1)*4)
		for rows.Next() {
			var ts time.Time
			var count float64
			err = rows.Scan(&ts, &count)
			if err != nil {
				return nil, err
			}
			results = append(results, &runtimev1.TimeSeriesValue{
				Ts:      ts.Format(IsoFormat),
				Records: sMap("count", count),
			})
		}
		return results, nil
	}

	sql := ` -- extract unix time
      WITH Q as (
        SELECT extract('epoch' from ` + escapedTimestampColumn + `) as t, "` + valueColumn + `" as v FROM "` + table + `"
      ),
      -- generate bounds
      M as (
        SELECT min(t) as t1, max(t) as t2, max(t) - min(t) as diff FROM Q
      )
      -- core logic
      SELECT 
        -- left boundary point
        min(t) * 1000  as min_t, 
        arg_min(v, t) as argmin_tv, 

        -- right boundary point
        max(t) * 1000 as max_t, 
        arg_max(v, t) as argmax_tv,

        -- smallest point within boundary
        min(v) as min_v, 
        arg_min(t, v) * 1000  as argmin_vt,

        -- largest point within boundary
        max(v) as max_v, 
        arg_max(t, v) * 1000  as argmax_vt,

        round(` + strconv.FormatInt(int64(pixels), 10) + ` * (t - (SELECT t1 FROM M)) / (SELECT diff FROM M)) AS bin
  
      FROM Q GROUP BY bin
      ORDER BY bin
    `

	rows, err := s.query(ctx, instanceId, &drivers.Statement{
		Query: sql,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]*runtimev1.TimeSeriesValue, 0, (pixels+1)*4)
	for rows.Next() {
		var minT, maxT, argminVT, argmaxVT int64
		var argminTV, argmaxTV, minV, maxV float64
		var bin float64
		err = rows.Scan(&minT, &argminTV, &maxT, &argmaxTV, &minV, &argminVT, &maxV, &argmaxVT, &bin)
		if err != nil {
			return nil, err
		}
		results = append(results, &runtimev1.TimeSeriesValue{
			Ts:      time.UnixMilli(minT).Format(IsoFormat),
			Bin:     &bin,
			Records: sMap("count", argminTV),
		})
		results = append(results, &runtimev1.TimeSeriesValue{
			Ts:      time.UnixMilli(argminVT).Format(IsoFormat),
			Bin:     &bin,
			Records: sMap("count", minV),
		})

		results = append(results, &runtimev1.TimeSeriesValue{
			Ts:      time.UnixMilli(argmaxVT).Format(IsoFormat),
			Bin:     &bin,
			Records: sMap("count", maxV),
		})

		results = append(results, &runtimev1.TimeSeriesValue{
			Ts:      time.UnixMilli(maxT).Format(IsoFormat),
			Bin:     &bin,
			Records: sMap("count", argmaxTV),
		})
		if argminVT > argmaxVT {
			i := len(results)
			results[i-3], results[i-2] = results[i-2], results[i-3]
		}
	}
	return results, nil
}

func convertToDateTruncSpecifier(specifier runtimev1.TimeGrain) string {
	switch specifier {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "MILLISECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "SECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "MINUTE"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "HOUR"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "DAY"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "WEEK"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "MONTH"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "YEAR"
	}
	panic(fmt.Errorf("unconvertable time grain specifier: %v", specifier))
}

func (s *Server) GenerateTimeSeries(ctx context.Context, request *runtimev1.GenerateTimeSeriesRequest) (*runtimev1.GenerateTimeSeriesResponse, error) {
	timeRange, err := s.normaliseTimeRange(ctx, request)
	if err != nil {
		return nil, err
	}
	var measures = normaliseMeasures(request.Measures)
	var timestampColumn = request.TimestampColumnName
	var tableName = request.TableName
	var filter string
	if request.Filters != nil {
		filter = getFilterFromMetricsViewFilters(request.Filters)
	}
	var timeGranularity = convertToDateTruncSpecifier(timeRange.Interval)
	var tsAlias string
	if timestampColumn == "ts" {
		tsAlias = "_ts"
	} else {
		tsAlias = "ts"
	}
	if filter != "" {
		filter = "WHERE " + filter
	}
	sql := `CREATE TEMPORARY TABLE _ts_ AS (
        -- generate a time series column that has the intended range
        WITH template as (
          SELECT 
            generate_series as ` + tsAlias + `
          FROM 
            generate_series(
              date_trunc('` + timeGranularity + `', TIMESTAMP '` + timeRange.Start.AsTime().Format(IsoFormat) + `'),
              date_trunc('` + timeGranularity + `', TIMESTAMP '` + timeRange.End.AsTime().Format(IsoFormat) + `'),
              interval '1 ` + timeGranularity + `')
        ),
        -- transform the original data, and optionally sample it.
        series AS (
          SELECT 
            date_trunc('` + timeGranularity + `', "` + EscapeDoubleQuotes(timestampColumn) + `") as ` + tsAlias + `,` + getExpressionColumnsFromMeasures(measures) + `
          FROM "` + EscapeDoubleQuotes(tableName) + `" ` + filter + `
          GROUP BY ` + tsAlias + ` ORDER BY ` + tsAlias + `
        )
        -- join the transformed data with the generated time series column,
        -- coalescing the first value to get the 0-default when the rolled up data
        -- does not have that value.
        SELECT 
          ` + getCoalesceStatementsMeasures(measures) + `,
          template.` + tsAlias + ` as ts from template
        LEFT OUTER JOIN series ON template.` + tsAlias + ` = series.` + tsAlias + `
        ORDER BY template.` + tsAlias + `
      )`
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: sql,
	})
	defer s.dropTempTable(ctx, request.InstanceId)
	if err != nil {
		return nil, err
	}
	rows.Close()
	rows, err = s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: "SELECT * from _ts_",
	})
	if err != nil {
		return nil, err
	}
	results, err := convertRowsToTimeSeriesValues(rows, len(measures)+1)
	if err != nil {
		return nil, err
	}
	var sparkValues []*runtimev1.TimeSeriesValue
	if request.Pixels != 0 {
		pixels := int(request.Pixels)
		sparkValues, err = s.createTimestampRollupReduction(ctx, request.InstanceId, "_ts_", "ts", "count", pixels)
		if err != nil {
			return nil, err
		}
	}

	return &runtimev1.GenerateTimeSeriesResponse{
		Rollup: &runtimev1.TimeSeriesResponse{
			Results:   results,
			TimeRange: timeRange,
			Spark:     sparkValues,
		},
	}, nil
}

func convertRowsToTimeSeriesValues(rows *drivers.Result, rowLength int) ([]*runtimev1.TimeSeriesValue, error) {
	results := make([]*runtimev1.TimeSeriesValue, 0)
	defer rows.Close()
	var converr error
	for rows.Next() {
		value := runtimev1.TimeSeriesValue{}
		results = append(results, &value)
		row := make(map[string]interface{}, rowLength)
		err := rows.MapScan(row)
		if err != nil {
			return results, err
		}
		value.Ts = row["ts"].(time.Time).Format(IsoFormat)
		delete(row, "ts")
		value.Records = make(map[string]float64, len(row))
		for k, v := range row {
			switch x := v.(type) {
			case int32:
				value.Records[k] = float64(x)
			case int64:
				value.Records[k] = float64(x)
			case float32:
				value.Records[k] = float64(x)
			case float64:
				value.Records[k] = x
			default:
				return nil, fmt.Errorf("unknown type %T ", v)
			}
		}
	}
	return results, converr
}

func (s *Server) dropTempTable(ctx context.Context, instanceId string) {
	rs, er := s.query(ctx, instanceId, &drivers.Statement{
		Query: "DROP TABLE _ts_",
	})
	if er == nil {
		rs.Close()
	}
}
