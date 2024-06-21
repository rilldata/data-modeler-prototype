import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceGetResource,
  runtimeServiceListResources,
  V1ListResourcesResponse,
  V1ReconcileStatus,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export enum ResourceKind {
  ProjectParser = "rill.runtime.v1.ProjectParser",
  Source = "rill.runtime.v1.Source",
  Connector = "rill.runtime.v1.Connector",
  Model = "rill.runtime.v1.Model",
  MetricsView = "rill.runtime.v1.MetricsView",
  Report = "rill.runtime.v1.Report",
  Alert = "rill.runtime.v1.Alert",
  Theme = "rill.runtime.v1.Theme",
  Component = "rill.runtime.v1.Component",
  Dashboard = "rill.runtime.v1.Dashboard",
  API = "rill.runtime.v1.API",
}
export type UserFacingResourceKinds = Exclude<
  ResourceKind,
  ResourceKind.ProjectParser
>;
export const SingletonProjectParserName = "parser";
export const ResourceShortNameToKind: Record<string, ResourceKind> = {
  source: ResourceKind.Source,
  model: ResourceKind.Model,
  metricsview: ResourceKind.MetricsView,
  metrics_view: ResourceKind.MetricsView,
  component: ResourceKind.Component,
  dashboard: ResourceKind.Dashboard,
  report: ResourceKind.Report,
  alert: ResourceKind.Alert,
  theme: ResourceKind.Theme,
  api: ResourceKind.API,
};

// In the UI, we shouldn't show the `rill.runtime.v1` prefix
export function prettyResourceKind(kind: string) {
  return kind.replace(/^rill\.runtime\.v1\./, "");
}

export function useResource<T = V1Resource>(
  instanceId: string,
  name: string,
  kind: ResourceKind,
  selector?: (data: V1Resource) => T,
  queryClient?: QueryClient,
) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": kind,
      "name.name": name,
    },
    {
      query: {
        select: (data) =>
          (selector ? selector(data?.resource) : data?.resource) as T,
        enabled: !!instanceId && !!name && !!kind,
        queryClient,
      },
    },
  );
}

export function useProjectParser(queryClient: QueryClient, instanceId: string) {
  return useResource(
    instanceId,
    SingletonProjectParserName,
    ResourceKind.ProjectParser,
    undefined,
    queryClient,
  );
}

export function useFilteredResources<T = Array<V1Resource>>(
  instanceId: string,
  kind: ResourceKind,
  selector: (data: V1ListResourcesResponse) => T = (data) =>
    data.resources as T,
) {
  return createRuntimeServiceListResources(
    instanceId,
    {
      kind,
    },
    {
      query: {
        select: selector,
      },
    },
  );
}

/**
 * Fetches all resources and filters them client side.
 * This is to improve network requests since we need the full list all the time as well.
 */
export function useClientFilteredResources(
  instanceId: string,
  kind: ResourceKind,
  filter: (res: V1Resource) => boolean = () => true,
) {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) =>
        data.resources?.filter(
          (res) => res.meta?.name?.kind === kind && filter(res),
        ) ?? [],
    },
  });
}

export function resourceIsLoading(resource?: V1Resource) {
  return (
    !!resource &&
    resource.meta?.reconcileStatus !== V1ReconcileStatus.RECONCILE_STATUS_IDLE
  );
}

export async function fetchResource(
  queryClient: QueryClient,
  instanceId: string,
  name: string,
  kind: ResourceKind,
) {
  const resp = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": name,
      "name.kind": kind,
    }),
    queryFn: () =>
      runtimeServiceGetResource(instanceId, {
        "name.name": name,
        "name.kind": kind,
      }),
  });
  return resp.resource;
}

export async function fetchResources(
  queryClient: QueryClient,
  instanceId: string,
) {
  const resp = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(instanceId),
    queryFn: () => runtimeServiceListResources(instanceId, {}),
  });
  return resp.resources ?? [];
}
