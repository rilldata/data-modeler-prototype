import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client/index.js";
import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
import { EntityType } from "@rilldata/web-common/features/entity-management/types.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import type { QueryFunction } from "@tanstack/svelte-query";
import { error } from "@sveltejs/kit";

export const ssr = false;

export async function load({ parent, params }) {
  const parentData = await parent();
  const dashboardName = params.name;

  const filePath = getFilePathFromNameAndType(
    dashboardName,
    EntityType.MetricsDefinition,
  );

  const instanceId = parentData.instance?.instanceId ?? "default";

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetFile>>
  > = ({ signal }) => runtimeServiceGetFile(instanceId, filePath, signal);

  const fileQuery = queryClient.fetchQuery({
    queryKey: [filePath],
    queryFn,
  });

  try {
    const file = await fileQuery;

    return {
      file,
    };
  } catch {
    throw error(404, "dashboard not found");
  }
}
