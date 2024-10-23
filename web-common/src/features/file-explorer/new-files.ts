import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { getName } from "@rilldata/utils";
import {
  ResourceKind,
  type UserFacingResourceKinds,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  runtimeServicePutFile,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

export async function createResourceFile(
  kind: ResourceKind,
  baseResource?: V1Resource,
): Promise<string> {
  if (!(kind in ResourceKindMap)) {
    throw new Error(`Unknown resource kind: ${kind}`);
  }

  const newPath = getPathForNewResourceFile(kind, baseResource);
  const instanceId = get(runtime).instanceId;

  await runtimeServicePutFile(instanceId, {
    path: newPath,
    blob: generateBlobForNewResourceFile(kind, baseResource),
    create: true,
    createOnly: true,
  });

  // Return the path to the new file, so we can navigate the user to it
  return `/${newPath}`;
}

export function getPathForNewResourceFile(
  newKind: ResourceKind,
  baseResource?: V1Resource,
) {
  const allNames =
    newKind === ResourceKind.Source || newKind === ResourceKind.Model
      ? // sources and models share the name
        [
          ...fileArtifacts.getNamesForKind(ResourceKind.Source),
          ...fileArtifacts.getNamesForKind(ResourceKind.Model),
        ]
      : fileArtifacts.getNamesForKind(newKind);

  const { folderName, extension } = ResourceKindMap[newKind];
  const baseName = getBaseNameForNewResourceFile(newKind, baseResource);
  const newName = getName(baseName, allNames);

  return `${folderName}/${newName}${extension}`;
}

const ResourceKindMap: Record<
  UserFacingResourceKinds,
  {
    folderName: string;
    baseName: string;
    extension: string;
  }
> = {
  [ResourceKind.Source]: {
    folderName: "sources",
    baseName: "source",
    extension: ".yaml",
  },
  [ResourceKind.Connector]: {
    folderName: "connectors",
    baseName: "connector",
    extension: ".yaml",
  },
  [ResourceKind.Model]: {
    folderName: "models",
    baseName: "model",
    extension: ".sql",
  },
  [ResourceKind.MetricsView]: {
    folderName: "metrics",
    baseName: "metrics_view",
    extension: ".yaml",
  },
  [ResourceKind.Explore]: {
    folderName: "dashboards",
    baseName: "explore",
    extension: ".yaml",
  },
  [ResourceKind.API]: {
    folderName: "apis",
    baseName: "api",
    extension: ".yaml",
  },
  [ResourceKind.Component]: {
    folderName: "components",
    baseName: "component",
    extension: ".yaml",
  },
  [ResourceKind.Canvas]: {
    folderName: "dashboards",
    baseName: "canvas",
    extension: ".yaml",
  },
  [ResourceKind.Theme]: {
    folderName: "themes",
    baseName: "theme",
    extension: ".yaml",
  },
  [ResourceKind.Report]: {
    folderName: "reports",
    baseName: "report",
    extension: ".yaml",
  },
  [ResourceKind.Alert]: {
    folderName: "alerts",
    baseName: "alert",
    extension: ".yaml",
  },
};

export function getBaseNameForNewResourceFile(
  newKind: ResourceKind,
  baseResource?: V1Resource,
) {
  switch (newKind) {
    case ResourceKind.Explore:
      return baseResource
        ? `${baseResource.meta!.name!.name}_explore`
        : ResourceKindMap[newKind].baseName;
    default:
      return ResourceKindMap[newKind].baseName;
  }
}

export function generateBlobForNewResourceFile(
  kind: ResourceKind,
  baseResource?: V1Resource,
) {
  switch (kind) {
    case ResourceKind.Connector:
      return ""; // This is constructed in the `features/connectors` directory
    case ResourceKind.Source:
      return ""; // This is constructed in the `features/sources/modal` directory
    case ResourceKind.Model:
      return `-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models

SELECT 'Hello, World!' AS Greeting`;
    case ResourceKind.MetricsView:
      return `# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
type: metrics_view

model: # Choose a model to underpin your metrics
timeseries: # Choose a timestamp column (if any) from your model

dimensions:
measures:
`;
    case ResourceKind.Explore:
      if (baseResource) {
        const metricsViewName = baseResource.meta!.name!.name;
        const metricsViewTitle =
          baseResource.metricsView?.state?.validSpec?.title;

        return `# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards

type: explore

title: "${metricsViewTitle ? metricsViewTitle : metricsViewName} dashboard"
metrics_view: ${metricsViewName}

dimensions: '*'
measures: '*'
`;
      }
      return `# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards

type: explore

title: "My metrics dashboard"
metrics_view: example_metrics_view # Choose a metrics view to underpin the dashboard

dimensions: '*'
measures: '*'
`;
    case ResourceKind.API:
      return `# API YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/apis
# Test your API endpoint at http://localhost:9009/v1/instances/default/api/<filename>

type: api

metrics_sql: |
  select measure, dimension from metrics_view
`;
    case ResourceKind.Component:
      return `# Component YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/components
    
type: component

data:
  sql: |
    SELECT * FROM (VALUES 
      ('Monday', 300),
      ('Tuesday', 150),
      ('Wednesday', 200),
      ('Thursday', 400),
      ('Friday', 650),
      ('Saturday', 575),
      ('Sunday', 500)
    ) AS t(day_of_week, revenue)

vega_lite: |
  {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "data": { "name": "table" },
    "mark": "line",
    "width": "container",
    "encoding": {
      "x": {
        "field": "day_of_week",
        "type": "ordinal",
        "axis": { "title": "Day of the Week" },
        "sort": [
          "Monday",
          "Tuesday",
          "Wednesday",
          "Thursday",
          "Friday",
          "Saturday",
          "Sunday"
        ]
      },
      "y": {
        "field": "revenue",
        "type": "quantitative",
        "axis": { "title": "Revenue" }
      }
    }
  }`;
    case ResourceKind.Canvas:
      return `type: canvas
title: "Canvas Dashboard"
columns: 24
gap: 2

items:
  - component:
      markdown:
        content: "First Component"
        css:
          font-size: "40px"
          background-color: "#fff"
    width: 4
    height: 3
    x: 2
    y: 1
`;
    case ResourceKind.Theme:
      return `# Theme YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/themes

type: theme

colors:
  primary: plum
  secondary: violet 
`;
    case ResourceKind.Alert:
      return `# Alert YAML
# Reference documentation: TODO

type: alert

...
`;
    case ResourceKind.Report:
      return `# Report YAML
# Reference documentation: TODO

type: report

...
`;
    default:
      throw new Error(`Unknown resource kind: ${kind}`);
  }
}
