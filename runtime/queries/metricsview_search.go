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
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewSearch struct {
	MetricsViewName    string                `json:"metrics_view_name,omitempty"`
	Dimensions         []string              `json:"dimensions,omitempty"`
	Search             string                `json:"search,omitempty"`
	TimeRange          *runtimev1.TimeRange  `json:"time_range,omitempty"`
	Where              *runtimev1.Expression `json:"where,omitempty"`
	Having             *runtimev1.Expression `json:"having,omitempty"`
	Priority           int32                 `json:"priority,omitempty"`
	Limit              *int64                `json:"limit,omitempty"`
	SecurityAttributes map[string]any        `json:"security_attributes,omitempty"`

	Result *runtimev1.MetricsViewSearchResponse
}

func (q *MetricsViewSearch) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewSearch:%s", string(r))
}

func (q *MetricsViewSearch) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewSearch) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewSearch) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewSearchResponse)
	if !ok {
		return fmt.Errorf("MetricsViewSearch: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewSearch) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	mv, lastUpdatedOn, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}
	resolvedSecurity, err := rt.ResolveMetricsViewSecurity(q.SecurityAttributes, instanceID, mv, lastUpdatedOn)
	if err != nil {
		return err
	}
	if resolvedSecurity != nil && !resolvedSecurity.Access {
		return ErrForbidden
	}
	for _, d := range q.Dimensions {
		if !checkFieldAccess(d, resolvedSecurity) {
			return ErrForbidden
		}
	}

	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid && olap.Dialect() != drivers.DialectClickHouse && olap.Dialect() != drivers.DialectPinot {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	if mv.TimeDimension == "" && !isTimeRangeNil(q.TimeRange) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", mv)
	}

	if !isTimeRangeNil(q.TimeRange) {
		start, end, err := ResolveTimeRange(q.TimeRange, mv)
		if err != nil {
			return err
		}
		q.TimeRange = &runtimev1.TimeRange{
			Start: timestamppb.New(start.In(time.UTC)),
			End:   timestamppb.New(end.In(time.UTC)),
		}
	}

	if olap.Dialect() == drivers.DialectDruid {
		return q.executeSearchInDruid(ctx, rt, instanceID, mv.Table)
	}

	sql, args, err := q.buildSearchQuerySQL(mv, olap.Dialect(), resolvedSecurity)
	if err != nil {
		return err
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil
	}
	defer rows.Close()

	q.Result = &runtimev1.MetricsViewSearchResponse{Results: make([]*runtimev1.MetricsViewSearchResponse_SearchResult, 0)}
	for rows.Next() {
		res := map[string]any{}
		err := rows.MapScan(res)
		if err != nil {
			return err
		}

		dimName, ok := res["dimension"].(string)
		if !ok {
			return fmt.Errorf("unknown result dimension: %q", dimName)
		}

		v, err := structpb.NewValue(res["value"])
		if err != nil {
			return err
		}

		q.Result.Results = append(q.Result.Results, &runtimev1.MetricsViewSearchResponse_SearchResult{
			Dimension: dimName,
			Value:     v,
		})
	}

	return nil
}

func (q *MetricsViewSearch) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return nil
}

func (q *MetricsViewSearch) executeSearchInDruid(ctx context.Context, rt *runtime.Runtime, instanceID, table string) error {
	// TODO: apply security filter
	inst, err := rt.Instance(ctx, instanceID)
	if err != nil {
		return err
	}

	dsn := ""
	for _, c := range inst.Connectors {
		if c.Name == "druid" {
			dsn = c.Config["dsn"]
		}
	}
	if dsn == "" {
		return fmt.Errorf("druid connector config not found in instance")
	}

	nq := druid.NewNativeQuery(strings.Replace(dsn, "/v2/sql/avatica-protobuf/", "/v2/", 1))
	req := druid.NewNativeSearchQueryRequest(table, q.Search, q.Dimensions, q.TimeRange.Start.AsTime(), q.TimeRange.End.AsTime())
	var res druid.NativeSearchQueryResponse
	err = nq.Do(ctx, req, &res, req.Context.QueryID)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewSearchResponse{Results: make([]*runtimev1.MetricsViewSearchResponse_SearchResult, 0)}
	for _, re := range res {
		for _, r := range re.Result {
			v, err := structpb.NewValue(r.Value)
			if err != nil {
				return err
			}
			q.Result.Results = append(q.Result.Results, &runtimev1.MetricsViewSearchResponse_SearchResult{
				Dimension: r.Dimension,
				Value:     v,
			})
		}
	}

	return nil
}

func (q *MetricsViewSearch) buildSearchQuerySQL(mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity) (string, []any, error) {
	var baseWhereClause string
	if policy != nil && policy.RowFilter != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
	}

	var args []any

	unions := make([]string, len(q.Dimensions))
	for i, dimName := range q.Dimensions {
		var dim *runtimev1.MetricsViewSpec_DimensionV2
		for _, d := range mv.Dimensions {
			if d.Name == dimName {
				dim = d
				break
			}
		}
		if dim == nil {
			return "", nil, fmt.Errorf("dimension not found: %q", q.Dimensions[i])
		}

		expr, _, unnest := dialect.DimensionSelectPair(mv.Database, mv.DatabaseSchema, mv.Table, dim)
		filterBuilder := &ExpressionBuilder{
			mv:      mv,
			dialect: dialect,
		}
		clause, clauseArgs, err := filterBuilder.buildExpression(expressionpb.Like(expressionpb.Identifier(dimName), expressionpb.String(fmt.Sprintf("%%%s%%", q.Search))))
		if err != nil {
			return "", nil, err
		}
		if clause != "" {
			clause = " AND " + clause
			args = append(args, clauseArgs...)
		}

		unions[i] = fmt.Sprintf(
			"SELECT %s as value, '%s' as dimension from %s %s WHERE 1=1 %s %s GROUP BY 1",
			expr,
			dimName,
			mv.Table,
			unnest,
			baseWhereClause,
			clause,
		)
	}

	return strings.Join(unions, " UNION "), args, nil
}
