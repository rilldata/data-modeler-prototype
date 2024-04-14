import { buildVegaLiteSpec } from "@rilldata/web-common/features/charts/templates/build-template";
import { TDDChartMap } from "@rilldata/web-common/features/charts/types";
import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { VisualizationSpec } from "svelte-vega";
import { TDDChart, TDDCustomCharts } from "../types";

export function reduceDimensionData(dimensionData: DimensionDataItem[]) {
  return dimensionData
    .map((dimension) =>
      dimension.data.map((datum) => ({
        dimension: dimension.value,
        ...datum,
      })),
    )
    .flat();
}

export function getVegaSpec(
  chartType: TDDCustomCharts,
  expandedMeasureName: string,
  isDimensional: boolean,
): VisualizationSpec {
  const temporalFields = ["ts"];
  const measureFields = [expandedMeasureName];

  const builderChartType = TDDChartMap[chartType];

  const spec = buildVegaLiteSpec(
    builderChartType,
    temporalFields,
    measureFields,
    isDimensional ? ["dimension"] : [],
  );

  return spec;
}

export function sanitizeSpecForTDD(
  spec,
  timeGrain: V1TimeGrain,
  xMin: Date,
  xMax: Date,
  chartType: TDDCustomCharts,
): VisualizationSpec {
  if (!spec) return spec;

  /**
   * Sub level types are not being exported from the vega-lite package.
   * This makes it hard to modify the specs without breaking typescript
   * interface. For now we have removed the types for the spec and will
   * add them back when we have the support for it.
   * More at https://github.com/vega/vega-lite/issues/9222
   */

  const sanitizedSpec = structuredClone(spec);
  let xEncoding;
  let yEncoding;
  if (sanitizedSpec.encoding) {
    xEncoding = sanitizedSpec.encoding.x;
    yEncoding = sanitizedSpec.encoding.y;

    xEncoding.scale = {
      domain: [xMin.toISOString(), xMax.toISOString()],
    };
  }

  if (!xEncoding || !yEncoding) {
    return sanitizedSpec;
  }

  // Set extents for x-axis
  xEncoding.scale = {
    domain: [xMin.toISOString(), xMax.toISOString()],
  };

  const timeLabelFormat = TIME_GRAIN[timeGrain]?.d3format as string;
  // Remove titles from axes
  xEncoding.axis = {
    ticks: false,
    orient: "top",
    title: "",
    formatType: "time",
    format: timeLabelFormat,
  };
  yEncoding.axis = { title: "" };

  if (
    chartType === TDDChart.STACKED_BAR ||
    chartType === TDDChart.GROUPED_BAR
  ) {
    // Set timeUnit for x-axis using timeGrain
    const timeUnit = timeGrainToVegaTimeUnitMap[timeGrain];
    xEncoding.timeUnit = timeUnit;
  }

  return sanitizedSpec;
}

const timeGrainToVegaTimeUnitMap = {
  [V1TimeGrain.TIME_GRAIN_SECOND]: "yearmonthdatehoursminutesseconds",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "yearmonthdatehoursminutes",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "yearmonthdatehours",
  [V1TimeGrain.TIME_GRAIN_DAY]: "yearmonthdate",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "yearweek",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "yearmonth",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "yearquarter",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "year",
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "yearmonthdate",
};
