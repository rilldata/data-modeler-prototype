import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { FormatPreset } from "../../humanize-numbers";
import { activeMeasure } from "./core-selectors";
import type { SelectorFnArgs } from "./types";

export const formattingSelectors = {
  /**
   * Gets the sort type for the dash (value, percent, delta, etc.)
   */
  activeMeasureFormatPreset: ([
    dashboard,
    metricsSpecQueryResult,
  ]: SelectorFnArgs): FormatPreset =>
    (activeMeasure([dashboard, metricsSpecQueryResult])
      ?.formatPreset as FormatPreset) ?? FormatPreset.HUMANIZE,

  activeMeasureFormatter: ([
    dashboard,
    metricsSpecQueryResult,
  ]: SelectorFnArgs) =>
    createMeasureValueFormatter(
      activeMeasure([dashboard, metricsSpecQueryResult])
    ),
};
