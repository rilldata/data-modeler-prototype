import {
  runtimeServiceListCatalogEntries,
  runtimeServiceListFiles,
  runtimeServiceReconcile,
} from "@rilldata/web-common/runtime-client";
import type { V1ReconcileResponse } from "@rilldata/web-common/runtime-client";
import { getMapFromArray } from "@rilldata/web-local/common/utils/arrayUtils";

export async function getArtifactErrors(
  instanceId: string
): Promise<V1ReconcileResponse> {
  try {
    const files = await runtimeServiceListFiles(instanceId, {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    });
    const catalogs = await runtimeServiceListCatalogEntries(instanceId);
    const catalogsMap = getMapFromArray(
      catalogs.entries,
      (catalog) => catalog.path
    );
    const missingFiles = files.paths.filter(
      (filePath) => !catalogsMap.has(filePath)
    );
    return runtimeServiceReconcile(instanceId, {
      changedPaths: missingFiles,
      forcedPaths: missingFiles,
      dry: true,
    });
  } catch (err) {
    if (err.response?.data?.code === 14) {
      // temporary fix when service is not unavailable
      setTimeout(() => {
        window.location.reload();
      }, 200);
    }
    throw err;
  }
}
