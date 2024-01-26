import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  V1Expression,
  V1MetricsViewAggregationResponseDataItem,
  V1MetricsViewFilter,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { ExpandedState, SortingState } from "@tanstack/svelte-table";

export interface PivotState {
  active: boolean;
  rows: string[];
  columns: string[];
  expanded: ExpandedState;
  sorting: SortingState;
  rowJoinType: "flat" | "nest";
}

export interface PivotDataRow {
  [key: string]: string | number | PivotDataRow[] | undefined;
  subRows?: PivotDataRow[];
}

export interface TimeFilters {
  timeStart: string;
  interval: V1TimeGrain;
}

export interface PivotTimeConfig {
  timeStart: string | undefined;
  timeEnd: string | undefined;
  timeZone: string;
  timeDimension: string;
  interval: V1TimeGrain;
}

/**
 * This is the config that is passed to the pivot data store methods
 */
export interface PivotDataStoreConfig {
  measureNames: string[];
  rowDimensionNames: string[];
  colDimensionNames: string[];
  allMeasures: MetricsViewSpecMeasureV2[];
  allDimensions: MetricsViewSpecDimensionV2[];
  filters: V1MetricsViewFilter;
  whereFilter: V1Expression;
  pivot: PivotState;
  time: PivotTimeConfig;
}

export interface PivotAxesData {
  isFetching: boolean;
  data?: Record<string, string[]> | undefined;
  totals?:
    | Record<string, V1MetricsViewAggregationResponseDataItem[]>
    | undefined;
}

// OLD PIVOT TYPES
export type PivotMeasure = {
  def: string;
  minichart?: boolean;
  minichartDimension?: string;
  /* expand with other props over time as needed */
};

export type PivotDimension = {
  def: string;
  /* other props like sort criteria, limits can go here */
};

export type PivotColumnSet = {
  dims: PivotDimension[];
  measures: PivotMeasure[];
};

export type PivotConfig = {
  rowDims: PivotDimension[];
  colSets: PivotColumnSet[];
  rowJoinType: "flat" | "nest";
  sort: any; // TBD
  expanded: any[];
};

export type PivotPos = {
  x0: number;
  x1: number;
  y0: number;
  y1: number;
};

export type PivotRenderCallback = (data: {
  x: number;
  y: number;
  value: any;
  element: HTMLElement;
}) => string | void;
