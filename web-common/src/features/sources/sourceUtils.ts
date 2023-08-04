import type {
  V1Connector,
  V1ReconcileError,
} from "@rilldata/web-common/runtime-client";
import { behaviourEvent, errorEvent } from "../../metrics/initMetrics";
import type { BehaviourEventMedium } from "../../metrics/service/BehaviourEventTypes";
import type {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "../../metrics/service/MetricsTypes";
import type { SourceConnectionType } from "../../metrics/service/SourceEventTypes";
import { getFilePathFromNameAndType } from "../entity-management/entity-mappers";
import { EntityType } from "../entity-management/types";
import { sanitizeEntityName } from "./extract-table-name";
import { categorizeSourceError } from "./modal/errors";

export function compileCreateSourceYAML(
  values: Record<string, unknown>,
  connectorName: string
) {
  const topLineComment = `# Visit https://docs.rilldata.com/reference/project-files/sources to learn more about Rill source files.`;

  if (
    connectorName !== "local_file" &&
    connectorName !== "motherduck" &&
    connectorName !== "bigquery"
  ) {
    values.uri = values.path;
    delete values.path;
  }

  const compiledKeyValues = Object.entries(values)
    .filter(([key]) => key !== "sourceName")
    .map(([key, value]) => `${key}: "${value}"`)
    .join("\n");

  return `${topLineComment}\n\ntype: "${connectorName}"\n` + compiledKeyValues;
}

export function inferSourceName(connector: V1Connector, path: string) {
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

export function getSourceError(errors: V1ReconcileError[], sourceName) {
  const path = getFilePathFromNameAndType(sourceName, EntityType.Table);

  return errors?.find((error) => error?.filePath === path);
}

export function emitSourceErrorTelemetry(
  space: MetricsEventSpace,
  screenName: MetricsEventScreenName,
  errorMessage: string,
  connectionType: SourceConnectionType,
  fileName: string
) {
  const categorizedError = categorizeSourceError(errorMessage);
  const fileType = getFileTypeFromPath(fileName);
  const isGlob = fileName.includes("*");

  errorEvent?.fireSourceErrorEvent(
    space,
    screenName,
    categorizedError,
    connectionType,
    fileType,
    isGlob
  );
}

export function emitSourceSuccessTelemetry(
  space: MetricsEventSpace,
  screenName: MetricsEventScreenName,
  medium: BehaviourEventMedium,
  connectionType: SourceConnectionType,
  fileName: string
) {
  const fileType = getFileTypeFromPath(fileName);
  const isGlob = fileName.includes("*");

  behaviourEvent?.fireSourceSuccessEvent(
    medium,
    screenName,
    space,
    connectionType,
    fileType,
    isGlob
  );
}
