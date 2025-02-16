import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getConnectorServiceOLAPListTablesQueryKey,
  getQueryServiceResolveCanvasQueryKey,
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  getRuntimeServiceGetModelPartitionsQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  type V1Resource,
  V1ResourceEvent,
  type V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import {
  invalidateComponentData,
  invalidateMetricsViewData,
  invalidateProfilingQueries,
} from "@rilldata/web-common/runtime-client/invalidation";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import { get } from "svelte/store";
import { connectorExplorerStore } from "../connectors/connector-explorer-store";
import { sourceImportedPath } from "../sources/sources-store";
import { isLeafResource } from "./dag-utils";

export class WatchResourcesClient {
  public readonly client: WatchRequestClient<V1WatchResourcesResponse>;
  private readonly instanceId = get(runtime).instanceId;
  private readonly connectorNames = new Set<string>();

  public constructor() {
    this.client = new WatchRequestClient<V1WatchResourcesResponse>();
    this.client.on("response", (res) => this.handleWatchResourceResponse(res));
    this.client.on("reconnect", () => this.invalidateAllRuntimeQueries());
  }

  private async handleWatchResourceResponse(res: V1WatchResourcesResponse) {
    // Log resource status to the browser console during e2e tests. Currently, our e2e tests make assertions
    // based on these logs. However, the e2e tests really should make UI-based assertions.
    if (import.meta.env.VITE_PLAYWRIGHT_TEST) {
      console.log(
        `[${res.resource?.meta?.reconcileStatus}] ${res.name?.kind}/${res.name?.name}`,
      );
    }

    // Type guards
    if (!res?.event || !res?.name || !res?.name?.name || !res?.name?.kind) {
      return;
    }

    // Get the previous resource from the query cache
    const previousResource = queryClient.getQueryData<{
      resource: V1Resource | undefined;
    }>(
      getRuntimeServiceGetResourceQueryKey(this.instanceId, {
        "name.name": res.name.name,
        "name.kind": res.name.kind,
      }),
    )?.resource;

    // Set the updated resource in the query cache
    queryClient.setQueryData(
      getRuntimeServiceGetResourceQueryKey(this.instanceId, {
        "name.name": res.name.name,
        "name.kind": res.name.kind,
      }),
      {
        resource: res?.resource,
      },
    );

    // Nothing more to do for the ProjectParser resource
    if ((res.name.kind as ResourceKind) === ResourceKind.ProjectParser) return;

    // Update the file artifacts client-side cache (which maps files to resources)
    switch (res.event) {
      case V1ResourceEvent.RESOURCE_EVENT_WRITE:
        if (res.resource) {
          fileArtifacts.updateArtifacts(res.resource);
        }
        break;
      case V1ResourceEvent.RESOURCE_EVENT_DELETE:
        fileArtifacts.deleteResource(res.name);
        break;
    }

    switch (res.event) {
      case V1ResourceEvent.RESOURCE_EVENT_WRITE: {
        // Type guards
        if (
          !res?.resource ||
          !res?.resource?.meta ||
          !res?.resource?.meta?.reconcileStatus ||
          !res?.resource?.meta?.stateVersion
        ) {
          return;
        }

        // Proceed to query invalidations only when the resource state has changed
        if (
          res.resource.meta.stateVersion ===
          previousResource?.meta?.stateVersion
        )
          return;

        // Refetch `ListResources` queries
        void queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(this.instanceId, undefined),
        );
        void queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(this.instanceId, {
            kind: res.name.kind,
          }),
        );

        switch (res.name.kind as ResourceKind) {
          case ResourceKind.Connector:
            // Invalidate the list of connectors
            void queryClient.invalidateQueries(
              getRuntimeServiceAnalyzeConnectorsQueryKey(this.instanceId),
            );

            // Invalidate the connector's list of tables
            void queryClient.invalidateQueries(
              getConnectorServiceOLAPListTablesQueryKey({
                instanceId: this.instanceId,
                connector: res.name.name,
              }),
            );

            // Done
            return;

          case ResourceKind.Source:
          case ResourceKind.Model: {
            // TODO: differentiate between a Model's executorConnector and resultConnector
            const connectorName =
              (res.name.kind as ResourceKind) === ResourceKind.Source
                ? res.resource.source?.state?.connector
                : res.resource.model?.state?.resultConnector;
            const previousConnectorName =
              (res.name.kind as ResourceKind) === ResourceKind.Source
                ? previousResource?.source?.state?.connector
                : previousResource?.model?.state?.resultConnector;

            // If the result table has changed, invalidate the connector's list of tables
            const sourceTableChanged =
              res.resource?.source?.state?.table !==
              previousResource?.source?.state?.table;
            const modelResultTableChanged =
              res.resource.model?.state?.resultTable !==
              previousResource?.model?.state?.resultTable;
            if (sourceTableChanged || modelResultTableChanged) {
              const connectorsToInvalidate = Array.from(
                new Set([connectorName, previousConnectorName].filter(Boolean)),
              );
              for (const connector of connectorsToInvalidate) {
                void queryClient.invalidateQueries(
                  getConnectorServiceOLAPListTablesQueryKey({
                    instanceId: this.instanceId,
                    connector: connector,
                  }),
                );
              }
            }

            // If the connector is new, invalidate the list of connectors
            // (This is needed because Sources and Models can implicitly create Connectors)
            if (connectorName && !this.connectorNames.has(connectorName)) {
              this.connectorNames.add(connectorName);
              void queryClient.invalidateQueries(
                getRuntimeServiceAnalyzeConnectorsQueryKey(this.instanceId),
              );
            }

            // Note: Sources/Models that fail to ingest will not have a table name
            const tableName =
              (res.name.kind as ResourceKind) === ResourceKind.Source
                ? res.resource.source?.state?.table
                : res.resource.model?.state?.resultTable;

            // The following invalidations are only needed if the Source/Model has an active table
            if (!connectorName || !tableName) return;

            // If it's a new source, show the "Source imported successfully" modal
            const isNewSource =
              res.name.kind === ResourceKind.Source &&
              res.resource.meta.specVersion === "1" && // First file version
              res.resource.meta.stateVersion === "2" && // First ingest is complete
              (await isLeafResource(res.resource, this.instanceId)); // Protects against existing projects reconciling anew
            if (isNewSource) {
              const filePath = res.resource?.meta?.filePaths?.[0] as string;
              sourceImportedPath.set(filePath);
            }

            // Invalidate the model partitions query
            if ((res.name.kind as ResourceKind) === ResourceKind.Model) {
              void queryClient.invalidateQueries(
                getRuntimeServiceGetModelPartitionsQueryKey(
                  this.instanceId,
                  res.name.name,
                ),
              );
            }

            // Invalidate profiling queries
            const failed = !!res.resource.meta?.reconcileError;
            void invalidateProfilingQueries(queryClient, tableName, failed);

            // Done
            return;
          }

          case ResourceKind.MetricsView: {
            const failed = !!res.resource.meta?.reconcileError;
            void invalidateMetricsViewData(queryClient, res.name.name, failed);

            // Done
            return;
          }

          case ResourceKind.Explore: {
            const failed = !!res.resource.meta?.reconcileError;
            if (res.resource.explore?.state?.validSpec?.metricsView) {
              void invalidateMetricsViewData(
                queryClient,
                res.resource.explore.state.validSpec.metricsView,
                failed,
              );
            }

            queryClient
              .refetchQueries(
                getRuntimeServiceGetExploreQueryKey(this.instanceId, {
                  name: res.name.name,
                }),
              )
              .catch(console.error);
            return;
          }

          case ResourceKind.Canvas: {
            void queryClient.refetchQueries(
              getQueryServiceResolveCanvasQueryKey(
                this.instanceId,
                res.name.name,
                {},
              ),
            );
            return;
          }

          case ResourceKind.Component: {
            const failed = !!res.resource.meta?.reconcileError;
            void invalidateComponentData(queryClient, res.name.name, failed);

            // Done
            return;
          }

          default:
            // No specific handling for the given resource kind
            return;
        }
      }

