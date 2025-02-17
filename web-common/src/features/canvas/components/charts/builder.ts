import type { ChartDataResult } from "@rilldata/web-common/features/canvas/components/charts/selector";
import type {
  ChartConfig,
  TooltipValue,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import { sanitizeValueForVega } from "@rilldata/web-common/features/templates/charts/utils";
import type { VisualizationSpec } from "svelte-vega";
import type {
  ColorDef,
  Field,
  PositionDef,
} from "vega-lite/build/src/channeldef";
import type { Encoding } from "vega-lite/build/src/encoding";
import type { TopLevelParameter } from "vega-lite/build/src/spec/toplevel";
import type { TopLevelUnitSpec } from "vega-lite/build/src/spec/unit";

export function createMultiLayerBaseSpec() {
  const baseSpec: VisualizationSpec = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    width: "container",
    data: { name: "metrics-view" },
    autosize: { type: "fit" },
    layer: [],
  };
  return baseSpec;
}

export function createSingleLayerBaseSpec(
  mark: "line" | "bar" | "point",
): TopLevelUnitSpec<Field> {
  return {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    description: `A ${mark} chart with embedded data.`,
    mark,
    width: "container",
    data: { name: "metrics-view" },
    autosize: { type: "fit" },
  };
}

export function createXEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): PositionDef<Field> {
  if (!config.x) return {};
  const metaData = data.fields[config.x.field];
  return {
    field: sanitizeValueForVega(config.x.field),
    title: metaData?.displayName || config.x.field,
    type: config.x.type,
    ...(metaData && "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
    ...(config.x.sort && { sort: config.x.sort }),
    axis: {
      ...(config.x.type === "quantitative" && {
        formatType: config.x.field,
      }),
      ...(metaData && "format" in metaData && { format: metaData.format }),
      ...(!config.x.showAxisTitle && { title: null }),
    },
  };
}

export function createYEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): PositionDef<Field> {
  if (!config.y) return {};
  const metaData = data.fields[config.y.field];
  return {
    field: sanitizeValueForVega(config.y.field),
    title: metaData?.displayName || config.y.field,
    type: config.y.type,
    ...(config.y.zeroBasedOrigin !== true && {
      scale: {
        zero: false,
      },
    }),
    axis: {
      ...(config.y.type === "quantitative" && {
        formatType: config.y.field,
      }),
      ...(!config.y.showAxisTitle && { title: null }),
      ...(metaData && "format" in metaData && { format: metaData.format }),
    },
    ...(metaData && "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
  };
}

export function createColorEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): ColorDef<Field> {
  if (!config.color) return {};
  if (typeof config.color === "object") {
    const metaData = data.fields[config.color.field];

    return {
      field: sanitizeValueForVega(config.color.field),
      title: metaData?.displayName || config.color.field,
      type: config.color.type,
      ...(metaData &&
        "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
    };
  }
  if (typeof config.color === "string") {
    return { value: config.color };
  }
  return {};
}

export function createOpacityEncoding(paramName: string) {
  return {
    condition: [
      { param: paramName, empty: false, value: 1 },
      {
        test: `length(data('${paramName}_store')) == 0`,
        value: 0.8,
      },
    ],
    value: 0.2,
  };
}

export function createLegendParam(
  paramName: string,
  field: string,
): TopLevelParameter {
  return {
    name: paramName,
    select: {
      type: "point",
      fields: [sanitizeValueForVega(field)],
    },
    bind: "legend",
  };
}

export function createDefaultTooltipEncoding(
  config: ChartConfig,
  data: ChartDataResult,
) {
  const tooltip: TooltipValue[] = [];

  if (config.x) {
    tooltip.push({
      field: sanitizeValueForVega(config.x.field),
      title: data.fields[config.x.field]?.displayName || config.x.field,
      type: config.x.type,
      ...(config.x.type === "quantitative" && {
        formatType: config.x.field,
      }),
      ...(config.x.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
    });
  }
  if (config.y) {
    tooltip.push({
      field: sanitizeValueForVega(config.y.field),
      title: data.fields[config.y.field]?.displayName || config.y.field,
      type: config.y.type,
      ...(config.y.type === "quantitative" && {
        formatType: config.y.field,
      }),
      ...(config.y.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
    });
  }
  if (typeof config.color === "object" && config.color.field) {
    tooltip.push({
      field: sanitizeValueForVega(config.color.field),
      title: data.fields[config.color.field]?.displayName || config.color.field,
      type: config.color.type,
    });
  }

  return tooltip;
}

export function createEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): Encoding<Field> {
  return {
    x: createXEncoding(config, data),
    y: createYEncoding(config, data),
    color: createColorEncoding(config, data),
    tooltip: createDefaultTooltipEncoding(config, data),
  };
}
