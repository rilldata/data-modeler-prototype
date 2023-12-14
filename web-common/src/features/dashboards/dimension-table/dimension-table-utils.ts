import DeltaChange from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChange.svelte";
import DeltaChangePercentage from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChangePercentage.svelte";
import PercentOfTotal from "./PercentOfTotal.svelte";

import { PERC_DIFF } from "../../../components/data-types/type-utils";
import type {
  MetricsViewDimension,
  MetricsViewSpecMeasureV2,
  V1MetricsViewComparisonRow,
  V1MetricsViewComparisonValue,
  V1MetricsViewFilter,
  V1MetricsViewToplistResponseDataItem,
} from "../../../runtime-client";

import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
import type { VirtualizedTableConfig } from "@rilldata/web-common/components/virtualized-table/types";

import type { SvelteComponent } from "svelte";
import { getDimensionColumn } from "../dashboard-utils";
import type { DimensionTableRow } from "./dimension-table-types";
import { getFiltersForOtherDimensions } from "../selectors";
import { SortType } from "../proto-state/derived-types";
import type { MetricsExplorerEntity } from "../stores/metrics-explorer-entity";
import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";

/** Returns an updated filter set for a given dimension on search */
export function updateFilterOnSearch(
  filterForDimension: V1MetricsViewFilter,
  searchText: string,
  dimensionName: string
): V1MetricsViewFilter {
  const filterSet = JSON.parse(JSON.stringify(filterForDimension));
  const addNull = "null".includes(searchText);
  if (searchText !== "") {
    let foundDimension = false;

    filterSet["include"].forEach((filter) => {
      if (filter.name === dimensionName) {
        filter.like = [`%${searchText}%`];
        foundDimension = true;
        if (addNull) filter.in.push(null);
      }
    });

    if (!foundDimension) {
      filterSet["include"].push({
        name: dimensionName,
        in: addNull ? [null] : [],
        like: [`%${searchText}%`],
      });
    }
  } else {
    filterSet["include"] = filterSet["include"].filter((f) => f.in.length);
    filterSet["include"].forEach((f) => {
      delete f.like;
    });
  }
  return filterSet;
}

export function getDimensionFilterWithSearch(
  filters: V1MetricsViewFilter,
  searchText: string,
  dimensionName: string
) {
  const filterForDimension = getFiltersForOtherDimensions(
    filters,
    dimensionName
  );

  return updateFilterOnSearch(filterForDimension, searchText, dimensionName);
}

export function computePercentOfTotal(
  values: V1MetricsViewToplistResponseDataItem[],
  total: number,
  measureName: string
) {
  for (const value of values) {
    if (total === 0 || total === null || total === undefined) {
      value[measureName + "_percent_of_total"] =
        PERC_DIFF.CURRENT_VALUE_NO_DATA;
    } else {
      value[measureName + "_percent_of_total"] =
        formatMeasurePercentageDifference(value[measureName] / total);
    }
  }

  return values;
}

export function getComparisonProperties(
  measureName: string,
  selectedMeasure: MetricsViewSpecMeasureV2
): {
  /**
   * "component" in this context is a Svelte component that will be
   * used to render the column header.
   */
  component: typeof SvelteComponent<any>;
  type: string;
  format: string;
  description: string;
} {
  if (measureName.includes("_delta_perc"))
    return {
      component: DeltaChangePercentage,
      type: "RILL_PERCENTAGE_CHANGE",
      format: FormatPreset.PERCENTAGE,
      description: "Perc. change over comparison period",
    };
  else if (measureName.includes("_delta")) {
    return {
      component: DeltaChange,
      type: "RILL_CHANGE",
      format: selectedMeasure.formatPreset ?? FormatPreset.HUMANIZE,
      description: "Change over comparison period",
    };
  } else if (measureName.includes("_percent_of_total")) {
    return {
      component: PercentOfTotal,
      type: "RILL_PERCENTAGE_CHANGE",
      format: FormatPreset.PERCENTAGE,
      description: "Percent of total",
    };
  }
  throw new Error(
    "Invalid measure name, getComparisonProperties must only be called on context columns"
  );
}

