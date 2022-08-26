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
import type { QueryClient, UseQueryOptions } from "@sveltestack/svelte-query";
import { queriesRepository } from "$lib/svelte-query/queries/QueriesRepository";
import { useQuery } from "@sveltestack/svelte-query";

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

export const getMetricViewMetadata = async (
  metricViewId: string
): Promise<MetricViewMetaResponse> => {
  const json = await fetchUrl(`metric-views/${metricViewId}/meta`, "GET");
  json.id = metricViewId;
  return json;
};

const MetaId = `v1/metric-view/meta`;
export const getMetricViewMetaQueryKey = (metricViewId: string) => {
  return [MetaId, metricViewId];
};

export const useGetMetricViewMeta = (
  metricViewId: string,
  queryOptions?: UseQueryOptions<MetricViewMetaResponse, Error>
) => {
  const queryKey =
    queryOptions?.queryKey ?? getMetricViewMetaQueryKey(metricViewId);
  const queryFn = () => getMetricViewMetadata(metricViewId);
  const query = useQuery<MetricViewMetaResponse, Error>(queryKey, queryFn, {
    enabled: !!metricViewId,
    ...queryOptions,
  });
  return {
    queryKey,
    ...query,
  };
};

// POST /api/v1/metric-views/{view-name}/timeseries

export const getMetricViewTimeSeries = async (
  metricViewId: string,
  request: MetricViewTimeSeriesRequest
): Promise<MetricViewTimeSeriesResponse> => {
  return fetchUrl(`metric-views/${metricViewId}/timeseries`, "POST", request);
};

const TimeSeriesId = `v1/metric-view/timeseries`;
export const getMetricViewTimeSeriesQueryKey = (
  metricViewId: string,
  request: MetricViewTimeSeriesRequest
) => {
  return [TimeSeriesId, metricViewId, request];
};

export const useGetMetricViewTimeSeries = (
  metricViewId: string,
  request: MetricViewTimeSeriesRequest,
  queryOptions?: UseQueryOptions<MetricViewTimeSeriesResponse, Error>
) => {
  const queryKey =
    queryOptions?.queryKey ??
    getMetricViewTimeSeriesQueryKey(metricViewId, request);
  const queryFn = () => getMetricViewTimeSeries(metricViewId, request);
  const query = useQuery<MetricViewTimeSeriesResponse, Error>(
    queryKey,
    queryFn,
    {
      enabled: !!(metricViewId && request.measures && request.time),
      ...queryOptions,
    }
  );
  return {
    queryKey,
    ...query,
  };
};

// POST /api/v1/metric-views/{view-name}/toplist/{dimension}

export const getMetricViewTopList = async (
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
export const getMetricViewTopListQueryKey = (
  metricViewId: string,
  dimensionId: string,
  request: MetricViewTopListRequest
) => {
  return [TopListId, metricViewId, dimensionId, request];
};

export const useGetMetricViewTopList = (
  metricViewId: string,
  dimensionId: string,
  request: MetricViewTopListRequest,
  queryOptions?: UseQueryOptions<MetricViewTopListResponse, Error>
) => {
  const queryKey =
    queryOptions?.queryKey ??
    getMetricViewTopListQueryKey(metricViewId, dimensionId, request);
  const queryFn = () =>
    getMetricViewTopList(metricViewId, dimensionId, request);
  const query = useQuery<MetricViewTopListResponse, Error>(queryKey, queryFn, {
    enabled: !!(
      metricViewId &&
      dimensionId &&
      request.limit &&
      request.measures &&
      request.offset !== undefined &&
      request.sort &&
      request.time
    ),
    ...queryOptions,
  });
  return {
    queryKey,
    ...query,
  };
};

// POST /api/v1/metric-views/{view-name}/totals

export const getMetricViewTotals = async (
  metricViewId: string,
  request: MetricViewTotalsRequest
): Promise<MetricViewTotalsResponse> => {
  return fetchUrl(`metric-views/${metricViewId}/totals`, "POST", request);
};

const TotalsId = `v1/metric-view/totals`;
export const getMetricViewTotalsQueryKey = (
  metricViewId: string,
  isReferenceValue: boolean,
  request: MetricViewTotalsRequest
) => {
  return [TotalsId, metricViewId, isReferenceValue, request];
};

export const useGetMetricViewTotals = (
  metricViewId: string,
  request: MetricViewTotalsRequest,
  isReferenceValue = false,
  queryOptions?: UseQueryOptions<MetricViewTotalsResponse, Error>
) => {
  const queryKey =
    queryOptions?.queryKey ??
    getMetricViewTotalsQueryKey(metricViewId, isReferenceValue, request);
  const queryFn = () => getMetricViewTotals(metricViewId, request);
  const query = useQuery<MetricViewTotalsResponse, Error>(queryKey, queryFn, {
    enabled: !!(metricViewId && request.measures && request.time),
    ...queryOptions,
  });
  return {
    queryKey,
    ...query,
  };
};

// invalidation helpers

export const invalidateMetricView = async (
  queryClient: QueryClient,
  metricViewId: string
) => {
  // wait for meta to be invalidated
  await queryClient.invalidateQueries([MetaId, metricViewId]);
  invalidateMetricViewData(queryClient, metricViewId);
};

export const invalidateMetricViewData = (
  queryClient: QueryClient,
  metricViewId: string
) => {
  queryClient.invalidateQueries([TopListId, metricViewId]);
  queryClient.invalidateQueries([TotalsId, metricViewId]);
  queryClient.invalidateQueries([TimeSeriesId, metricViewId]);
};
