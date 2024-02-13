import type { QueryClient } from "@tanstack/query-core";
import { get } from "svelte/store";
import {
  V1StructType,
  connectorServiceOLAPGetTable,
  runtimeServiceGetInstance,
  runtimeServicePutFile,
} from "../../runtime-client";
import { runtime } from "../../runtime-client/runtime-store";
import { useDashboardFileNames } from "../dashboards/selectors";
import {
  getFileAPIPathFromNameAndType,
  getFilePathFromNameAndType,
} from "../entity-management/entity-mappers";
import { getName } from "../entity-management/name-utils";
import { waitForResource } from "../entity-management/resource-status-utils";
import { EntityType } from "../entity-management/types";
import { generateDashboardYAMLForTable } from "../metrics-views/metrics-internal-store";

export async function createDashboardFromExternalTable(
  queryClient: QueryClient,
  table: string,
): Promise<string> {
  const instanceId = get(runtime).instanceId;

  // Get the OLAP connector
  const instance = await runtimeServiceGetInstance(instanceId);
  const olapConnector = instance.instance?.olapConnector;

  // Get the schema of the table
  const tableSchema = await connectorServiceOLAPGetTable({
    instanceId: instanceId,
    connector: olapConnector,
    table: table,
  });

  // Get a unique name for the new dashboard
  const dashboardNamesStore = useDashboardFileNames(instanceId);
  const dashboardNames = get(dashboardNamesStore).data;
  if (!dashboardNames) {
    throw new Error("Could not get dashboard names");
  }
  const newDashboardName = getName(`${table}_dashboard`, dashboardNames);

  // Create the dashboard
  const dashboardYAML = generateDashboardYAMLForTable(
    table,
    tableSchema.schema as V1StructType,
    newDashboardName,
  );
  await runtimeServicePutFile(
    instanceId,
    getFileAPIPathFromNameAndType(
      newDashboardName,
      EntityType.MetricsDefinition,
    ),
    {
      blob: dashboardYAML,
      create: true,
      createOnly: true,
    },
  );
  await waitForResource(
    queryClient,
    instanceId,
    getFilePathFromNameAndType(newDashboardName, EntityType.MetricsDefinition),
  );

  return newDashboardName;
}
