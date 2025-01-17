export const DUCKDB_SOURCE_CONNECTORS = [
  "duckdb",
  "motherduck",
  "postgres",
  "gcs",
  "bigquery",
  "snowflake",
  "s3",
  "athena",
  "redshift",
  "mysql",
  "sqlite",
  "azure",
  "salesforce",
  "local_file",
  "https",
];

export const CLICKHOUSE_SOURCE_CONNECTORS = [
  "postgres",
  "mysql",
  "gcs",
  "s3",
  "azure",
  // "local_file",
] as const;

export const OLAP_CONNECTORS = ["clickhouse", "druid", "pinot"];
