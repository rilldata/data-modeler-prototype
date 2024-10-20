import { createSparkline } from "@rilldata/web-common/components/data-graphic/marks/sparkline";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  getDimensionValueTimeSeries,
  type DimensionDataItem,
} from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import { type TimeSeriesDatum } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
import { formatProperFractionAsPercent } from "@rilldata/web-common/lib/number-formatting/proper-fraction-formatter";
import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";

import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import { derived, writable, type Readable } from "svelte/store";
import { memoizeMetricsStore } from "../state-managers/memoize-metrics-store";
import type {
  ChartInteractionColumns,
  HeaderData,
  HighlightedCell,
  TDDCellData,
  TDDComparison,
  TableData,
  TablePosition,
} from "./types";
import { transposeArray } from "./util";

type MeasureValue = number | null | undefined;

export type TimeDimensionDataState = {
  isFetching: boolean;
  isError?: boolean;
  comparing?: TDDComparison;
  data?: TableData;
};

export type TimeSeriesDataStore = Readable<TimeDimensionDataState>;

function sanitizeMeasure(value: unknown): MeasureValue {
  if (value === null || value === undefined) return value;
  if (typeof value === "number") return value;

  console.warn("Invalid type for measure value", value);
  // fail safely by returning null
  return null;
}

function getHeaderDataForRow(
  row: DimensionDataItem,
  isAllTime: boolean,
  measureName: string,

  validPercentOfTotal: boolean,
  unfilteredTotal: number,
) {
  const rowData = isAllTime ? row?.data?.slice(1) : row?.data?.slice(1, -1);
  const dataRow = [
    { value: row?.value },
    {
      value: row?.total?.toString() ?? "",
      spark: createSparkline(rowData, (v) =>
        typeof v?.[measureName] === "number" ? (v[measureName] as number) : 0,
      ),
    },
  ];
  if (validPercentOfTotal) {
    const percOfTotal = (row?.total ?? 0) / unfilteredTotal;
    dataRow.push({
      value: isNaN(percOfTotal)
        ? "...%"
        : numberPartsToString(formatProperFractionAsPercent(percOfTotal)),
    });
  }
  return dataRow;
}

/***
 * Add totals row from time series data
 * Add rest of dimension values from dimension table data
 * Transpose the data to columnar format
 */
export function prepareDimensionData(
  totalsData: TimeSeriesDatum[] | undefined,
  data: DimensionDataItem[],
  total: number,
  unfilteredTotal: number,
  measure: MetricsViewSpecMeasureV2 | undefined,
  selectedValues: string[],
  isAllTime: boolean,
  pinIndex: number,
): TableData | undefined {
  if (!data || !totalsData || !measure) return undefined;

  const measureName = measure?.name as string;
  const validPercentOfTotal = measure?.validPercentOfTotal as boolean;

  // Prepare Columns
  const totalsTableData = isAllTime
    ? totalsData?.slice(1)
    : totalsData?.slice(1, -1);
  const columnHeaderData = totalsTableData?.map((v) => [{ value: v.ts }]);

  const columnCount = columnHeaderData?.length;

  // Prepare Row order
  let orderedData: DimensionDataItem[] = [];

  if (pinIndex > -1 && selectedValues.length && data.length) {
    const selectedValuesIndex = selectedValues
      .slice(0, pinIndex + 1)
      .map((v) => data.findIndex((d) => d.value === v))
      .sort((a, b) => a - b);

    // return if computing on old data
    if (selectedValuesIndex.some((v) => v === -1)) return;

    orderedData = orderedData.concat(
      selectedValuesIndex?.map((i) => {
        return data[i];
      }),
    );

    orderedData = orderedData.concat(
      data?.filter((_, i) => !selectedValuesIndex.includes(i)),
    );
  } else {
    orderedData = data;
  }

  // Add totals row to count
  const rowCount = data?.length + 1;

  const totalsRow = [
    { value: "Total" },
    {
      value: total?.toString(),
      spark: createSparkline(totalsTableData, (v) =>
        typeof v?.[measureName] === "number" ? (v[measureName] as number) : 0,
      ),
    },
  ];

  let fixedColCount = 2;
  if (validPercentOfTotal) {
    fixedColCount = 3;
    const percOfTotal = total / unfilteredTotal;
    totalsRow.push({
      value: isNaN(percOfTotal)
        ? "...%"
        : numberPartsToString(formatProperFractionAsPercent(percOfTotal)),
    });
  }
  let rowHeaderData: HeaderData<string>[][] = [totalsRow];

  rowHeaderData = rowHeaderData.concat(
    orderedData?.map((row) => {
      return getHeaderDataForRow(
        row,
        isAllTime,
        measureName,
        validPercentOfTotal,
        unfilteredTotal,
      );
    }),
  );

  let body: TDDCellData[][] = [
    totalsTableData?.map((v) => sanitizeMeasure(v[measureName])) || [],
  ];

  body = body?.concat(
    orderedData?.map((v) => {
      if (v?.isFetching)
        return new Array(columnCount).fill(undefined) as undefined[];
      const dimData = isAllTime ? v?.data?.slice(1) : v?.data?.slice(1, -1);
      return dimData?.map((v) => sanitizeMeasure(v[measureName]));
    }),
  );
  /* 
    Important: regular-table expects body data in columnar format,
    aka an array of arrays where outer array is the columns,
    inner array is the row values for a specific column
  */
  const columnarBody = transposeArray(body, rowCount, columnCount);

  return {
    rowCount,
    fixedColCount,
    rowHeaderData,
    columnCount,
    columnHeaderData,
    body: columnarBody,
    selectedValues: selectedValues,
  };
}

