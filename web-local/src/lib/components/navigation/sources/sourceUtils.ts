import type { V1Connector } from "@rilldata/web-common/runtime-client";
import { sanitizeEntityName } from "../../../util/extract-table-name";

export function compileCreateSourceYAML(
  values: Record<string, unknown>,
  connectorName: string
) {
  const topLineComment = `# Visit https://docs.rilldata.com/ to learn more about Rill code artifacts.`;

  if (connectorName !== "file") {
    values.uri = values.path;
    delete values.path;
  }

  const compiledKeyValues = Object.entries(values)
    .filter(([key]) => key !== "sourceName")
    .map(([key, value]) => `${key}: "${value}"`)
    .join("\n");

  return `${topLineComment}\n\ntype: "${connectorName}"\n` + compiledKeyValues;
}

export function compileAutogeneratedDashboardYAML(
  dashboardName: string,
  modelName: string
) {
  return `
display_name: "${dashboardName}"
description: "a description that appears in the UI"

# model
#optional to declare this, otherwise it is the model.sql file in the same directory
from: "${modelName}"

# populate with the first datetime type in the OBT
timeseries: ""

# default to opionated option around estimated timegrain,
# first in order is default time grain
timegrains:
  - 1 day
# the timegrain that users will see when they first visit the dashboard.
default_timegrain: "1 day"

# measures
# measures are presented in the order that they are written in this file.
measures: []

# dimensions
# dimensions are presented in the order that they are written in this file.
dimensions: []
`;
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
