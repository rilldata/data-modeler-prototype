import { goto } from "$app/navigation";
import {
  RpcStatus,
  runtimeServiceGetCatalogEntry,
  runtimeServicePutFileAndReconcile,
  useRuntimeServiceListFiles,
  V1DeleteFileAndReconcileResponse,
  V1ReconcileError,
  V1RenameFileAndReconcileResponse,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import type { ActiveEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getNextEntityName } from "@rilldata/web-local/common/utils/getNextEntityId";
import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
import { notifications } from "@rilldata/web-local/lib/components/notifications";
import { invalidateAfterReconcile } from "@rilldata/web-local/lib/svelte-query/invalidation";
import {
  getFileFromName,
  getLabel,
  getNameFromFile,
  getRouteFromName,
} from "@rilldata/web-local/lib/util/entity-mappers";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
import {
  MutationFunction,
  useMutation,
  UseMutationOptions,
} from "@sveltestack/svelte-query";
import { generateMeasuresAndDimension } from "../application-state-stores/metrics-internal-store";

export function useAllNames(instanceId: string) {
  return useRuntimeServiceListFiles(
    instanceId,
    {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        select: (data) =>
          data.paths?.map((path) => getNameFromFile(path)) ?? [],
      },
    }
  );
}

export function isDuplicateName(name: string, names: Array<string>) {
  return names.findIndex((n) => n.toLowerCase() === name.toLowerCase()) >= 0;
}

export async function renameFileArtifact(
  queryClient: QueryClient,
  instanceId: string,
  fromName: string,
  toName: string,
  type: EntityType,
  renameMutation: UseMutationResult<V1RenameFileAndReconcileResponse>
) {
  const resp = await renameMutation.mutateAsync({
    data: {
      instanceId,
      fromPath: getFileFromName(fromName, type),
      toPath: getFileFromName(toName, type),
    },
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

  httpRequestQueue.removeByName(fromName);
  notifications.send({
    message: `Renamed ${getLabel(type)} ${fromName} to ${toName}`,
  });

  invalidateAfterReconcile(queryClient, instanceId, resp);
  goto(getRouteFromName(toName, type), {
    replaceState: true,
  });
}

export async function deleteFileArtifact(
  queryClient: QueryClient,
  instanceId: string,
  name: string,
  type: EntityType,
  deleteMutation: UseMutationResult<V1DeleteFileAndReconcileResponse>,
  activeEntity: ActiveEntity,
  names: Array<string>,
  showNotification = true
) {
  try {
    const resp = await deleteMutation.mutateAsync({
      data: {
        instanceId,
        path: getFileFromName(name, type),
      },
    });
    fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

    httpRequestQueue.removeByName(name);
    if (showNotification) {
      notifications.send({ message: `Deleted ${getLabel(type)} ${name}` });
    }

    invalidateAfterReconcile(queryClient, instanceId, resp);
    if (activeEntity?.name === name) {
      goto(getRouteFromName(getNextEntityName(names, name), type));
    }
  } catch (err) {
    console.error(err);
  }
}

export interface CreateDashboardFromSourceRequest {
  instanceId: string;
  sourceName: string;
  newModelName: string;
  newDashboardName: string;
}

export interface CreateDashboardFromSourceResponse {
  affectedPaths?: string[];
  errors?: V1ReconcileError[];
}

export const useCreateDashboardFromSource = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<Promise<CreateDashboardFromSourceResponse>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<Promise<CreateDashboardFromSourceResponse>>,
    { data: CreateDashboardFromSourceRequest }
  > = async (props) => {
    const { data } = props ?? {};

    // first, create model from source

    await runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
      path: `models/${data.newModelName}.sql`,
      blob: `select * from ${data.sourceName}`,
    });

    // second, create dashboard from model

    const model = await runtimeServiceGetCatalogEntry(
      data.instanceId,
      data.newModelName
    );
    const generatedYAML = generateMeasuresAndDimension(model.entry.model, {
      display_name: `${data.sourceName} dashboard`,
    });

    const response = await runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
      path: getFileFromName(
        data.newDashboardName,
        EntityType.MetricsDefinition
      ),
      blob: generatedYAML,
      create: true,
      createOnly: true,
      strict: false,
    });

    return {
      affectedPaths: response?.affectedPaths,
      errors: response?.errors,
    };
  };

  return useMutation<
    Awaited<Promise<CreateDashboardFromSourceResponse>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >(mutationFn, mutationOptions);
};
