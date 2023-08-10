import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  RpcStatus,
  runtimeServiceGetCatalogEntry,
  runtimeServicePutFileAndReconcile,
  V1ReconcileError,
} from "@rilldata/web-common/runtime-client";
import {
  createMutation,
  CreateMutationOptions,
  MutationFunction,
} from "@tanstack/svelte-query";
import { generateDashboardYAMLForModel } from "../metrics-views/metrics-internal-store";

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
  mutation?: CreateMutationOptions<
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
      path: getFilePathFromNameAndType(data.newModelName, EntityType.Model),
      blob: `select * from ${data.sourceName}`,
    });

    // second, create dashboard from model

    const model = await runtimeServiceGetCatalogEntry(
      data.instanceId,
      data.newModelName
    );
    const dashboardYAML = generateDashboardYAMLForModel(
      model.entry.model,
      data.newDashboardName
    );

    const response = await runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
      path: getFilePathFromNameAndType(
        data.newDashboardName,
        EntityType.MetricsDefinition
      ),
      blob: dashboardYAML,
      create: true,
      createOnly: true,
      strict: false,
    });

    return {
      affectedPaths: response?.affectedPaths,
      errors: response?.errors,
    };
  };

  return createMutation<
    Awaited<Promise<CreateDashboardFromSourceResponse>>,
    TError,
    { data: CreateDashboardFromSourceRequest },
    TContext
  >(mutationFn, mutationOptions);
};
