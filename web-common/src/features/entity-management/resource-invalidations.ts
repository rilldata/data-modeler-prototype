import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1WatchResourcesResponse } from "@rilldata/web-common/runtime-client";
import {
  V1ReconcileStatus,
  V1Resource,
  getConnectorServiceOLAPListTablesQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
} from "@rilldata/web-common/runtime-client";
import {
  invalidateMetricsViewData,
  invalidateProfilingQueries,
  invalidationForMetricsViewData,
} from "@rilldata/web-common/runtime-client/invalidation";
import { isProfilingQuery } from "@rilldata/web-common/runtime-client/query-matcher";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export const MainResourceKinds: {
  [kind in ResourceKind]?: true;
} = {
  [ResourceKind.Source]: true,
  [ResourceKind.Model]: true,
  [ResourceKind.MetricsView]: true,
  [ResourceKind.Chart]: true,
};
const UsedResourceKinds: {
  [kind in ResourceKind]?: true;
} = {
  [ResourceKind.ProjectParser]: true,
  [ResourceKind.Theme]: true,
  ...MainResourceKinds,
};

export function invalidateResourceResponse(
  queryClient: QueryClient,
  res: V1WatchResourcesResponse,
) {
  // only process for the `ResourceKind` present in `UsedResourceKinds`
  if (!res.name?.kind || !res.resource || !UsedResourceKinds[res.name.kind])
    return;

  const instanceId = get(runtime).instanceId;
  if (
    MainResourceKinds[res.name.kind] &&
    shouldSkipResource(queryClient, instanceId, res.resource)
  ) {
    return;
  }

  // Reconcile does a soft delete 1st by populating deletedOn
  // We then get an event with DELETE after reconcile ends, but without a resource object.
  // So we need to check for deletedOn to be able to use resource.meta, especially the filePaths
  const isSoftDelete = !!res.resource?.meta?.deletedOn;

  // invalidations will wait until the re-fetched query is completed
  // so, we should not `await` here
  if (isSoftDelete) {
    void invalidateRemovedResource(queryClient, instanceId, res.resource);
  } else {
    void invalidateResource(queryClient, instanceId, res.resource);
  }

  // only re-fetch list queries for kinds in `MainResources`
  if (!MainResourceKinds[res.name.kind]) return;
  return queryClient.refetchQueries(
    // we only use individual kind's queries
    getRuntimeServiceListResourcesQueryKey(instanceId, {
      kind: res.name.kind,
    }),
  );
}

async function invalidateResource(
  queryClient: QueryClient,
  instanceId: string,
  resource: V1Resource,
) {
  if (!resource.meta) return;
  void refreshResource(queryClient, instanceId, resource);

  const lastStateUpdatedOn = fileArtifacts.getFileArtifact(
    resource.meta?.filePaths?.[0] ?? "",
  ).lastStateUpdatedOn;
  if (
    resource.meta.reconcileStatus !== V1ReconcileStatus.RECONCILE_STATUS_IDLE &&
    !lastStateUpdatedOn
  ) {
    // When a resource is created it can send an event with status = IDLE just before it is queued for reconcile.
    // So handle the case when it is 1st queued and status != IDLE
    fileArtifacts.updateLastUpdated(resource);
    return;
  }

  // avoid refreshing for cases where event is sent for a resource that has not changed since we last saw it
  if (
    resource.meta.reconcileStatus !== V1ReconcileStatus.RECONCILE_STATUS_IDLE ||
    lastStateUpdatedOn === resource.meta.stateUpdatedOn
  )
    return;

  if (fileArtifacts.wasRenaming(resource)) {
    void queryClient.invalidateQueries(
      getConnectorServiceOLAPListTablesQueryKey(),
    );
  }
  fileArtifacts.updateArtifacts(resource);
  const failed = !!resource.meta.reconcileError;

  const name = resource.meta?.name?.name ?? "";
  switch (resource.meta.name?.kind) {
    case ResourceKind.Source:
      if (resource.source?.state?.table)
        // make sure table is populated
        return invalidateProfilingQueries(queryClient, name, failed);
      break;

    case ResourceKind.Model:
      if (resource.model?.state?.table)
        // make sure table is populated
        return invalidateProfilingQueries(queryClient, name, failed);
      break;

    case ResourceKind.MetricsView:
      return invalidateMetricsViewData(queryClient, name, failed);
  }
}

function invalidateRemovedResource(
  queryClient: QueryClient,
  instanceId: string,
  resource: V1Resource,
) {
  const name = resource.meta?.name?.name ?? "";
  queryClient.removeQueries(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": name,
      "name.kind": resource.meta?.name?.kind,
    }),
  );
  fileArtifacts.deleteResource(resource);
  // cancel queries to make sure any pending requests are cancelled.
  // There could still be some errors because of the race condition between a view/table deleted and we getting the event
  switch (resource?.meta?.name?.kind) {
    case ResourceKind.Source:
    case ResourceKind.Model:
      void queryClient.cancelQueries({
        predicate: (query) => isProfilingQuery(query, name),
      });
      void queryClient.invalidateQueries(
        getConnectorServiceOLAPListTablesQueryKey(),
      );
      break;
    case ResourceKind.MetricsView:
      void queryClient.cancelQueries({
        predicate: (query) => invalidationForMetricsViewData(query, name),
      });
      break;
  }
}

// We should not invalidate queries when resource is either queued or is running reconcile
function shouldSkipResource(
  queryClient: QueryClient,
  instanceId: string,
  res: V1Resource,
) {
  switch (res.meta?.reconcileStatus) {
    case V1ReconcileStatus.RECONCILE_STATUS_UNSPECIFIED:
      return true;

    case V1ReconcileStatus.RECONCILE_STATUS_PENDING:
    case V1ReconcileStatus.RECONCILE_STATUS_RUNNING:
      void refreshResource(queryClient, instanceId, res);
      fileArtifacts.updateReconciling(res);
      return true;
  }

  return false;
}

export function refreshResource(
  queryClient: QueryClient,
  instanceId: string,
  res: V1Resource,
) {
  return queryClient.resetQueries(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": res.meta?.name?.name,
      "name.kind": res.meta?.name?.kind,
    }),
  );
}

export async function invalidateAllResources(queryClient: QueryClient) {
  return queryClient.resetQueries({
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${get(runtime).instanceId}`),
  });
}
