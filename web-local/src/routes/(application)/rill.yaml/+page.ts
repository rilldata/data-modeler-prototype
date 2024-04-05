import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import type { QueryFunction } from "@tanstack/svelte-query";
const instanceId = "default";
const path = "rill.yaml";

export async function load({ depends }) {
  depends("rill.yaml");
  const queryKey = getRuntimeServiceGetFileQueryKey(instanceId, path);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetFile>>
  > = ({ signal }) => runtimeServiceGetFile(instanceId, path, signal);

  const query = queryClient.fetchQuery({
    queryKey,
    queryFn,
  });

  const data = await query;

  console.log(data);

  return data;
}
