import { V1DeploymentStatus } from "@rilldata/web-admin/client";
import type { V1GetProjectResponse } from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListCatalogEntries,
  createRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
import Axios from "axios";
import { derived, Readable } from "svelte/store";

export interface DashboardListItem {
  name: string;
  title?: string;
  isValid: boolean;
}

export async function getDashboardsForProject(
  projectData: V1GetProjectResponse
): Promise<DashboardListItem[]> {
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

  // get all valid and invalid dashboards
  const filesRequest = axios.get(
    `/v1/instances/${projectData.prodDeployment.runtimeInstanceId}/files?glob=dashboards/*.yaml`
  );

  // get the valid dashboards
  const catalogEntriesRequest = axios.get(
    `/v1/instances/${projectData.prodDeployment.runtimeInstanceId}/catalog?type=OBJECT_TYPE_METRICS_VIEW`
  );

  const [filesResponse, catalogEntriesResponse] = await Promise.all([
    filesRequest,
    catalogEntriesRequest,
  ]);

  const filePaths = filesResponse.data?.paths;
  const catalogEntries = catalogEntriesResponse.data?.entries;

  // compose the dashboard list items
  const dashboardListItems = getDashboardListItemsFromFilesAndCatalogEntries(
    filePaths,
    catalogEntries
  );

  return dashboardListItems;
}

export function getDashboardListItemsFromFilesAndCatalogEntries(
  filePaths: string[],
  catalogEntries: V1CatalogEntry[]
): DashboardListItem[] {
  const dashboardListings = filePaths?.map((path: string) => {
    const name = path.replace("/dashboards/", "").replace(".yaml", "");
    const catalogEntry = catalogEntries?.find(
      (entry: V1CatalogEntry) => entry.path === path
    );
    const title = catalogEntry?.metricsView?.label;
    // invalid dashboards are not in the catalog
    const isValid = !!catalogEntry;
    return {
      name,
      title,
      isValid,
    };
  });

  return dashboardListings;
}

export function useDashboardListItems(
  instanceId: string,
  projectStatus: V1DeploymentStatus
): Readable<{
  items: DashboardListItem[];
  isSuccess: boolean;
}> {
  const hasProjectStatus = !!projectStatus;
  const isProfiling =
    hasProjectStatus &&
    (projectStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
      projectStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING);

  return derived(
    [
      createRuntimeServiceListFiles(
        instanceId,
        {
          glob: "dashboards/*.yaml",
        },
        {
          query: {
            placeholderData: undefined,
            enabled: !isProfiling && hasProjectStatus && !!instanceId,
          },
        }
      ),
      createRuntimeServiceListCatalogEntries(
        instanceId,
        {
          type: "OBJECT_TYPE_METRICS_VIEW",
        },
        {
          query: {
            placeholderData: undefined,
            enabled: !isProfiling && hasProjectStatus && !!instanceId,
          },
        }
      ),
    ],
    ([dashboardFiles, dashboardCatalogEntries]) => {
      if (!dashboardFiles.isSuccess || !dashboardCatalogEntries.isSuccess)
        return {
          isSuccess: false,
          items: [],
        };

      return {
        isSuccess: true,
        items: getDashboardListItemsFromFilesAndCatalogEntries(
          dashboardFiles?.data?.paths ?? [],
          dashboardCatalogEntries?.data?.entries ?? []
        ),
      };
    }
  );
}
