import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  runtimeServiceListResources,
  runtimeServiceGetInstance,
} from "@rilldata/web-common/runtime-client";
import type { QueryFunction } from "@tanstack/svelte-query";
import {
  getRuntimeServiceListResourcesQueryKey,
  getRuntimeServiceListInstancesQueryKey,
} from "@rilldata/web-common/runtime-client";

export const load = async () => {
  const instanceId = "default";

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceListResources>>
  > = ({ signal }) => runtimeServiceListResources(instanceId, {}, signal);

  const instanceQuery: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetInstance>>
  > = ({ signal }) => runtimeServiceGetInstance(instanceId, signal);

  const instanceData = await queryClient.fetchQuery({
    queryFn: instanceQuery,
    queryKey: getRuntimeServiceListInstancesQueryKey(),
  });

  return {
    ...(await queryClient.fetchQuery({
      queryFn,
      queryKey: getRuntimeServiceListResourcesQueryKey(instanceId),
    })),
    instance: {
      ...instanceData.instance,
      instanceId: instanceData.instance?.instanceId ?? "default",
    },
  };
};
