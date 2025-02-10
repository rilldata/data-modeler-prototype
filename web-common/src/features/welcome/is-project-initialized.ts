import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  type V1ListFilesResponse,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceGetInstance,
  runtimeServiceListFiles,
  runtimeServiceUnpackEmpty,
} from "@rilldata/web-common/runtime-client";
import { EMPTY_PROJECT_TITLE } from "./constants";

export async function isProjectInitialized(instanceId: string) {
  try {
    const files = await queryClient.fetchQuery<V1ListFilesResponse>({
      queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
      queryFn: ({ signal }) => {
        return runtimeServiceListFiles(instanceId, undefined, signal);
      },
      // Sometimes, after unpacking an example project, this request fails with a "TypeError: Failed to fetch".
      // So, we retry a handful of times before giving up.
      retry: (failureCount, error) => {
        console.error("RuntimeServiceListFiles", error);
        return failureCount < 10;
      },
    });

    const hasRillYaml = !!files.files?.some(
      ({ path }) => path === "/rill.yaml",
    );

    return hasRillYaml;
  } catch {
    return false;
  }
}

export async function handleUninitializedProject(instanceId: string) {
  // If the project is not initialized, determine what page to route to dependent on the OLAP connector
  const instance = await runtimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  const olapConnector = instance.instance?.olapConnector;

  if (!olapConnector) {
    throw new Error("OLAP connector is not defined");
  }

  // DuckDB-backed projects should head to the Welcome page for user-guided initialization
  if (olapConnector !== "duckdb") {
    // Clickhouse and Druid-backed projects should be initialized immediately
    await runtimeServiceUnpackEmpty(instanceId, {
      displayName: EMPTY_PROJECT_TITLE,
      force: true,
    });

    return true;
  }

  return false;
}
