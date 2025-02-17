import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { isHTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper.js";
import { getRuntimeServiceGetResourceQueryOptions } from "@rilldata/web-common/runtime-client/query-options";
import { error } from "@sveltejs/kit";

export async function load({ params, parent }) {
  const { runtime } = await parent();

  const reportData = await queryClient
    .fetchQuery(
      getRuntimeServiceGetResourceQueryOptions(runtime.instanceId, {
        "name.kind": ResourceKind.Report,
        "name.name": params.report,
      }),
    )
    .catch((e) => {
      if (!isHTTPError(e)) {
        throw error(500, "Error fetching report");
      }
      throw error(e.response.status, e.response.data.message);
    });

  return {
    report: reportData.resource,
  };
}
