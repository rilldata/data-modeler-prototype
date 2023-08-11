import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
import { createModel } from "@rilldata/web-common/features/models/createModel";
import type {
  CreateBaseMutationResult,
  QueryClient,
} from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { notifications } from "../../components/notifications";
import {
  runtimeServicePutFileAndReconcile,
  type V1PutFileAndReconcileResponse,
} from "../../runtime-client";
import { invalidateAfterReconcile } from "../../runtime-client/invalidation";
import { runtime } from "../../runtime-client/runtime-store";
import { getFilePathFromNameAndType } from "../entity-management/entity-mappers";
import { fileArtifactsStore } from "../entity-management/file-artifacts-store";
import { EntityType } from "../entity-management/types";
import { getModelNames } from "../models/selectors";

export async function createModelFromSource(
  queryClient: QueryClient,
  instanceId: string,
  modelNames: Array<string>,
  sourceName: string,
  sourceNameInQuery: string,
  createModelMutation: CreateBaseMutationResult<V1PutFileAndReconcileResponse>, // TODO: type
  setAsActive = true
): Promise<string> {
  const newModelName = getName(`${sourceName}_model`, modelNames);
  await createModel(
    queryClient,
    instanceId,
    newModelName,
    createModelMutation,
    `select * from ${sourceNameInQuery}`,
    setAsActive
  );
  notifications.send({
    message: `Queried ${sourceNameInQuery} in workspace`,
  });
  return newModelName;
}

export async function createModelFromSourceV2(
  queryClient: QueryClient,
  sourceName: string
): Promise<string> {
  const instanceId = get(runtime).instanceId;

  // Get new model name
  const modelNames = await getModelNames(queryClient, instanceId);
  const newModelName = getName(`${sourceName}_model`, modelNames);

  // Create model
  const resp = await runtimeServicePutFileAndReconcile({
    instanceId,
    path: getFilePathFromNameAndType(newModelName, EntityType.Model),
    blob: `select * from ${sourceName}`,
    createOnly: true,
    strict: true,
  });

  // Handle errors
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);

  // Invalidate relevant queries
  invalidateAfterReconcile(queryClient, instanceId, resp);

  // Done
  return newModelName;
}
