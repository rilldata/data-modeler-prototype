import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type {
  V1MetricsViewAggregationRequest,
  V1MetricsViewComparisonRequest,
  V1MetricsViewRowsRequest,
  V1MetricsViewSpec,
  V1MetricsViewTimeSeriesRequest,
  V1MetricsViewToplistRequest,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export type QueryRequests =
  | V1MetricsViewAggregationRequest
  | V1MetricsViewToplistRequest
  | V1MetricsViewRowsRequest
  | V1MetricsViewTimeSeriesRequest
  | V1MetricsViewComparisonRequest;

export type QueryMapperArgs<R extends QueryRequests> = {
  queryClient: QueryClient;
  instanceId: string;
  dashboard: MetricsExplorerEntity;
  req: R;
  metricsView: V1MetricsViewSpec;
  timeRangeSummary: V1TimeRangeSummary;
  executionTime: string;
  annotations: Record<string, string>;
};
