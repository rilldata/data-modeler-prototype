import * as yup from "yup";
import { INVALID_NAME_MESSAGE, VALID_NAME_PATTERN } from "@rilldata/utils";

export const getYupSchema = {
  s3: yup.object().shape({
    path: yup
      .string()
      .matches(/^s3:\/\//, "Must be an S3 URI (e.g. s3://bucket/path)")
      .required("S3 URI is required"),
    aws_region: yup.string(),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  gcs: yup.object().shape({
    path: yup
      .string()
      .matches(/^gs:\/\//, "Must be a GS URI (e.g. gs://bucket/path)")
      .required("GS URI is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  https: yup.object().shape({
    path: yup
      .string()
      .matches(/^https?:\/\//, 'Path must start with "http(s)://"')
      .required("Path is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  duckdb: yup.object().shape({
    db: yup.string().required("db is required"),
    sql: yup.string().required("sql is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  motherduck: yup.object().shape({
    dsn: yup.string().required("Connection string is required"),
    sql: yup.string().required("SQL is required"),
    token: yup.string().required("Access token is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  sqlite: yup.object().shape({
    db: yup.string().required("db is required"),
    table: yup.string().required("table is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  bigquery: yup.object().shape({
    sql: yup.string().required("sql is required"),
    project_id: yup.string().required("project_id is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  azure: yup.object().shape({
    path: yup
      .string()
      .matches(
        /^azure:\/\//,
        "Must be an Azure URI (e.g. azure://container/path)",
      )
      .required("Path is required"),
    account: yup.string(),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  postgres: yup.object().shape({
    sql: yup.string().required("sql is required"),
    database_url: yup.string(),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  snowflake: yup.object().shape({
    sql: yup.string().required("sql is required"),
    dsn: yup.string(),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  salesforce: yup.object().shape({
    soql: yup.string().required("soql is required"),
    sobject: yup.string().required("sobject is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  athena: yup.object().shape({
    sql: yup.string().required("sql is required"),
    output_location: yup.string(),
    workgroup: yup.string(),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  redshift: yup.object().shape({
    sql: yup.string().required("SQL is required"),
    database: yup.string().required("database name is required"),
    output_location: yup.string().required("S3 location for temporary files"),
    workgroup: yup.string().optional(),
    cluster_identifier: yup.string().optional(),
    role_arn: yup
      .string()
      .required("Role ARN associated with the Redshift cluster"),
    region: yup.string().optional(),
    name: yup
      .string()
      .matches(
        /^[a-zA-Z_][a-zA-Z0-9_]*$/,
        "Source name must start with a letter or underscore and contain only letters, numbers, and underscores",
      )
      .required("Source name is required"),
  }),

  mysql: yup.object().shape({
    sql: yup.string().required("sql is required"),
    dsn: yup.string(),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  clickhouse: yup.object().shape({
    host: yup
      .string()
      .required("Host is required")
      .matches(
        /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
        "Do not prefix the host with `http(s)://`", // It will be added by the runtime
      ),
    port: yup
      .string() // Purposefully using a string input, not a numeric input
      .matches(/^\d+$/, "Port must be a number")
      .required("Port is required"),
    username: yup.string(),
    password: yup.string(),
    ssl: yup.boolean(),
    name: yup.string(), // Required for typing
    // User-provided connector names requires a little refactor. Commenting out for now.
    // name: yup
    //   .string()
    //   .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
    //   .required("Connector name is required"),
  }),

  druid: yup.object().shape({
    host: yup
      .string()
      .required("Host is required")
      .matches(
        /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
        "Do not prefix the host with `http(s)://`", // It will be added by the runtime
      ),
    port: yup
      .string() // Purposefully using a string input, not a numeric input
      .matches(/^\d+$/, "Port must be a number")
      .required("Port is required"),
    username: yup.string(),
    password: yup.string(),
    ssl: yup.boolean(),
    name: yup.string(), // Required for typing
    // User-provided connector names requires a little refactor. Commenting out for now.
    // name: yup
    //   .string()
    //   .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
    //   .required("Connector name is required"),
  }),

  pinot: yup.object().shape({
    host: yup
      .string()
      .required("Host is required")
      .matches(
        /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
        "Do not prefix the host with `http(s)://`", // It will be added by the runtime
      ),
    port: yup
      .string() // Purposefully using a string input, not a numeric input
      .matches(/^\d+$/, "Port must be a number")
      .required("Port is required"),
    username: yup.string(),
    password: yup.string(),
    ssl: yup.boolean(),
    name: yup.string(), // Required for typing
    // User-provided connector names requires a little refactor. Commenting out for now.
    // name: yup
    //   .string()
    //   .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
    //   .required("Connector name is required"),
  }),
};

export function toYupFriendlyKey(key: string) {
  return key.replace(/\./g, "_");
}

export function fromYupFriendlyKey(key: string) {
  return key.replace(/_/g, ".");
}
