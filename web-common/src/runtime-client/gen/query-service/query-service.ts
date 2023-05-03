/**
 * Generated by orval v6.13.1 🍺
 * Do not edit manually.
 * rill/runtime/v1/schema.proto
 * OpenAPI spec version: version not set
 */
import { createQuery, createMutation } from "@tanstack/svelte-query";
import type {
  CreateQueryOptions,
  CreateMutationOptions,
  QueryFunction,
  MutationFunction,
  CreateQueryResult,
  QueryKey,
} from "@tanstack/svelte-query";
import type {
  V1ColumnCardinalityResponse,
  RpcStatus,
  QueryServiceColumnCardinalityParams,
  V1TableColumnsResponse,
  QueryServiceTableColumnsParams,
  V1ColumnDescriptiveStatisticsResponse,
  QueryServiceColumnDescriptiveStatisticsParams,
  V1MetricsViewComparisonToplistResponse,
  QueryServiceMetricsViewComparisonToplistBody,
  V1MetricsViewTimeSeriesResponse,
  QueryServiceMetricsViewTimeSeriesBody,
  V1MetricsViewToplistResponse,
  QueryServiceMetricsViewToplistBody,
  V1MetricsViewTotalsResponse,
  QueryServiceMetricsViewTotalsBody,
  V1ColumnNullCountResponse,
  QueryServiceColumnNullCountParams,
  V1ColumnNumericHistogramResponse,
  QueryServiceColumnNumericHistogramParams,
  V1ColumnRollupIntervalResponse,
  QueryServiceColumnRollupIntervalBody,
  V1TableRowsResponse,
  QueryServiceTableRowsParams,
  V1ColumnRugHistogramResponse,
  QueryServiceColumnRugHistogramParams,
  V1ColumnTimeGrainResponse,
  QueryServiceColumnTimeGrainParams,
  V1TableCardinalityResponse,
  QueryServiceTableCardinalityParams,
  V1ColumnTimeRangeResponse,
  QueryServiceColumnTimeRangeParams,
  V1ColumnTimeSeriesResponse,
  QueryServiceColumnTimeSeriesBody,
  V1ColumnTopKResponse,
  QueryServiceColumnTopKBody,
  V1QueryResponse,
  QueryServiceQueryBody,
} from "../index.schemas";
import { httpClient } from "../../http-client";

/**
 * @summary Get cardinality for a column
 */
export const queryServiceColumnCardinality = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnCardinalityParams,
  signal?: AbortSignal
) => {
  return httpClient<V1ColumnCardinalityResponse>({
    url: `/v1/instances/${instanceId}/queries/column-cardinality/tables/${tableName}`,
    method: "get",
    params,
    signal,
  });
};

export const getQueryServiceColumnCardinalityQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnCardinalityParams
) =>
  [
    `/v1/instances/${instanceId}/queries/column-cardinality/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceColumnCardinalityQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnCardinality>>
>;
export type QueryServiceColumnCardinalityQueryError = RpcStatus;

export const createQueryServiceColumnCardinality = <
  TData = Awaited<ReturnType<typeof queryServiceColumnCardinality>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnCardinalityParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnCardinality>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnCardinalityQueryKey(instanceId, tableName, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnCardinality>>
  > = ({ signal }) =>
    queryServiceColumnCardinality(instanceId, tableName, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnCardinality>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary TableColumns returns column profiles
 */
export const queryServiceTableColumns = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceTableColumnsParams
) => {
  return httpClient<V1TableColumnsResponse>({
    url: `/v1/instances/${instanceId}/queries/columns-profile/tables/${tableName}`,
    method: "post",
    params,
  });
};

export const getQueryServiceTableColumnsQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceTableColumnsParams
) =>
  [
    `/v1/instances/${instanceId}/queries/columns-profile/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceTableColumnsQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceTableColumns>>
>;
export type QueryServiceTableColumnsQueryError = RpcStatus;

export const createQueryServiceTableColumns = <
  TData = Awaited<ReturnType<typeof queryServiceTableColumns>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceTableColumnsParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceTableColumns>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceTableColumnsQueryKey(instanceId, tableName, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceTableColumns>>
  > = () => queryServiceTableColumns(instanceId, tableName, params);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceTableColumns>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary Get basic stats for a numeric column like min, max, mean, stddev, etc
 */
export const queryServiceColumnDescriptiveStatistics = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnDescriptiveStatisticsParams,
  signal?: AbortSignal
) => {
  return httpClient<V1ColumnDescriptiveStatisticsResponse>({
    url: `/v1/instances/${instanceId}/queries/descriptive-statistics/tables/${tableName}`,
    method: "get",
    params,
    signal,
  });
};

export const getQueryServiceColumnDescriptiveStatisticsQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnDescriptiveStatisticsParams
) =>
  [
    `/v1/instances/${instanceId}/queries/descriptive-statistics/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceColumnDescriptiveStatisticsQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnDescriptiveStatistics>>
>;
export type QueryServiceColumnDescriptiveStatisticsQueryError = RpcStatus;

export const createQueryServiceColumnDescriptiveStatistics = <
  TData = Awaited<ReturnType<typeof queryServiceColumnDescriptiveStatistics>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnDescriptiveStatisticsParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnDescriptiveStatistics>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnDescriptiveStatisticsQueryKey(
      instanceId,
      tableName,
      params
    );

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnDescriptiveStatistics>>
  > = ({ signal }) =>
    queryServiceColumnDescriptiveStatistics(
      instanceId,
      tableName,
      params,
      signal
    );

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnDescriptiveStatistics>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

export const queryServiceMetricsViewComparisonToplist = (
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewComparisonToplistBody: QueryServiceMetricsViewComparisonToplistBody
) => {
  return httpClient<V1MetricsViewComparisonToplistResponse>({
    url: `/v1/instances/${instanceId}/queries/metrics-views/${metricsViewName}/compare-toplist`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: queryServiceMetricsViewComparisonToplistBody,
  });
};

export type QueryServiceMetricsViewComparisonToplistMutationResult =
  NonNullable<
    Awaited<ReturnType<typeof queryServiceMetricsViewComparisonToplist>>
  >;
export type QueryServiceMetricsViewComparisonToplistMutationBody =
  QueryServiceMetricsViewComparisonToplistBody;
export type QueryServiceMetricsViewComparisonToplistMutationError = RpcStatus;

export const createQueryServiceMetricsViewComparisonToplist = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof queryServiceMetricsViewComparisonToplist>>,
    TError,
    {
      instanceId: string;
      metricsViewName: string;
      data: QueryServiceMetricsViewComparisonToplistBody;
    },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof queryServiceMetricsViewComparisonToplist>>,
    {
      instanceId: string;
      metricsViewName: string;
      data: QueryServiceMetricsViewComparisonToplistBody;
    }
  > = (props) => {
    const { instanceId, metricsViewName, data } = props ?? {};

    return queryServiceMetricsViewComparisonToplist(
      instanceId,
      metricsViewName,
      data
    );
  };

  return createMutation<
    Awaited<ReturnType<typeof queryServiceMetricsViewComparisonToplist>>,
    TError,
    {
      instanceId: string;
      metricsViewName: string;
      data: QueryServiceMetricsViewComparisonToplistBody;
    },
    TContext
  >(mutationFn, mutationOptions);
};
/**
 * @summary MetricsViewTimeSeries returns time series for the measures in the metrics view.
It's a convenience API for querying a metrics view.
 */
export const queryServiceMetricsViewTimeSeries = (
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewTimeSeriesBody: QueryServiceMetricsViewTimeSeriesBody
) => {
  return httpClient<V1MetricsViewTimeSeriesResponse>({
    url: `/v1/instances/${instanceId}/queries/metrics-views/${metricsViewName}/timeseries`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: queryServiceMetricsViewTimeSeriesBody,
  });
};

export const getQueryServiceMetricsViewTimeSeriesQueryKey = (
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewTimeSeriesBody: QueryServiceMetricsViewTimeSeriesBody
) =>
  [
    `/v1/instances/${instanceId}/queries/metrics-views/${metricsViewName}/timeseries`,
    queryServiceMetricsViewTimeSeriesBody,
  ] as const;

export type QueryServiceMetricsViewTimeSeriesQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceMetricsViewTimeSeries>>
>;
export type QueryServiceMetricsViewTimeSeriesQueryError = RpcStatus;

export const createQueryServiceMetricsViewTimeSeries = <
  TData = Awaited<ReturnType<typeof queryServiceMetricsViewTimeSeries>>,
  TError = RpcStatus
>(
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewTimeSeriesBody: QueryServiceMetricsViewTimeSeriesBody,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceMetricsViewTimeSeries>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceMetricsViewTimeSeriesQueryKey(
      instanceId,
      metricsViewName,
      queryServiceMetricsViewTimeSeriesBody
    );

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceMetricsViewTimeSeries>>
  > = () =>
    queryServiceMetricsViewTimeSeries(
      instanceId,
      metricsViewName,
      queryServiceMetricsViewTimeSeriesBody
    );

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceMetricsViewTimeSeries>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && metricsViewName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary MetricsViewToplist returns the top dimension values of a metrics view sorted by one or more measures.
It's a convenience API for querying a metrics view.
 */
export const queryServiceMetricsViewToplist = (
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewToplistBody: QueryServiceMetricsViewToplistBody
) => {
  return httpClient<V1MetricsViewToplistResponse>({
    url: `/v1/instances/${instanceId}/queries/metrics-views/${metricsViewName}/toplist`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: queryServiceMetricsViewToplistBody,
  });
};

export const getQueryServiceMetricsViewToplistQueryKey = (
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewToplistBody: QueryServiceMetricsViewToplistBody
) =>
  [
    `/v1/instances/${instanceId}/queries/metrics-views/${metricsViewName}/toplist`,
    queryServiceMetricsViewToplistBody,
  ] as const;

export type QueryServiceMetricsViewToplistQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceMetricsViewToplist>>
>;
export type QueryServiceMetricsViewToplistQueryError = RpcStatus;

export const createQueryServiceMetricsViewToplist = <
  TData = Awaited<ReturnType<typeof queryServiceMetricsViewToplist>>,
  TError = RpcStatus
>(
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewToplistBody: QueryServiceMetricsViewToplistBody,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceMetricsViewToplist>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceMetricsViewToplistQueryKey(
      instanceId,
      metricsViewName,
      queryServiceMetricsViewToplistBody
    );

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceMetricsViewToplist>>
  > = () =>
    queryServiceMetricsViewToplist(
      instanceId,
      metricsViewName,
      queryServiceMetricsViewToplistBody
    );

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceMetricsViewToplist>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && metricsViewName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary MetricsViewTotals returns totals over a time period for the measures in a metrics view.
It's a convenience API for querying a metrics view.
 */
export const queryServiceMetricsViewTotals = (
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewTotalsBody: QueryServiceMetricsViewTotalsBody
) => {
  return httpClient<V1MetricsViewTotalsResponse>({
    url: `/v1/instances/${instanceId}/queries/metrics-views/${metricsViewName}/totals`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: queryServiceMetricsViewTotalsBody,
  });
};

export const getQueryServiceMetricsViewTotalsQueryKey = (
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewTotalsBody: QueryServiceMetricsViewTotalsBody
) =>
  [
    `/v1/instances/${instanceId}/queries/metrics-views/${metricsViewName}/totals`,
    queryServiceMetricsViewTotalsBody,
  ] as const;

export type QueryServiceMetricsViewTotalsQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceMetricsViewTotals>>
>;
export type QueryServiceMetricsViewTotalsQueryError = RpcStatus;

export const createQueryServiceMetricsViewTotals = <
  TData = Awaited<ReturnType<typeof queryServiceMetricsViewTotals>>,
  TError = RpcStatus
>(
  instanceId: string,
  metricsViewName: string,
  queryServiceMetricsViewTotalsBody: QueryServiceMetricsViewTotalsBody,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceMetricsViewTotals>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceMetricsViewTotalsQueryKey(
      instanceId,
      metricsViewName,
      queryServiceMetricsViewTotalsBody
    );

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceMetricsViewTotals>>
  > = () =>
    queryServiceMetricsViewTotals(
      instanceId,
      metricsViewName,
      queryServiceMetricsViewTotalsBody
    );

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceMetricsViewTotals>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && metricsViewName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary Get the number of nulls in a column
 */
export const queryServiceColumnNullCount = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnNullCountParams,
  signal?: AbortSignal
) => {
  return httpClient<V1ColumnNullCountResponse>({
    url: `/v1/instances/${instanceId}/queries/null-count/tables/${tableName}`,
    method: "get",
    params,
    signal,
  });
};

export const getQueryServiceColumnNullCountQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnNullCountParams
) =>
  [
    `/v1/instances/${instanceId}/queries/null-count/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceColumnNullCountQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnNullCount>>
>;
export type QueryServiceColumnNullCountQueryError = RpcStatus;

export const createQueryServiceColumnNullCount = <
  TData = Awaited<ReturnType<typeof queryServiceColumnNullCount>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnNullCountParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnNullCount>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnNullCountQueryKey(instanceId, tableName, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnNullCount>>
  > = ({ signal }) =>
    queryServiceColumnNullCount(instanceId, tableName, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnNullCount>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary Get the histogram for values in a column
 */
export const queryServiceColumnNumericHistogram = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnNumericHistogramParams,
  signal?: AbortSignal
) => {
  return httpClient<V1ColumnNumericHistogramResponse>({
    url: `/v1/instances/${instanceId}/queries/numeric-histogram/tables/${tableName}`,
    method: "get",
    params,
    signal,
  });
};

export const getQueryServiceColumnNumericHistogramQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnNumericHistogramParams
) =>
  [
    `/v1/instances/${instanceId}/queries/numeric-histogram/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceColumnNumericHistogramQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnNumericHistogram>>
>;
export type QueryServiceColumnNumericHistogramQueryError = RpcStatus;

export const createQueryServiceColumnNumericHistogram = <
  TData = Awaited<ReturnType<typeof queryServiceColumnNumericHistogram>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnNumericHistogramParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnNumericHistogram>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnNumericHistogramQueryKey(
      instanceId,
      tableName,
      params
    );

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnNumericHistogram>>
  > = ({ signal }) =>
    queryServiceColumnNumericHistogram(instanceId, tableName, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnNumericHistogram>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary ColumnRollupInterval returns the minimum time granularity (as well as the time range) for a specified timestamp column
 */
export const queryServiceColumnRollupInterval = (
  instanceId: string,
  tableName: string,
  queryServiceColumnRollupIntervalBody: QueryServiceColumnRollupIntervalBody
) => {
  return httpClient<V1ColumnRollupIntervalResponse>({
    url: `/v1/instances/${instanceId}/queries/rollup-interval/tables/${tableName}`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: queryServiceColumnRollupIntervalBody,
  });
};

export const getQueryServiceColumnRollupIntervalQueryKey = (
  instanceId: string,
  tableName: string,
  queryServiceColumnRollupIntervalBody: QueryServiceColumnRollupIntervalBody
) =>
  [
    `/v1/instances/${instanceId}/queries/rollup-interval/tables/${tableName}`,
    queryServiceColumnRollupIntervalBody,
  ] as const;

export type QueryServiceColumnRollupIntervalQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnRollupInterval>>
>;
export type QueryServiceColumnRollupIntervalQueryError = RpcStatus;

export const createQueryServiceColumnRollupInterval = <
  TData = Awaited<ReturnType<typeof queryServiceColumnRollupInterval>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  queryServiceColumnRollupIntervalBody: QueryServiceColumnRollupIntervalBody,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnRollupInterval>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnRollupIntervalQueryKey(
      instanceId,
      tableName,
      queryServiceColumnRollupIntervalBody
    );

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnRollupInterval>>
  > = () =>
    queryServiceColumnRollupInterval(
      instanceId,
      tableName,
      queryServiceColumnRollupIntervalBody
    );

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnRollupInterval>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary TableRows returns table rows
 */
export const queryServiceTableRows = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceTableRowsParams,
  signal?: AbortSignal
) => {
  return httpClient<V1TableRowsResponse>({
    url: `/v1/instances/${instanceId}/queries/rows/tables/${tableName}`,
    method: "get",
    params,
    signal,
  });
};

export const getQueryServiceTableRowsQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceTableRowsParams
) =>
  [
    `/v1/instances/${instanceId}/queries/rows/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceTableRowsQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceTableRows>>
>;
export type QueryServiceTableRowsQueryError = RpcStatus;

export const createQueryServiceTableRows = <
  TData = Awaited<ReturnType<typeof queryServiceTableRows>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceTableRowsParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceTableRows>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceTableRowsQueryKey(instanceId, tableName, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceTableRows>>
  > = ({ signal }) =>
    queryServiceTableRows(instanceId, tableName, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceTableRows>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary Get outliers for a numeric column
 */
export const queryServiceColumnRugHistogram = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnRugHistogramParams,
  signal?: AbortSignal
) => {
  return httpClient<V1ColumnRugHistogramResponse>({
    url: `/v1/instances/${instanceId}/queries/rug-histogram/tables/${tableName}`,
    method: "get",
    params,
    signal,
  });
};

export const getQueryServiceColumnRugHistogramQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnRugHistogramParams
) =>
  [
    `/v1/instances/${instanceId}/queries/rug-histogram/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceColumnRugHistogramQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnRugHistogram>>
>;
export type QueryServiceColumnRugHistogramQueryError = RpcStatus;

export const createQueryServiceColumnRugHistogram = <
  TData = Awaited<ReturnType<typeof queryServiceColumnRugHistogram>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnRugHistogramParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnRugHistogram>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnRugHistogramQueryKey(instanceId, tableName, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnRugHistogram>>
  > = ({ signal }) =>
    queryServiceColumnRugHistogram(instanceId, tableName, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnRugHistogram>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary Estimates the smallest time grain present in the column
 */
export const queryServiceColumnTimeGrain = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnTimeGrainParams,
  signal?: AbortSignal
) => {
  return httpClient<V1ColumnTimeGrainResponse>({
    url: `/v1/instances/${instanceId}/queries/smallest-time-grain/tables/${tableName}`,
    method: "get",
    params,
    signal,
  });
};

export const getQueryServiceColumnTimeGrainQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnTimeGrainParams
) =>
  [
    `/v1/instances/${instanceId}/queries/smallest-time-grain/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceColumnTimeGrainQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnTimeGrain>>
>;
export type QueryServiceColumnTimeGrainQueryError = RpcStatus;

export const createQueryServiceColumnTimeGrain = <
  TData = Awaited<ReturnType<typeof queryServiceColumnTimeGrain>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnTimeGrainParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnTimeGrain>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnTimeGrainQueryKey(instanceId, tableName, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnTimeGrain>>
  > = ({ signal }) =>
    queryServiceColumnTimeGrain(instanceId, tableName, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnTimeGrain>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary TableCardinality returns row count
 */
export const queryServiceTableCardinality = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceTableCardinalityParams,
  signal?: AbortSignal
) => {
  return httpClient<V1TableCardinalityResponse>({
    url: `/v1/instances/${instanceId}/queries/table-cardinality/tables/${tableName}`,
    method: "get",
    params,
    signal,
  });
};

export const getQueryServiceTableCardinalityQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceTableCardinalityParams
) =>
  [
    `/v1/instances/${instanceId}/queries/table-cardinality/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceTableCardinalityQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceTableCardinality>>
>;
export type QueryServiceTableCardinalityQueryError = RpcStatus;

export const createQueryServiceTableCardinality = <
  TData = Awaited<ReturnType<typeof queryServiceTableCardinality>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceTableCardinalityParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceTableCardinality>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceTableCardinalityQueryKey(instanceId, tableName, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceTableCardinality>>
  > = ({ signal }) =>
    queryServiceTableCardinality(instanceId, tableName, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceTableCardinality>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary Get the time range summaries (min, max) for a column
 */
export const queryServiceColumnTimeRange = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnTimeRangeParams,
  signal?: AbortSignal
) => {
  return httpClient<V1ColumnTimeRangeResponse>({
    url: `/v1/instances/${instanceId}/queries/time-range-summary/tables/${tableName}`,
    method: "get",
    params,
    signal,
  });
};

export const getQueryServiceColumnTimeRangeQueryKey = (
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnTimeRangeParams
) =>
  [
    `/v1/instances/${instanceId}/queries/time-range-summary/tables/${tableName}`,
    ...(params ? [params] : []),
  ] as const;

export type QueryServiceColumnTimeRangeQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnTimeRange>>
>;
export type QueryServiceColumnTimeRangeQueryError = RpcStatus;

export const createQueryServiceColumnTimeRange = <
  TData = Awaited<ReturnType<typeof queryServiceColumnTimeRange>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  params?: QueryServiceColumnTimeRangeParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnTimeRange>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnTimeRangeQueryKey(instanceId, tableName, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnTimeRange>>
  > = ({ signal }) =>
    queryServiceColumnTimeRange(instanceId, tableName, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnTimeRange>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary Generate time series for the given measures (aggregation expressions) along with the sparkline timeseries
 */
export const queryServiceColumnTimeSeries = (
  instanceId: string,
  tableName: string,
  queryServiceColumnTimeSeriesBody: QueryServiceColumnTimeSeriesBody
) => {
  return httpClient<V1ColumnTimeSeriesResponse>({
    url: `/v1/instances/${instanceId}/queries/timeseries/tables/${tableName}`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: queryServiceColumnTimeSeriesBody,
  });
};

export const getQueryServiceColumnTimeSeriesQueryKey = (
  instanceId: string,
  tableName: string,
  queryServiceColumnTimeSeriesBody: QueryServiceColumnTimeSeriesBody
) =>
  [
    `/v1/instances/${instanceId}/queries/timeseries/tables/${tableName}`,
    queryServiceColumnTimeSeriesBody,
  ] as const;

export type QueryServiceColumnTimeSeriesQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnTimeSeries>>
>;
export type QueryServiceColumnTimeSeriesQueryError = RpcStatus;

export const createQueryServiceColumnTimeSeries = <
  TData = Awaited<ReturnType<typeof queryServiceColumnTimeSeries>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  queryServiceColumnTimeSeriesBody: QueryServiceColumnTimeSeriesBody,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnTimeSeries>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnTimeSeriesQueryKey(
      instanceId,
      tableName,
      queryServiceColumnTimeSeriesBody
    );

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnTimeSeries>>
  > = () =>
    queryServiceColumnTimeSeries(
      instanceId,
      tableName,
      queryServiceColumnTimeSeriesBody
    );

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnTimeSeries>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary Get TopK elements from a table for a column given an agg function
agg function and k are optional, defaults are count(*) and 50 respectively
 */
export const queryServiceColumnTopK = (
  instanceId: string,
  tableName: string,
  queryServiceColumnTopKBody: QueryServiceColumnTopKBody
) => {
  return httpClient<V1ColumnTopKResponse>({
    url: `/v1/instances/${instanceId}/queries/topk/tables/${tableName}`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: queryServiceColumnTopKBody,
  });
};

export const getQueryServiceColumnTopKQueryKey = (
  instanceId: string,
  tableName: string,
  queryServiceColumnTopKBody: QueryServiceColumnTopKBody
) =>
  [
    `/v1/instances/${instanceId}/queries/topk/tables/${tableName}`,
    queryServiceColumnTopKBody,
  ] as const;

export type QueryServiceColumnTopKQueryResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceColumnTopK>>
>;
export type QueryServiceColumnTopKQueryError = RpcStatus;

export const createQueryServiceColumnTopK = <
  TData = Awaited<ReturnType<typeof queryServiceColumnTopK>>,
  TError = RpcStatus
>(
  instanceId: string,
  tableName: string,
  queryServiceColumnTopKBody: QueryServiceColumnTopKBody,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof queryServiceColumnTopK>>,
      TError,
      TData
    >;
  }
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getQueryServiceColumnTopKQueryKey(
      instanceId,
      tableName,
      queryServiceColumnTopKBody
    );

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof queryServiceColumnTopK>>
  > = () =>
    queryServiceColumnTopK(instanceId, tableName, queryServiceColumnTopKBody);

  const query = createQuery<
    Awaited<ReturnType<typeof queryServiceColumnTopK>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!(instanceId && tableName),
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary Query runs a SQL query against the instance's OLAP datastore.
 */
export const queryServiceQuery = (
  instanceId: string,
  queryServiceQueryBody: QueryServiceQueryBody
) => {
  return httpClient<V1QueryResponse>({
    url: `/v1/instances/${instanceId}/query`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: queryServiceQueryBody,
  });
};

export type QueryServiceQueryMutationResult = NonNullable<
  Awaited<ReturnType<typeof queryServiceQuery>>
>;
export type QueryServiceQueryMutationBody = QueryServiceQueryBody;
export type QueryServiceQueryMutationError = RpcStatus;

export const createQueryServiceQuery = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof queryServiceQuery>>,
    TError,
    { instanceId: string; data: QueryServiceQueryBody },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof queryServiceQuery>>,
    { instanceId: string; data: QueryServiceQueryBody }
  > = (props) => {
    const { instanceId, data } = props ?? {};

    return queryServiceQuery(instanceId, data);
  };

  return createMutation<
    Awaited<ReturnType<typeof queryServiceQuery>>,
    TError,
    { instanceId: string; data: QueryServiceQueryBody },
    TContext
  >(mutationFn, mutationOptions);
};