export function estimateColumnCharacterWidths(
  columns: VirtualizedTableColumns[],
  rows: V1MetricsViewToplistResponseDataItem[]
) {
  const columnWidths: { [key: string]: number } = {};
  let largestColumnLength = 0;
  columns.forEach((column, i) => {
    // get values
    const values = rows
      .filter((row) => row[column.name] !== null)
      .map(
        (row) =>
          `${row["__formatted_" + column.name] || row[column.name]}`.length
      );
    values.sort();
    const largest = Math.max(...values);
    columnWidths[column.name] = largest;
    if (i != 0) {
      largestColumnLength = Math.max(
        largestColumnLength,
        column.label?.length || column.name.length
      );
    }
  });
  return { columnWidths, largestColumnLength };
}

/** this is a perceived character width value, in pixels, when our monospace
 * font is 12px high. */
const CHARACTER_WIDTH = 7;
const CHARACTER_X_PAD = 16 * 2;
const HEADER_ICON_WIDTHS = 16;
const HEADER_X_PAD = CHARACTER_X_PAD;
const HEADER_FLEX_SPACING = 14;
// const CHARACTER_LIMIT_FOR_WRAPPING = 9;

export function estimateColumnSizes(
  columns: VirtualizedTableColumns[],
  columnWidths: {
    [key: string]: number;
  },
  containerWidth: number,
  config: VirtualizedTableConfig
): number[] {
  const estimateColumnSize = columns.map((column, i) => {
    if (column.name.includes("delta")) return config.comparisonColumnWidth;
    if (i != 0) return config.defaultColumnWidth;

    const largestStringLength =
      columnWidths[column.name] * CHARACTER_WIDTH + CHARACTER_X_PAD;

    /** The header width is largely a function of the total number of characters in the column.*/
    const headerWidth =
      (column.label?.length || column.name.length) * CHARACTER_WIDTH +
      HEADER_ICON_WIDTHS +
      HEADER_X_PAD +
      HEADER_FLEX_SPACING;

    /** If the header is bigger than the largestStringLength and that's not at threshold, default to threshold.
     * This will prevent the case where we have very long column names for very short column values.
     */
    const effectiveHeaderWidth =
      headerWidth > 160 && largestStringLength < 160
        ? config.minHeaderWidthWhenColumsAreSmall
        : headerWidth;

    return largestStringLength
      ? Math.min(
          config.maxColumnWidth,
          Math.max(
            largestStringLength,
            effectiveHeaderWidth,
            /** All columns must be minColumnWidth regardless of user settings. */
            config.minColumnWidth
          )
        )
      : /** if there isn't a longet string length for some reason, let's go with a
         * default column width. We should not be in this state.
         */
        config.defaultColumnWidth;
  });

  const measureColumnSizeSum = estimateColumnSize
    .slice(1)
    .reduce((a, b) => a + b, 0);

  /* Dimension column should expand to cover whole container */
  estimateColumnSize[0] = Math.max(
    containerWidth - measureColumnSizeSum - config.indexWidth,
    estimateColumnSize[0]
  );

  return estimateColumnSize;
}

