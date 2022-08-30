/**
 * TODO: Instead of writing this file by hand, a better approach would be to use an OpenAPI spec and
 * autogenerate `svelte-query`-specific client code. One such tool is: https://orval.dev/guides/svelte-query
 */

import type {
  MetricViewMetaResponse,
  MetricViewTimeSeriesRequest,
  MetricViewTimeSeriesResponse,
  MetricViewTopListRequest,
  MetricViewTopListResponse,
  MetricViewTotalsRequest,
  MetricViewTotalsResponse,
} from "$common/rill-developer-service/MetricViewActions";
import { config } from "$lib/application-state-stores/application-store";
import { queriesRepository } from "$lib/svelte-query/queries/QueriesRepository";
import {
  QueryClient,
  useQuery,
  UseQueryOptions,
  UseQueryStoreResult,
} from "@sveltestack/svelte-query";

async function fetchUrl(path: string, method: string, body?) {
  const resp = await fetch(`${config.server.serverUrl}/api/v1/${path}`, {
    method,
    headers: { "Content-Type": "application/json" },
    ...(body ? { body: JSON.stringify(body) } : {}),
  });
  const json = await resp.json();
  if (!resp.ok) {
    const err = new Error(json.messages[0].message);
    return Promise.reject(err);
  }
  return json;
}

// GET /api/v1/metric-views/{view-name}/meta

export const getMetricsViewMetadata = async (
  metricViewId: string
): Promise<MetricViewMetaResponse> => {
  const json = await fetchUrl(`metric-views/${metricViewId}/meta`, "GET");
  json.id = metricViewId;
  return json;
};

const MetaId = `v1/metric-view/meta`;
export const getMetaQueryKey = (metricViewId: string) => {
  return [MetaId, metricViewId];
};

export const useMetaQuery = (
  metricViewId: string,
  queryOptions: UseQueryOptions<MetricViewMetaResponse, Error> = {}
) => {
  const queryKey = queryOptions?.queryKey ?? getMetaQueryKey(metricViewId);
  const queryFn = () => getMetricsViewMetadata(metricViewId);
  const query = queriesRepository.useQuery<MetricViewMetaResponse, Error>(
    queryKey,
    queryFn,
    {
      ...queryOptions,
      enabled: !!metricViewId,
    }
  ) as UseQueryStoreResult<MetricViewMetaResponse, Error>;
  return {
    queryKey,
    ...query,
  };
};

// POST /api/v1/metric-views/{view-name}/timeseries

export const getMetricsViewTimeSeries = async (
  metricViewId: string,
  request: MetricViewTimeSeriesRequest
): Promise<MetricViewTimeSeriesResponse> => {
  return fetchUrl(`metric-views/${metricViewId}/timeseries`, "POST", request);
};

const TimeSeriesId = `v1/metric-view/timeseries`;
export const getTimeSeriesQueryKey = (
  metricViewId: string,
  request: MetricViewTimeSeriesRequest
) => {
  return [TimeSeriesId, metricViewId, request];
};

export const useTimeSeriesQuery = (
  metricViewId: string,
  request: MetricViewTimeSeriesRequest,
  queryOptions: UseQueryOptions<MetricViewTimeSeriesResponse, Error> = {}
) => {
  const queryKey =
    queryOptions?.queryKey ?? getTimeSeriesQueryKey(metricViewId, request);
  const queryFn = () => getMetricsViewTimeSeries(metricViewId, request);
  const query = queriesRepository.useQuery<MetricViewTimeSeriesResponse, Error>(
    queryKey,
    queryFn,
    {
      ...queryOptions,
      enabled:
        !!(metricViewId && request.measures && request.time) &&
        (!("enabled" in queryOptions) || queryOptions.enabled),
    }
  ) as UseQueryStoreResult<MetricViewTimeSeriesResponse, Error>;
  return {
    queryKey,
    ...query,
  };
};

// POST /api/v1/metric-views/{view-name}/toplist/{dimension}

export const getMetricsViewTopList = async (
  metricViewId: string,
  dimensionId: string,
  request: MetricViewTopListRequest
): Promise<MetricViewTopListResponse> => {
  if (!metricViewId || !dimensionId)
    return {
      meta: [],
      data: [],
    };

  return fetchUrl(
    `metric-views/${metricViewId}/toplist/${dimensionId}`,
    "POST",
    request
  );
};

const TopListId = `v1/metric-view/toplist`;
export const getTopListQueryKey = (
  metricViewId: string,
  dimensionId: string,
  request: MetricViewTopListRequest
) => {
  return [TopListId, metricViewId, dimensionId, request];
};

function getTopListQueryOptions(
  metricsDefId,
  dimensionId,
  topListQueryRequest
) {
  return {
    staleTime: 30 * 1000,
    enabled: !!(
      metricsDefId &&
      dimensionId &&
      topListQueryRequest.limit &&
      topListQueryRequest.measures.length >= 1 &&
      topListQueryRequest.offset !== undefined &&
      topListQueryRequest.sort &&
      topListQueryRequest.time
    ),
  };
}

/** custom hook to fetch a toplist result set, given a metricsDefId,
 * dimensionId,
 * and a request parameter.
 * The request parameter matches the API signature needed for the toplist request.
 */
export function useTopListQuery(metricsDefId, dimensionId, requestParameter) {
  const topListQueryKey = getTopListQueryKey(
    metricsDefId,
    dimensionId,
    requestParameter
  );
  const topListQueryFn = () => {
    return getMetricsViewTopList(metricsDefId, dimensionId, requestParameter);
  };
  const topListQueryOptions = getTopListQueryOptions(
    metricsDefId,
    dimensionId,
    requestParameter
  );
  return useQuery(topListQueryKey, topListQueryFn, topListQueryOptions);
}

// POST /api/v1/metric-views/{view-name}/totals

export const getMetricsViewTotals = async (
  metricViewId: string,
  request: MetricViewTotalsRequest
): Promise<MetricViewTotalsResponse> => {
  return fetchUrl(`metric-views/${metricViewId}/totals`, "POST", request);
};

const TotalsId = `v1/metric-view/totals`;
export const getTotalsQueryKey = (
  metricViewId: string,
  request: MetricViewTotalsRequest
) => {
  return [TotalsId, metricViewId, request];
};

export const useTotalsQuery = (
  metricViewId: string,
  request: MetricViewTotalsRequest
) => {
  const totalsQueryKey = getTotalsQueryKey(metricViewId, request);
  const totalsQueryFn = () => getMetricsViewTotals(metricViewId, request);
  return useQuery<MetricViewTotalsResponse, Error>(
    totalsQueryKey,
    totalsQueryFn,
    {
      staleTime: 30 * 1000,
      enabled: !!(metricViewId && request.measures && request.time),
    }
  );
};

// invalidation helpers

export const invalidateMetricView = async (
  queryClient: QueryClient,
  metricViewId: string
) => {
  // wait for meta to be invalidated
  console.log("invalidateMetricView", metricViewId);
  await queryClient.invalidateQueries([MetaId, metricViewId]);
  // invalidateMetricViewData(queryClient, metricViewId);
};

export const invalidateMetricViewData = (
  queryClient: QueryClient,
  metricViewId: string
) => {
  queryClient.invalidateQueries([TopListId, metricViewId]);
  queryClient.invalidateQueries([TotalsId, metricViewId]);
  queryClient.invalidateQueries([TimeSeriesId, metricViewId]);
};
