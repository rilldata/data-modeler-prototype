export const OLAP_CONNECTOR_NAMES = ["duckdb", "clickhouse", "druid"];

export type OlapDriver = (typeof OLAP_CONNECTOR_NAMES)[number];

export const OLAP_DRIVERS_WITHOUT_MODELING = ["clickhouse", "druid", "pinot"];

export function makeFullyQualifiedTableName(
  driver: string,
  database: string,
  databaseSchema: string,
  table: string,
) {
  switch (driver) {
    case "clickhouse":
      return `${databaseSchema}.${table}`;
    case "druid":
      return `${databaseSchema}.${table}`;
    case "duckdb":
      return `${database}.${databaseSchema}.${table}`;
    case "pinot":
      return table;
    default:
      throw new Error(`Unsupported OLAP connector: ${driver}`);
  }
}

/**
 * Returns a sufficiently qualified table name for the given OLAP connector.
 * Notably, the table name is *not* qualified with the database and databaseSchema when they are the default values for the given connector.
 */
export function makeSufficientlyQualifiedTableName(
  driver: string,
  database: string,
  databaseSchema: string,
  table: string,
) {
  switch (driver) {
    case "clickhouse":
      if (databaseSchema === "default" || databaseSchema === "") return table;
      return `${databaseSchema}.${table}`;
    case "druid":
      // TODO
      return table;
    case "duckdb":
      if (database === "main_db" || database === "") {
        if (databaseSchema === "main" || databaseSchema === "") return table;
        return `${databaseSchema}.${table}`;
      }
      return `${database}.${databaseSchema}.${table}`;
    case "pinot":
      // TODO
      return table;
    default:
      throw new Error(`Unsupported OLAP connector: ${driver}`);
  }
}

export function makeTablePreviewHref(
  driver: string,
  connectorName: string,
  database: string,
  databaseSchema: string,
  table: string,
): string {
  switch (driver) {
    case "clickhouse":
      return `/connector/clickhouse/${connectorName}/${databaseSchema}/${table}`;
    case "druid":
      return `/connector/druid/${connectorName}/${databaseSchema}/${table}`;
    case "duckdb":
      return `/connector/duckdb/${connectorName}/${database}/${databaseSchema}/${table}`;
    case "pinot":
      return `/connector/pinot/${connectorName}/${table}`;
    default:
      throw new Error(`Unsupported connector: ${driver}`);
  }
}
