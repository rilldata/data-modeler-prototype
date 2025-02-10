import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import {
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error } from "@sveltejs/kit";
import type { QueryFunction } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export const load = async ({ params, depends }) => {
  const { instanceId } = get(runtime);

  const dashboardName = params.name;
  depends(dashboardName, "dashboard");

  const queryParams = {
    "name.kind": ResourceKind.Canvas,
    "name.name": dashboardName,
  };

  const queryKey = getRuntimeServiceGetResourceQueryKey(
    instanceId,
    queryParams,
  );

  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetResource>>
  > = ({ signal }) =>
    runtimeServiceGetResource(instanceId, queryParams, signal);

  try {
    const response = await queryClient.fetchQuery({
      queryFn: queryFunction,
      queryKey,
    });

    const dashboard = response.resource?.canvas;

    if (!dashboard || !dashboard.spec) {
      throw error(404, "Dashboard not found");
    }

    return {
      dashboardName,
      dashboard,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Dashboard not found");
  }
};
