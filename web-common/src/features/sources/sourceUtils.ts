import { extractFileExtension } from "@rilldata/web-common/features/entity-management/file-path-utils";
import {
  ConnectorDriverPropertyType,
  type V1ConnectorDriver,
  type V1SourceV2,
} from "@rilldata/web-common/runtime-client";
import { makeDotEnvConnectorKey } from "../connectors/code-utils";
import { sanitizeEntityName } from "../entity-management/name-utils";

// Helper text that we put at the top of every Source YAML file
const TOP_OF_FILE = `# Source YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/sources

type: source`;

export function compileSourceYAML(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
) {
  // Get the secret property keys
  const secretPropertyKeys =
    connector.sourceProperties
      ?.filter((property) => property.secret)
      .map((property) => property.key) || [];

  // Get the string property keys
  const stringPropertyKeys =
    connector.sourceProperties
      ?.filter(
        (property) => property.type === ConnectorDriverPropertyType.TYPE_STRING,
      )
      .map((property) => property.key) || [];

  // Compile key value pairs
  const compiledKeyValues = Object.keys(formValues)
    .filter((key) => formValues[key] !== undefined)
    .filter((key) => key !== "name")
    .map((key) => {
      const value = formValues[key] as string;

      const isSecretProperty = secretPropertyKeys.includes(key);
      if (isSecretProperty) {
        // In Source YAML, explictly referencing `.env` secrets is not yet supported
        // For now, `.env` secrets are implicitly referenced
        return;

        return `${key}: "{{ .env.${makeDotEnvConnectorKey(
          connector.name as string,
          key,
        )} }}"`;
      }

      const isStringProperty = stringPropertyKeys.includes(key);
      if (isStringProperty) {
        return `${key}: "${value}"`;
      }

      return `${key}: ${value}`;
    })
    .join("\n");

  // Return the compiled YAML
  return (
    `${TOP_OF_FILE}\n\nconnector: "${connector.name}"\n` + compiledKeyValues
  );
}

export function compileLocalFileSourceYAML(path: string) {
  return `${TOP_OF_FILE}\n\nconnector: "duckdb"\nsql: "${buildDuckDbQuery(
    path,
  )}"`;
}

function buildDuckDbQuery(path: string): string {
  const extension = extractFileExtension(path);
  if (extensionContainsParts(extension, [".csv", ".tsv", ".txt"])) {
    return `select * from read_csv('${path}', auto_detect=true, ignore_errors=1, header=true)`;
  } else if (extensionContainsParts(extension, [".parquet"])) {
    return `select * from read_parquet('${path}')`;
  } else if (extensionContainsParts(extension, [".json", ".ndjson"])) {
    return `select * from read_json('${path}', auto_detect=true, format='auto')`;
  }

  return `select * from '${path}'`;
}

/**
 * Checks if a file extension '.v1.parquet.gz' contains parts like '.parquet'
 */
function extensionContainsParts(
  fileExtension: string,
  extensionParts: Array<string>,
) {
  for (const extension of extensionParts) {
    if (fileExtension.includes(extension)) return true;
  }
  return false;
}

export function inferSourceName(connector: V1ConnectorDriver, path: string) {
  if (
    !path ||
    path.endsWith("/") ||
    (connector.name === "gcs" && !path.startsWith("gs://")) ||
    (connector.name === "s3" && !path.startsWith("s3://")) ||
    (connector.name === "https" &&
      !path.startsWith("https://") &&
      !path.startsWith("http://"))
  )
    return;

  const slug = path
    .split("/")
    .filter((s: string) => s.length > 0)
    .pop();

  if (!slug) return;

  const fileName = slug.split(".").shift();

  if (!fileName) return;

  return sanitizeEntityName(fileName);
}

export function getFileTypeFromPath(fileName) {
  if (!fileName.includes(".")) return "";
  const fileType = fileName.split(/[#?]/)[0].split(".").pop();

  if (!fileType) return "";

  if (fileType === "gz") {
    return fileName.split(".").slice(-2).shift();
  }

  return fileType;
}

/**
 * Convert applicable connectors to DuckDB. We do this to leverage DuckDB's native,
 * well-documented file reading capabilities.
 */
export function maybeRewriteToDuckDb(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
): [V1ConnectorDriver, Record<string, unknown>] {
  // Create a copy of the connector, so that we don't overwrite the original
  const connectorCopy = { ...connector };

  switch (connector.name) {
    case "s3":
    case "gcs":
    case "https":
    case "azure":
    case "local_file":
      connectorCopy.name = "duckdb";

      formValues.sql = buildDuckDbQuery(formValues.path as string);
      delete formValues.path;

      break;
    case "sqlite":
      connectorCopy.name = "duckdb";

      formValues.sql = `SELECT * FROM sqlite_scan('${formValues.db as string}', '${
        formValues.table as string
      }');`;
      delete formValues.db;
      delete formValues.table;

      break;

    default:
      return [connector, formValues];
  }

  connectorCopy.sourceProperties = [
    {
      key: "sql",
      type: ConnectorDriverPropertyType.TYPE_STRING,
    },
  ];

  return [connectorCopy, formValues];
}

export function getFileExtension(source: V1SourceV2): string {
  const path = source?.spec?.properties?.path?.toLowerCase();
  if (path?.includes(".csv")) return "CSV";
  if (path?.includes(".parquet")) return "Parquet";
  if (path?.includes(".json")) return "JSON";
  if (path?.includes(".ndjson")) return "JSON";
  return "";
}

export function formatConnectorType(source: V1SourceV2) {
  switch (source?.spec?.sourceConnector) {
    case "s3":
      return "S3";
    case "gcs":
      return "GCS";
    case "https":
      return "http(s)";
    case "local_file":
      return "Local file";
    default:
      return source?.state?.connector ?? "";
  }
}
