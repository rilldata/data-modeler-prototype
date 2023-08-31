import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import {
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  V1Resource,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export function watchResourcesClient(queryClient: QueryClient) {
  return new WatchRequestClient<V1WatchResourcesResponse>(
    (runtime) =>
      `${runtime.host}/v1/instances/${runtime.instanceId}/resources/-/watch`,
    (res) => invalidateResourceResponse(queryClient, res),
    () => invalidateAllResources(queryClient)
  ).start();
}

function invalidateResourceResponse(
  queryClient: QueryClient,
  res: V1WatchResourcesResponse
) {
  if (!res.resource) return;

  const instanceId = get(runtime).instanceId;
  // invalidations will wait until the re-fetched query is completed
  // so, we should not `await` here
  switch (res.event) {
    case "RESOURCE_EVENT_ADDED":
      queryClient.refetchQueries(
        getRuntimeServiceListResourcesQueryKey(instanceId)
      );
    // eslint-disable-next-line no-fallthrough
    case "RESOURCE_EVENT_UPDATED_SPEC":
    case "RESOURCE_EVENT_UPDATED_STATE":
      invalidateResource(queryClient, instanceId, res.resource);
      break;

    case "RESOURCE_EVENT_DELETED":
      invalidateRemovedResource(queryClient, instanceId, res.resource);
      queryClient.refetchQueries(
        getRuntimeServiceListResourcesQueryKey(instanceId)
      );
      break;
  }
}

async function invalidateResource(
  queryClient: QueryClient,
  instanceId: string,
  resource: V1Resource
) {
  return queryClient.refetchQueries(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": resource.meta.name.name,
      "name.kind": resource.meta.name.kind,
    })
  );
  // TODO: invalidate individual queries when we swap over
}

async function invalidateRemovedResource(
  queryClient: QueryClient,
  instanceId: string,
  resource: V1Resource
) {
  queryClient.removeQueries(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": resource.meta.name.name,
      "name.kind": resource.meta.name.kind,
    })
  );
  // TODO: remove individual queries when we swap over
}

async function invalidateAllResources(queryClient: QueryClient) {
  const instanceId = get(runtime).instanceId;
  queryClient.removeQueries({
    type: "inactive",
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${instanceId}/resources`),
  });

  return queryClient.refetchQueries({
    type: "active",
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${instanceId}/resources`),
  });
  // TODO: invalidate individual queries when we swap over
}
