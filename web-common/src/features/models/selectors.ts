import {
  createRuntimeServiceGetCatalogEntry,
  createRuntimeServiceGetFile,
  createRuntimeServiceListFiles,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
  StructTypeField,
  V1ListFilesResponse,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/query-core";
import { TIMESTAMPS } from "../../lib/duckdb-data-types";

export function useModelNames(instanceId: string) {
  return createRuntimeServiceListFiles(
    instanceId,
    {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        // refetchInterval: 1000,
        select: (data) =>
          data.paths
            ?.filter((path) => path.includes("models/"))
            .map((path) => path.replace("/models/", "").replace(".sql", ""))
            // sort alphabetically case-insensitive
            .sort((a, b) =>
              a.localeCompare(b, undefined, { sensitivity: "base" })
            ),
      },
    }
  );
}

export async function getModelNames(
  queryClient: QueryClient,
  instanceId: string
) {
  const files = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    }),
    queryFn: () => {
      return runtimeServiceListFiles(instanceId, {
        glob: "{sources,models,dashboards}/*.{yaml,sql}",
      });
    },
  });
  const modelNames = files.paths
    ?.filter((path) => path.includes("models/"))
    .map((path) => path.replace("/models/", "").replace(".sql", ""))
    // sort alphabetically case-insensitive
    .sort((a, b) => a.localeCompare(b, undefined, { sensitivity: "base" }));
  return modelNames;
}

export function useModelFileIsEmpty(instanceId, modelName) {
  return createRuntimeServiceGetFile(instanceId, `/models/${modelName}.sql`, {
    query: {
      select(data) {
        return data?.blob?.length === 0;
      },
    },
  });
}

export function useModelTimestampColumns(
  instanceId: string,
  modelName: string
) {
  return createRuntimeServiceGetCatalogEntry(instanceId, modelName, {
    query: {
      select: (data) =>
        data?.entry?.model?.schema?.fields?.filter((field: StructTypeField) =>
          TIMESTAMPS.has(field.type.code as string)
        ) ?? [].map((field) => field.name),
    },
  });
}
