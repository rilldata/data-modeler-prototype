import { ChartType } from "@rilldata/web-common/features/canvas-components/types";
import { VisualizationSpec } from "svelte-vega";
import { buildArea } from "./area";
import { buildGroupedBar } from "./grouped-bar";
import { buildLine } from "./line";
import {
  buildSimpleBarSingleLayer,
  buildSimpleBarMultiLayer,
} from "./simple-bar";
import { buildStackedArea } from "./stacked-area";
import { buildStackedBar } from "./stacked-bar";

const BAR_LIKE_CHARTS = [
  ChartType.BAR,
  ChartType.GROUPED_BAR,
  ChartType.STACKED_BAR,
];
const LINE_LIKE_CHARTS = [
  ChartType.LINE,
  ChartType.AREA,
  ChartType.STACKED_AREA,
];

export interface ChartField {
  name: string;
  label: string;
  // for custom tooltip field names
  tooltipName?: string;
  // for mapping nominal values to colors
  values?: (number | undefined | string | null)[];
  formatterFunction?: string;
}

export function buildVegaLiteSpec(
  chartType: ChartType,
  timeFields: ChartField[],
  quantitativeFields: ChartField[],
  nominalFields: ChartField[] = [],
): VisualizationSpec {
  if (!timeFields.length) throw "No time fields found";
  const hasNominalFields = nominalFields.length > 0;

  if (BAR_LIKE_CHARTS.includes(chartType)) {
    if (!hasNominalFields) {
      return buildSimpleBarMultiLayer(timeFields[0], quantitativeFields[0]);
    } else if (chartType === ChartType.GROUPED_BAR) {
      // TODO: to configure multi layer for full height hover when comparison is enabled
      return buildGroupedBar(
        timeFields[0],
        quantitativeFields[0],
        nominalFields[0],
      );
    } else if (chartType === ChartType.STACKED_BAR) {
      // TODO: to configure multi layer for full height hover when comparison is enabled
      return buildStackedBar(
        timeFields[0],
        quantitativeFields[0],
        nominalFields[0],
      );
    } else
      return buildSimpleBarSingleLayer(timeFields[0], quantitativeFields[0]);
  } else if (LINE_LIKE_CHARTS.includes(chartType)) {
    if (chartType == ChartType.AREA) {
      return buildArea(timeFields[0], quantitativeFields[0]);
    } else if (chartType == ChartType.LINE) {
      return buildLine(timeFields[0], quantitativeFields[0], nominalFields[0]);
    } else if (chartType == ChartType.STACKED_AREA && hasNominalFields) {
      return buildStackedArea(
        timeFields[0],
        quantitativeFields[0],
        nominalFields[0],
      );
    } else return buildArea(timeFields[0], quantitativeFields[0]);
  } else {
    throw new Error(`Chart type "${chartType}" not supported.`);
  }
}
