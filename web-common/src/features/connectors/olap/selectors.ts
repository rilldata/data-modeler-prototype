import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import { TableInfo } from "../../../proto/gen/rill/runtime/v1/connectors_pb";
import {
  type V1TableInfo,
  createConnectorServiceOLAPListTables,
  createRuntimeServiceAnalyzeConnectors,
  createRuntimeServiceGetInstance,
  createRuntimeServiceListConnectorDrivers,
} from "../../../runtime-client";
import { featureFlags } from "../../feature-flags";
import {
  CLICKHOUSE_SOURCE_CONNECTORS,
  DUCKDB_SOURCE_CONNECTORS,
} from "../connector-availability";
import { OLAP_DRIVERS_WITHOUT_MODELING } from "./olap-config";

export function useCurrentOlapConnector(instanceId: string) {
  return createRuntimeServiceGetInstance(
    instanceId,
    { sensitive: true },
    {
      query: {
        select: (data) => {
          const {
            instance: {
              olapConnector: olapConnectorName = "",
              connectors = [],
              projectConnectors = [],
            } = {},
          } = data || {};
          const olapConnector = [...connectors, ...projectConnectors].find(
            (connector) => connector.name === olapConnectorName,
          );
          return olapConnector;
        },
      },
    },
  );
}

export function useSourceConnectorsForCurrentOlapConnector(instanceId: string) {
  return derived(
    [
      createRuntimeServiceListConnectorDrivers(),
      useCurrentOlapConnector(instanceId),
    ],
    ([connectors, olapConnector]) => {
      const allConnectorDrivers = connectors.data?.connectors ?? [];
      const olapConnectorType = olapConnector.data?.type;

      if (!allConnectorDrivers || !olapConnectorType) {
        return [];
      }

      const sourceConnectorNames = (
        olapConnectorType === "clickhouse"
          ? CLICKHOUSE_SOURCE_CONNECTORS
          : olapConnectorType === "duckdb"
            ? DUCKDB_SOURCE_CONNECTORS
            : []
      ) as string[];

      const sourceConnectors = allConnectorDrivers
        .filter((a) => {
          return a.name && sourceConnectorNames.includes(a.name);
        })
        .sort(
          // CAST SAFETY: we have filtered out any connectors that
          // don't have a `name` in the previous filter
          (a, b) =>
            sourceConnectorNames.indexOf(a.name as string) -
            sourceConnectorNames.indexOf(b.name as string),
        );

      return sourceConnectors;
    },
  );
}

export function useIsModelingSupportedForOlapDriver(
  instanceId: string,
  driver: string,
) {
  const { clickhouseModeling } = featureFlags;
  return derived(
    [createRuntimeServiceAnalyzeConnectors(instanceId), clickhouseModeling],
    ([$connectorsQuery, $clickhouseModeling]) => {
      const { connectors = [] } = $connectorsQuery.data || {};
      const olapConnector = connectors.find(
        (connector) => connector.name === driver,
      );
      const olapDriverName = olapConnector?.driver?.name ?? "";

      if (olapDriverName === "clickhouse") {
        return $clickhouseModeling;
      }

      return !OLAP_DRIVERS_WITHOUT_MODELING.includes(olapDriverName);
    },
  );
}

export function useIsModelingSupportedForDefaultOlapDriver(instanceId: string) {
  const { clickhouseModeling } = featureFlags;
  return derived(
    [
      createRuntimeServiceGetInstance(instanceId, { sensitive: true }),
      createRuntimeServiceAnalyzeConnectors(instanceId),
      clickhouseModeling,
    ],
    ([$instanceQuery, $connectorsQuery, $clickhouseModeling]) => {
      const { instance: { olapConnector: olapConnectorName = "" } = {} } =
        $instanceQuery.data || {};
      const { connectors = [] } = $connectorsQuery.data || {};

      const olapConnector = connectors.find(
        (connector) => connector.name === olapConnectorName,
      );

      const olapDriverName = olapConnector?.driver?.name ?? "";

      if (olapDriverName === "clickhouse") {
        return $clickhouseModeling;
      }

      return !OLAP_DRIVERS_WITHOUT_MODELING.includes(olapDriverName);
    },
  );
}

export function useDatabases(instanceId: string, connector: string) {
  return createConnectorServiceOLAPListTables(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          // Get the unique databases
          return (
            data.tables
              ?.map((tableInfo: TableInfo) => tableInfo.database)
              .filter((value, index, self) => self.indexOf(value) === index) ??
            []
          );
        },
      },
    },
  );
}

export function useDatabaseSchemas(
  instanceId: string,
  connector: string,
  database: string,
) {
  return createConnectorServiceOLAPListTables(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          return (
            data.tables
              ?.filter((table) => table.database === database)
              .map((table) => table.databaseSchema)
              .filter((value, index, self) => self.indexOf(value) === index) ??
            []
          );
        },
      },
    },
  );
}

export function useTables(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
): CreateQueryResult<V1TableInfo[]> {
  return createConnectorServiceOLAPListTables(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          return (
            data.tables?.filter(
              (table) =>
                table.database === database &&
                table.databaseSchema === databaseSchema,
            ) ?? []
          );
        },
      },
    },
  );
}
