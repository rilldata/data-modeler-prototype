// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file rill/ui/v1/dashboard.proto (package rill.ui.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Timestamp } from "@bufbuild/protobuf";
import { MetricsViewFilter } from "../../runtime/v1/queries_pb.js";
import { Expression } from "../../runtime/v1/expression_pb.js";
import { TimeGrain } from "../../runtime/v1/time_grain_pb.js";

/**
 * DashboardState represents the dashboard as seen by the user
 *
 * @generated from message rill.ui.v1.DashboardState
 */
export class DashboardState extends Message<DashboardState> {
  /**
   * Selected time range
   *
   * @generated from field: rill.ui.v1.DashboardTimeRange time_range = 1;
   */
  timeRange?: DashboardTimeRange;

  /**
   * Dimension filters applied
   *
   * @generated from field: rill.runtime.v1.MetricsViewFilter filters = 2;
   */
  filters?: MetricsViewFilter;

  /**
   * Expression format for dimension filters
   *
   * @generated from field: rill.runtime.v1.Expression where = 20;
   */
  where?: Expression;

  /**
   * Expression format for measure filters
   *
   * @generated from field: repeated rill.ui.v1.DashboardDimensionFilter having = 21;
   */
  having: DashboardDimensionFilter[] = [];

  /**
   * Selected time granularity
   *
   * @generated from field: rill.runtime.v1.TimeGrain time_grain = 3;
   */
  timeGrain = TimeGrain.UNSPECIFIED;

  /**
   * @generated from field: rill.ui.v1.DashboardTimeRange compare_time_range = 4;
   */
  compareTimeRange?: DashboardTimeRange;

  /**
   * Selected measure for the leaderboard
   *
   * @generated from field: optional string leaderboard_measure = 5;
   */
  leaderboardMeasure?: string;

  /**
   * Focused dimension
   *
   * @generated from field: optional string selected_dimension = 6;
   */
  selectedDimension?: string;

  /**
   * @generated from field: optional bool show_time_comparison = 7;
   */
  showTimeComparison?: boolean;

  /**
   * Selected measures and dimensions to be shown
   *
   * @generated from field: repeated string visible_measures = 8;
   */
  visibleMeasures: string[] = [];

  /**
   * @generated from field: optional bool all_measures_visible = 9;
   */
  allMeasuresVisible?: boolean;

  /**
   * @generated from field: repeated string visible_dimensions = 10;
   */
  visibleDimensions: string[] = [];

  /**
   * @generated from field: optional bool all_dimensions_visible = 11;
   */
  allDimensionsVisible?: boolean;

  /**
   * @generated from field: optional rill.ui.v1.DashboardState.LeaderboardContextColumn leaderboard_context_column = 12;
   */
  leaderboardContextColumn?: DashboardState_LeaderboardContextColumn;

  /**
   * Selected timezone for the dashboard
   *
   * @generated from field: optional string selected_timezone = 13;
   */
  selectedTimezone?: string;

  /**
   * Scrub time range
   *
   * @generated from field: optional rill.ui.v1.DashboardTimeRange scrub_range = 14;
   */
  scrubRange?: DashboardTimeRange;

  /**
   * @generated from field: optional rill.ui.v1.DashboardState.LeaderboardSortDirection leaderboard_sort_direction = 15;
   */
  leaderboardSortDirection?: DashboardState_LeaderboardSortDirection;

  /**
   * @generated from field: optional rill.ui.v1.DashboardState.LeaderboardSortType leaderboard_sort_type = 16;
   */
  leaderboardSortType?: DashboardState_LeaderboardSortType;

  /**
   * @generated from field: optional string comparison_dimension = 17;
   */
  comparisonDimension?: string;

  /**
   * Expanded measure for TDD view
   *
   * @generated from field: optional string expanded_measure = 18;
   */
  expandedMeasure?: string;

  /**
   * @generated from field: optional int32 pin_index = 19;
   */
  pinIndex?: number;

