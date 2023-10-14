import type { V1GetProjectResponse } from "@rilldata/web-admin/client";
import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createRuntimeServiceListResources,
  type V1CatalogEntry,
  type V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import Axios from "axios";

export interface DashboardListItem {
  name: string;
  title?: string;
  description?: string;
  isValid: boolean;
}

export async function getDashboardsForProject(
  projectData: V1GetProjectResponse
): Promise<V1MetricsView[]> {
  // There may not be a prodDeployment if the project was hibernated
  if (!projectData.prodDeployment) {
    return [];
  }

  // Hack: in development, the runtime host is actually on port 8081
  const runtimeHost = projectData.prodDeployment.runtimeHost.replace(
    "localhost:9091",
    "localhost:8081"
  );

  const axios = Axios.create({
    baseURL: runtimeHost,
    headers: {
      Authorization: `Bearer ${projectData.jwt}`,
    },
  });

  const catalogEntriesResponse = await axios.get(
    `/v1/instances/${projectData.prodDeployment.runtimeInstanceId}/catalog?type=OBJECT_TYPE_METRICS_VIEW`
  );

  const catalogEntries = catalogEntriesResponse.data
    ?.entries as V1CatalogEntry[];

  const dashboards = catalogEntries?.map(
    (entry: V1CatalogEntry) => entry.metricsView
  );

  return dashboards;
}

export function useDashboards(instanceId: string) {
  return useFilteredResources(instanceId, ResourceKind.MetricsView);
}

export function useDashboardsV2(instanceId: string) {
  return createRuntimeServiceListResources(instanceId, {
    kind: ResourceKind.MetricsView,
  });
}