export function prepareVirtualizedDimTableColumns(
  dash: MetricsExplorerEntity,
  allMeasures: MetricsViewSpecMeasureV2[],
  measureTotals: { [key: string]: number },
  dimension: MetricsViewDimension,
  timeComparison: boolean,
  validPercentOfTotal: boolean
): VirtualizedTableColumns[] {
  const sortType = dash.dashboardSortType;
  const sortDirection = dash.sortDirection;

  const measureNames = allMeasures.map((m) => m.name);
  const leaderboardMeasureName = dash.leaderboardMeasureName;
  const selectedMeasure = allMeasures.find(
    (m) => m.name === leaderboardMeasureName
  );

  const dimensionColumn = getDimensionColumn(dimension);

  // copy column names so we don't mutate the original
  const columnNames = [...dash.visibleMeasureKeys];

  // don't add context columns if sorting by dimension
  if (selectedMeasure && sortType !== SortType.DIMENSION) {
    addContextColumnNames(
      columnNames,
      timeComparison,
      validPercentOfTotal,
      selectedMeasure
    );
  }
  // Make dimension the first column
  columnNames.unshift(dimensionColumn);

  const columns = columnNames
    .map((name) => {
      let highlight = false;
      if (sortType === SortType.DIMENSION) {
        highlight = name === dimensionColumn;
      } else {
        highlight =
          name === selectedMeasure?.name ||
          name.endsWith("_delta") ||
          name.endsWith("_delta_perc") ||
          name.endsWith("_percent_of_total");
      }

      let sorted;
      if (name.endsWith("_delta") && sortType === SortType.DELTA_ABSOLUTE) {
        sorted = sortDirection;
      } else if (
        name.endsWith("_delta_perc") &&
        sortType === SortType.DELTA_PERCENT
      ) {
        sorted = sortDirection;
      } else if (
        name.endsWith("_percent_of_total") &&
        sortType === SortType.PERCENT
      ) {
        sorted = sortDirection;
      } else if (
        name === selectedMeasure?.name &&
        sortType === SortType.VALUE
      ) {
        sorted = sortDirection;
      }

      let columnOut: VirtualizedTableColumns | undefined = undefined;
      if (measureNames.includes(name)) {
        // Handle all regular measures
        const measure = allMeasures.find((m) => m.name === name);
        columnOut = {
          name,
          type: "INT",
          label: measure?.label || measure?.expression,
          description: measure?.description,
          total: measureTotals[measure?.name ?? ""] || 0,
          enableResize: false,
          format: measure?.formatPreset,
          highlight,
          sorted,
        };
      } else if (name === dimensionColumn) {
        // Handle dimension column
        columnOut = {
          name,
          type: "VARCHAR",
          label: dimension?.label,
          enableResize: true,
          highlight,
          sorted,
        };
      } else if (selectedMeasure !== undefined) {
        // Handle delta, delta_perc, and percent_of_total columns
        const comparison = getComparisonProperties(name, selectedMeasure);
        columnOut = {
          name,
          type: comparison.type,
          label: comparison.component,
          description: comparison.description,
          enableResize: false,
          format: comparison.format,
          highlight,
          sorted,
        };
      }
      return columnOut;
    })
    .filter((column) => column !== undefined);

  // cast is safe, because we filtered out undefined columns
  return (columns as VirtualizedTableColumns[]) ?? [];
}

/**
 * Splices the context column names into the list of dimension
 * table column names.
 *
 * This mutates the columnNames array.
 */
export function addContextColumnNames(
  columnNames: string[],
  timeComparison: boolean,
  validPercentOfTotal: boolean,
  selectedMeasure: MetricsViewSpecMeasureV2
) {
  const name = selectedMeasure?.name;
  if (!name) return;

  const sortByColumnIndex = columnNames.indexOf(name);
  // Add comparison columns if available
  let percentOfTotalSpliceIndex = 1;
  const isPercent = selectedMeasure?.formatPreset === FormatPreset.PERCENTAGE;
  if (timeComparison) {
    percentOfTotalSpliceIndex = 2;
    columnNames.splice(sortByColumnIndex + 1, 0, `${name}_delta`);

    // Only push percentage delta column if selected measure is not a percentage
    if (!isPercent) {
      percentOfTotalSpliceIndex = 3;
      columnNames.splice(sortByColumnIndex + 2, 0, `${name}_delta_perc`);
    }
  }
  // Only push percentage-of-total if selected measure is
  // validPercentOfTotal and not a percentage
  if (validPercentOfTotal && !isPercent) {
    columnNames.splice(
      sortByColumnIndex + percentOfTotalSpliceIndex,
      0,
      `${name}_percent_of_total`
    );
  }
}

function castUnknownToNumberOrNull(val: unknown): number | null {
  if (typeof val === "number") return val;
  if (val === null || val === undefined) return null;
  console.warn(
    `castUnknownNumberOrNull should only be used to cast unknowns that should be numbers, null, or undefined to numbers or null. Got: ${val}`
  );
  return val as number;
}

/**
 * This function prepares the data for the dimension table
 * from data returned by the createQueryServiceMetricsViewComparison
 * API.
 *
 */