  constructor(data?: PartialMessage<DashboardState>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "rill.ui.v1.DashboardState";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "time_range", kind: "message", T: DashboardTimeRange },
    { no: 2, name: "filters", kind: "message", T: MetricsViewFilter },
    { no: 20, name: "where", kind: "message", T: Expression },
    { no: 21, name: "having", kind: "message", T: DashboardDimensionFilter, repeated: true },
    { no: 3, name: "time_grain", kind: "enum", T: proto3.getEnumType(TimeGrain) },
    { no: 4, name: "compare_time_range", kind: "message", T: DashboardTimeRange },
    { no: 5, name: "leaderboard_measure", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 6, name: "selected_dimension", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 7, name: "show_time_comparison", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 8, name: "visible_measures", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 9, name: "all_measures_visible", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 10, name: "visible_dimensions", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 11, name: "all_dimensions_visible", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 12, name: "leaderboard_context_column", kind: "enum", T: proto3.getEnumType(DashboardState_LeaderboardContextColumn), opt: true },
    { no: 13, name: "selected_timezone", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 14, name: "scrub_range", kind: "message", T: DashboardTimeRange, opt: true },
    { no: 15, name: "leaderboard_sort_direction", kind: "enum", T: proto3.getEnumType(DashboardState_LeaderboardSortDirection), opt: true },
    { no: 16, name: "leaderboard_sort_type", kind: "enum", T: proto3.getEnumType(DashboardState_LeaderboardSortType), opt: true },
    { no: 17, name: "comparison_dimension", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 18, name: "expanded_measure", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 19, name: "pin_index", kind: "scalar", T: 5 /* ScalarType.INT32 */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DashboardState {
    return new DashboardState().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DashboardState {
    return new DashboardState().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DashboardState {
    return new DashboardState().fromJsonString(jsonString, options);
  }

  static equals(a: DashboardState | PlainMessage<DashboardState> | undefined, b: DashboardState | PlainMessage<DashboardState> | undefined): boolean {
    return proto3.util.equals(DashboardState, a, b);
  }
}

/**
 * @generated from enum rill.ui.v1.DashboardState.LeaderboardContextColumn
 */
export enum DashboardState_LeaderboardContextColumn {
  /**
   * @generated from enum value: LEADERBOARD_CONTEXT_COLUMN_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: LEADERBOARD_CONTEXT_COLUMN_PERCENT = 1;
   */
  PERCENT = 1,

  /**
   * @generated from enum value: LEADERBOARD_CONTEXT_COLUMN_DELTA_PERCENT = 2;
   */
  DELTA_PERCENT = 2,

  /**
   * @generated from enum value: LEADERBOARD_CONTEXT_COLUMN_DELTA_ABSOLUTE = 3;
   */
  DELTA_ABSOLUTE = 3,

  /**
   * @generated from enum value: LEADERBOARD_CONTEXT_COLUMN_HIDDEN = 4;
   */
  HIDDEN = 4,
}
// Retrieve enum metadata with: proto3.getEnumType(DashboardState_LeaderboardContextColumn)
proto3.util.setEnumType(DashboardState_LeaderboardContextColumn, "rill.ui.v1.DashboardState.LeaderboardContextColumn", [
  { no: 0, name: "LEADERBOARD_CONTEXT_COLUMN_UNSPECIFIED" },
  { no: 1, name: "LEADERBOARD_CONTEXT_COLUMN_PERCENT" },
  { no: 2, name: "LEADERBOARD_CONTEXT_COLUMN_DELTA_PERCENT" },
  { no: 3, name: "LEADERBOARD_CONTEXT_COLUMN_DELTA_ABSOLUTE" },
  { no: 4, name: "LEADERBOARD_CONTEXT_COLUMN_HIDDEN" },
]);

/**
 * @generated from enum rill.ui.v1.DashboardState.LeaderboardSortDirection
 */
export enum DashboardState_LeaderboardSortDirection {
  /**
   * @generated from enum value: LEADERBOARD_SORT_DIRECTION_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: LEADERBOARD_SORT_DIRECTION_ASCENDING = 1;
   */
  ASCENDING = 1,

  /**
   * @generated from enum value: LEADERBOARD_SORT_DIRECTION_DESCENDING = 2;
   */
  DESCENDING = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(DashboardState_LeaderboardSortDirection)
proto3.util.setEnumType(DashboardState_LeaderboardSortDirection, "rill.ui.v1.DashboardState.LeaderboardSortDirection", [
  { no: 0, name: "LEADERBOARD_SORT_DIRECTION_UNSPECIFIED" },
  { no: 1, name: "LEADERBOARD_SORT_DIRECTION_ASCENDING" },
  { no: 2, name: "LEADERBOARD_SORT_DIRECTION_DESCENDING" },
]);

/**
 * *
 * SortType is used to determine how to sort the leaderboard
 * and dimension detail table, as well as where to place the
 * sort arrow.
 *
 * By default, the leaderboards+table will be sorted by VALUE,
 * using the value of the currently selected dashboard measure.
 *
 * If DELTA_ABSOLUTE or DELTA_PERCENT is selected, the
 * leaderboards+table will be sorted by the absolute or percentage
 * delta change of the currently selected dashboard measure.
 *
 * If PERCENT is selected, the table will be sorted by the value
 * of the currently selected dashboard measure, which will return
 * the same ordering as the percent-of-total sort for measures
 * with valid percent-of-total. However, the sort arrow will be
 * placed next to the percent-of-total icon.
 *
 * As of 2023-08, DIMENSION is not implemented, but at that time
 * the plan was to only apply DIMENSTION sort to the dimension
 * detail table, and not the leaderboards.
 *
 * @generated from enum rill.ui.v1.DashboardState.LeaderboardSortType
 */
export enum DashboardState_LeaderboardSortType {
  /**
   * @generated from enum value: LEADERBOARD_SORT_TYPE_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: LEADERBOARD_SORT_TYPE_VALUE = 1;
   */
  VALUE = 1,

  /**
   * @generated from enum value: LEADERBOARD_SORT_TYPE_DIMENSION = 2;
   */
  DIMENSION = 2,

  /**
   * @generated from enum value: LEADERBOARD_SORT_TYPE_PERCENT = 3;
   */
  PERCENT = 3,

  /**
   * @generated from enum value: LEADERBOARD_SORT_TYPE_DELTA_PERCENT = 4;
   */
  DELTA_PERCENT = 4,

  /**
   * @generated from enum value: LEADERBOARD_SORT_TYPE_DELTA_ABSOLUTE = 5;
   */
  DELTA_ABSOLUTE = 5,
}
// Retrieve enum metadata with: proto3.getEnumType(DashboardState_LeaderboardSortType)
proto3.util.setEnumType(DashboardState_LeaderboardSortType, "rill.ui.v1.DashboardState.LeaderboardSortType", [
  { no: 0, name: "LEADERBOARD_SORT_TYPE_UNSPECIFIED" },
  { no: 1, name: "LEADERBOARD_SORT_TYPE_VALUE" },
  { no: 2, name: "LEADERBOARD_SORT_TYPE_DIMENSION" },
  { no: 3, name: "LEADERBOARD_SORT_TYPE_PERCENT" },
  { no: 4, name: "LEADERBOARD_SORT_TYPE_DELTA_PERCENT" },
  { no: 5, name: "LEADERBOARD_SORT_TYPE_DELTA_ABSOLUTE" },
]);

/**
 * @generated from message rill.ui.v1.DashboardTimeRange
 */
export class DashboardTimeRange extends Message<DashboardTimeRange> {
  /**
   * @generated from field: optional string name = 1;
   */
  name?: string;

  /**
   * @generated from field: optional google.protobuf.Timestamp time_start = 2;
   */
  timeStart?: Timestamp;

  /**
   * @generated from field: optional google.protobuf.Timestamp time_end = 3;
   */
  timeEnd?: Timestamp;

  constructor(data?: PartialMessage<DashboardTimeRange>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "rill.ui.v1.DashboardTimeRange";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 2, name: "time_start", kind: "message", T: Timestamp, opt: true },
    { no: 3, name: "time_end", kind: "message", T: Timestamp, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DashboardTimeRange {
    return new DashboardTimeRange().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DashboardTimeRange {
    return new DashboardTimeRange().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DashboardTimeRange {
    return new DashboardTimeRange().fromJsonString(jsonString, options);
  }

  static equals(a: DashboardTimeRange | PlainMessage<DashboardTimeRange> | undefined, b: DashboardTimeRange | PlainMessage<DashboardTimeRange> | undefined): boolean {
    return proto3.util.equals(DashboardTimeRange, a, b);
  }
}

/**
 * @generated from message rill.ui.v1.DashboardDimensionFilter
 */
export class DashboardDimensionFilter extends Message<DashboardDimensionFilter> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: rill.runtime.v1.Expression filter = 2;
   */
  filter?: Expression;

  constructor(data?: PartialMessage<DashboardDimensionFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "rill.ui.v1.DashboardDimensionFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "filter", kind: "message", T: Expression },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DashboardDimensionFilter {
    return new DashboardDimensionFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DashboardDimensionFilter {
    return new DashboardDimensionFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DashboardDimensionFilter {
    return new DashboardDimensionFilter().fromJsonString(jsonString, options);
  }

  static equals(a: DashboardDimensionFilter | PlainMessage<DashboardDimensionFilter> | undefined, b: DashboardDimensionFilter | PlainMessage<DashboardDimensionFilter> | undefined): boolean {
    return proto3.util.equals(DashboardDimensionFilter, a, b);
  }
}