      /**
       * Note: Resource "deletes" occur in two stages:
       * 1. A `WRITE` event marks the resource for deletion by setting the `deletedOn` property ("soft" delete).
       * 2. A `DELETE` event signals that the resource has actually been deleted ("hard" delete).
       */
      case V1ResourceEvent.RESOURCE_EVENT_DELETE:
        // Refetch `ListResources` queries
        void queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(this.instanceId, undefined),
        );
        void queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(this.instanceId, {
            kind: res.name.kind,
          }),
        );

        switch (res.name.kind as ResourceKind) {
          case ResourceKind.Connector:
            // Invalidate the list of connectors
            void queryClient.invalidateQueries(
              getRuntimeServiceAnalyzeConnectorsQueryKey(this.instanceId),
            );

            // Remove the connector's state from the connector explorer store
            connectorExplorerStore.deleteItem(res.name.name);

            // Done
            return;

          case ResourceKind.Source:
          case ResourceKind.Model: {
            // Get the now-deleted resource's connector name
            const connectorName =
              (res.name.kind as ResourceKind) === ResourceKind.Source
                ? previousResource?.source?.state?.connector
                : previousResource?.model?.state?.resultConnector;

            // Invalidate the connector's list of tables
            void queryClient.invalidateQueries(
              getConnectorServiceOLAPListTablesQueryKey({
                instanceId: this.instanceId,
                connector: connectorName,
              }),
            );

            // Done
            return;
          }

          default:
            // No specific handling for the given resource kind
            return;
        }
    }
  }

  private invalidateAllRuntimeQueries() {
    return queryClient.invalidateQueries({
      predicate: (query) =>
        query.queryHash.includes(`v1/instances/${get(runtime).instanceId}`),
    });
  }
}
