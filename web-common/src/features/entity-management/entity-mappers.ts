import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";

// Temporary solution for the issue with leading `/` for these files.
// TODO: find a solution that works across backend and frontend
export function getFileAPIPathFromNameAndType(
  name: string,
  type: EntityType,
): string {
  switch (type) {
    case EntityType.Connector:
      return `connectors/${name}.yaml`;
    case EntityType.Source:
    case EntityType.Table:
      return `sources/${name}.yaml`;
    case EntityType.Model:
      return `models/${name}.sql`;
    case EntityType.MetricsDefinition:
      return `dashboards/${name}.yaml`;
    case EntityType.Chart:
      return `charts/${name}.yaml`;
    case EntityType.Canvas:
      return `canvas/${name}.yaml`;
    default:
      throw new Error("Unrecognized EntityType");
  }
}

export function getNameFromFile(fileName: string): string {
  // TODO: do we need a library here?
  const splits = fileName.split("/");
  const extensionSplits = splits[splits.length - 1]?.split(".");
  return extensionSplits[0];
}

// Remove a leading slash, if it exists
// In certain backend APIs where path is part of the url.
// Leading slash leads to issues where the end point redirects to one without the slash.
export function removeLeadingSlash(path: string): string {
  return path.replace(/^\//, "");
}

// Add a leading slash if it doesn't exist.
// Temporary, we should eventually make sure this is added in all places
export function addLeadingSlash(path: string): string {
  if (path.startsWith("/")) return path;
  return "/" + path;
}

export const FolderNameToResourceKind: Record<string, ResourceKind> = {
  sources: ResourceKind.Source,
  models: ResourceKind.Model,
  // dashboards: ResourceKind.MetricsView,
  metrics: ResourceKind.MetricsView,
  // explores: ResourceKind.Explore, // This does not happen on backend
  charts: ResourceKind.Component,
  canvas: ResourceKind.Canvas,
};

export const ResourceShortNameToResourceKind: Record<string, ResourceKind> = {
  alert: ResourceKind.Alert,
  api: ResourceKind.API,
  component: ResourceKind.Component,
  canvas: ResourceKind.Canvas,
  metrics_view: ResourceKind.MetricsView,
  metricsview: ResourceKind.MetricsView,
  explore: ResourceKind.Explore,
  model: ResourceKind.Model,
  report: ResourceKind.Report,
  source: ResourceKind.Source,
  theme: ResourceKind.Theme,
};
