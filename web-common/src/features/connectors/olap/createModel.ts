import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { eventBus } from "@rilldata/events";
import { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { hasSpaces, getName } from "@rilldata/utils";
import {
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  getRuntimeServiceGetInstanceQueryKey,
  runtimeServiceAnalyzeConnectors,
  runtimeServiceGetInstance,
  runtimeServicePutFile,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import { makeSufficientlyQualifiedTableName } from "./olap-config";

export async function createModelFromTable(
  queryClient: QueryClient,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
  addDevLimit: boolean = true,
): Promise<[string, string]> {
  const instanceId = get(runtime).instanceId;

  // Get driver name
  const analyzeConnectorsQueryKey =
    getRuntimeServiceAnalyzeConnectorsQueryKey(instanceId);
  const analyzeConnectorsQueryFn = async () =>
    runtimeServiceAnalyzeConnectors(instanceId);
  const connectors = await queryClient.fetchQuery({
    queryKey: analyzeConnectorsQueryKey,
    queryFn: analyzeConnectorsQueryFn,
  });
  const analyzedConnector = connectors?.connectors?.find(
    (c) => c.name === connector,
  );
  if (!analyzedConnector) {
    throw new Error(`Could not find connector ${connector}`);
  }
  const driverName = analyzedConnector.driver?.name as string;

  // Determine whether the connector is the default OLAP connector
  const runtimeInstanceQueryKey =
    getRuntimeServiceGetInstanceQueryKey(instanceId);
  const runtimeInstanceQueryFn = async () =>
    runtimeServiceGetInstance(instanceId, { sensitive: true });
  const runtimeInstance = await queryClient.fetchQuery({
    queryKey: runtimeInstanceQueryKey,
    queryFn: runtimeInstanceQueryFn,
  });
  if (!runtimeInstance) {
    throw new Error(`Could not find runtime instance ${instanceId}`);
  }
  const isDefaultOLAPConnector =
    runtimeInstance?.instance?.olapConnector === connector;

  // Get new model name
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];
  const newModelName = getName(`${table}_model`, allNames);
  const newModelPath = `models/${newModelName}.sql`;

  // Get sufficiently qualified table name
  const sufficientlyQualifiedTableName = makeSufficientlyQualifiedTableName(
    driverName,
    database,
    databaseSchema,
    table,
  );

  // Create model
  const topComments =
    "-- Model SQL\n-- Reference documentation: https://docs.rilldata.com/reference/project-files/models";
  const connectorLine = `-- @connector: ${connector}`;
  const selectStatement = hasSpaces(sufficientlyQualifiedTableName)
    ? `select * from "${sufficientlyQualifiedTableName}"`
    : `select * from ${sufficientlyQualifiedTableName}`;
  const devLimit = "{{ if dev }} limit 100000 {{ end}}";

  let modelSQL = `${topComments}\n`;

  if (!isDefaultOLAPConnector) {
    modelSQL += `${connectorLine}\n`;
  }

  modelSQL += `\n${selectStatement}`;

  if (addDevLimit) {
    modelSQL += `\n${devLimit}`;
  }

  await runtimeServicePutFile(instanceId, {
    path: newModelPath,
    blob: modelSQL,
    createOnly: true,
  });

  eventBus.emit("notification", {
    message: `Queried ${table} in workspace`,
  });

  // Done
  return ["/" + newModelPath, newModelName];
}