export function prepareDimensionTableRows(
  queryRows: V1MetricsViewComparisonRow[],
  // all of the measures defined for this metrics spec,
  // including those that are not visible
  allMeasuresForSpec: MetricsViewSpecMeasureV2[],
  activeMeasureName: string,
  dimensionColumn: string,
  addDeltas: boolean,
  addPercentOfTotal: boolean,
  unfilteredTotal: number
): DimensionTableRow[] {
  if (!queryRows || !queryRows.length) return [];

  const formattersForMeasures: { [key: string]: (val: number) => string } =
    Object.fromEntries(
      allMeasuresForSpec.map((m) => [m.name, createMeasureValueFormatter(m)])
    );

  const tableRows: DimensionTableRow[] = queryRows
    .filter(
      (row) => row.measureValues !== undefined && row.measureValues !== null
    )
    .map((row) => {
      // cast is safe since we filtered out rows without measureValues
      const measureValues = row.measureValues as V1MetricsViewComparisonValue[];

      const rawVals: [string, number | null][] = measureValues.map((m) => [
        m.measureName?.toString() ?? "",
        castUnknownToNumberOrNull(m.baseValue),
      ]);

      const formattedVals: [string, string | number | PERC_DIFF][] =
        rawVals.map(([name, val]) => [
          "__formatted_" + name,
          val !== null
            ? formattersForMeasures[name](val)
            : PERC_DIFF.CURRENT_VALUE_NO_DATA,
        ]);

      const rowOut: DimensionTableRow = Object.fromEntries([
        [dimensionColumn, row.dimensionValue as string],
        ...rawVals,
        ...formattedVals,
      ]);

      const activeMeasure = measureValues.find(
        (m) => m.measureName === activeMeasureName
      );

      if (addDeltas && activeMeasure) {
        rowOut[`${activeMeasureName}_delta`] = castUnknownToNumberOrNull(
          activeMeasure.deltaAbs
        );

        rowOut[`__formatted_${activeMeasureName}_delta`] =
          activeMeasure.deltaAbs
            ? formattersForMeasures[activeMeasureName](
                activeMeasure.deltaAbs as number
              )
            : PERC_DIFF.PREV_VALUE_NO_DATA;

        rowOut[`${activeMeasureName}_delta_perc`] = castUnknownToNumberOrNull(
          activeMeasure.deltaRel
        );

        rowOut[`__formatted_${activeMeasureName}_delta_perc`] =
          activeMeasure.deltaRel
            ? formatMeasurePercentageDifference(
                activeMeasure.deltaRel as number
              )
            : PERC_DIFF.PREV_VALUE_NO_DATA;
      }

      if (addPercentOfTotal && activeMeasure) {
        const value = castUnknownToNumberOrNull(activeMeasure.baseValue);

        if (value === null || unfilteredTotal === 0 || !unfilteredTotal) {
          rowOut[activeMeasureName + "_percent_of_total"] =
            PERC_DIFF.CURRENT_VALUE_NO_DATA;

          rowOut[`__formatted_${activeMeasureName}_percent_of_total`] =
            PERC_DIFF.CURRENT_VALUE_NO_DATA;
        } else {
          rowOut[activeMeasureName + "_percent_of_total"] =
            value / unfilteredTotal;

          rowOut[`__formatted_${activeMeasureName}_percent_of_total`] =
            formatMeasurePercentageDifference(value / unfilteredTotal);
        }
      }

      return rowOut;
    });
  return tableRows;
}

export function getSelectedRowIndicesFromFilters(
  rows: DimensionTableRow[],
  filters: V1MetricsViewFilter,
  dimensionName: string,
  excludeMode: boolean
): number[] {
  const selectedDimValues =
    ((excludeMode
      ? filters?.exclude?.find((d) => d.name === dimensionName)?.in
      : filters?.include?.find((d) => d.name === dimensionName)
          ?.in) as string[]) ?? [];

  return selectedDimValues
    .map((label) => {
      return rows.findIndex((row) => row[dimensionName] === label);
    })
    .filter((i) => i >= 0);
}