/***
 * Add totals row from time series data
 * Add Current, Previous, Percentage Change, Absolute Change rows for time comparison
 * Transpose the data to columnar format
 */
export function prepareTimeData(
  data: TimeSeriesDatum[] | undefined,
  total: number,
  comparisonTotal: number | undefined,
  currentLabel: string,
  comparisonLabel: string,
  measure: MetricsViewSpecMeasureV2 | undefined,
  hasTimeComparison: boolean,
  isAllTime: boolean,
): TableData | undefined {
  if (!data || !measure) return undefined;

  const measureName = measure?.name ?? "";

  /** Strip out data points out of chart view */
  const tableData = isAllTime ? data?.slice(1) : data?.slice(1, -1);
  const columnHeaderData = tableData?.map((v) => [{ value: v.ts }]);

  const columnCount = columnHeaderData?.length;

  let rowHeaderData: HeaderData<string>[][] = [];
  rowHeaderData.push([
    { value: "Total" },
    {
      value: total?.toString() ?? "",
      spark: createSparkline(tableData, (v) =>
        typeof v?.[measureName] === "number" ? (v[measureName] as number) : 0,
      ),
    },
  ]);

  const body: TDDCellData[][] = [];

  if (hasTimeComparison) {
    rowHeaderData = rowHeaderData.concat([
      [
        { value: currentLabel },
        {
          value: total?.toString() ?? "",
          spark: createSparkline(tableData, (v) =>
            typeof v?.[measureName] === "number"
              ? (v[measureName] as number)
              : 0,
          ),
        },
      ],
      [
        { value: comparisonLabel },
        {
          value: comparisonTotal?.toString() ?? "",
          spark: createSparkline(tableData, (v) =>
            typeof v?.[`comparison.${measureName}`] === "number"
              ? (v[`comparison.${measureName}`] as number)
              : 0,
          ),
        },
      ],
      [{ value: "Percentage Change" }],
      [{ value: "Absolute Change" }],
    ]);

    // Push totals
    body.push(
      tableData?.map((v) => {
        if (v[measureName] === null && v[`comparison.${measureName}`] === null)
          return null;

        const total =
          (sanitizeMeasure(v[measureName]) || 0) +
          (sanitizeMeasure(v[`comparison.${measureName}`]) || 0);
        return total;
      }),
    );

    // Push current range
    body.push(tableData?.map((v) => sanitizeMeasure(v[measureName])));

    body.push(
      tableData?.map((v) => sanitizeMeasure(v[`comparison.${measureName}`])),
    );

    // Push percentage change
    body.push(
      tableData?.map((v) => {
        const comparisonValue = v[`comparison.${measureName}`] as
          | number
          | null
          | undefined;
        const currentValue = sanitizeMeasure(v[measureName]);
        const comparisonPercChange =
          comparisonValue && currentValue !== undefined && currentValue !== null
            ? (currentValue - comparisonValue) / comparisonValue
            : null;
        if (comparisonPercChange === null) return null;
        return numberPartsToString(
          formatMeasurePercentageDifference(comparisonPercChange),
        );
      }),
    );

    // Push absolute change
    body.push(
      tableData?.map((v) => {
        const comparisonValue = v[`comparison.${measureName}`] as
          | number
          | null
          | undefined;
        const currentValue = sanitizeMeasure(v[measureName]);
        const change =
          comparisonValue && currentValue !== undefined && currentValue !== null
            ? currentValue - comparisonValue
            : null;

        if (change === null) return null;
        return change;
      }),
    );
  } else {
    body.push(tableData?.map((v) => sanitizeMeasure(v[measureName])));
  }

  const rowCount = rowHeaderData.length;
  const columnarBody = transposeArray(body, rowCount, columnCount);

  return {
    rowCount,
    fixedColCount: 2,
    rowHeaderData,
    columnCount,
    columnHeaderData,
    body: columnarBody,
    selectedValues: [],
  };
}

function createDimensionTableData(
  ctx: StateManagers,
): Readable<DimensionDataItem[]> {
  return derived(ctx.dashboardStore, (dashboardStore, set) => {
    const measureName = dashboardStore?.tdd?.expandedMeasureName;
    if (!measureName) return set([]);
    return derived(
      getDimensionValueTimeSeries(ctx, [measureName], "table"),
      (data) => data,
    ).subscribe(set);
  });
}

/**
 * Memoized version of the table data. Currently, memoized by metrics view name.
 */
export const useDimensionTableData = memoizeMetricsStore<
  Readable<DimensionDataItem[]>
>((ctx: StateManagers) => createDimensionTableData(ctx));

/**
 * Stores for handling interactions between chart and table
 * Two separate stores created to avoid looped updates and renders
 */
export const tableInteractionStore = writable<HighlightedCell>({
  dimensionValue: undefined,
  time: undefined,
});

export const chartInteractionColumn = writable<ChartInteractionColumns>({
  yHover: undefined,
  xHover: undefined,
  scrubStart: undefined,
  scrubEnd: undefined,
});

export const lastKnownPosition = writable<TablePosition>(undefined);
