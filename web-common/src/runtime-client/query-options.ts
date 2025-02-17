import type { FetchQueryOptions } from "@tanstack/svelte-query";
import type { RuntimeServiceGetResourceParams } from "./gen/index.schemas";
import type { RuntimeServiceGetResourceQueryResult } from "./gen/runtime-service/runtime-service";
import {
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
} from "./gen/runtime-service/runtime-service";

export function getRuntimeServiceGetResourceQueryOptions(
  instanceId: string,
  params: RuntimeServiceGetResourceParams,
) {
  return <FetchQueryOptions<RuntimeServiceGetResourceQueryResult>>{
    queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, params),
    queryFn: () => runtimeServiceGetResource(instanceId, params),
    staleTime: Infinity,
  };
}
