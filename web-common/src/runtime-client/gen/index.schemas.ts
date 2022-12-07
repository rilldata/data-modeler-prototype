/**
 * Generated by orval v6.10.1 🍺
 * Do not edit manually.
 * rill/runtime/v1/schema.proto
 * OpenAPI spec version: version not set
 */
export type RuntimeServiceReconcileBody = {
  /** Changed paths provides a way to "hint" what files have changed in the repo, enabling
reconciliation to execute faster by not scanning all code artifacts for changes. */
  changedPaths?: string[];
  dry?: boolean;
  strict?: boolean;
};

export type RuntimeServiceQueryBody = {
  args?: unknown[];
  dryRun?: boolean;
  priority?: number;
  sql?: string;
};

/**
 * Request for RuntimeService.GetTopK. Returns the top K values for a given column using agg function for table table_name.
 */
export type RuntimeServiceGetTopKBody = {
  agg?: string;
  k?: number;
  priority?: number;
};

export type RuntimeServiceGenerateTimeSeriesBody = {
  filters?: V1MetricsViewRequestFilter;
  measures?: GenerateTimeSeriesRequestBasicMeasure[];
  pixels?: number;
  priority?: number;
  sampleSize?: number;
  timeRange?: V1TimeSeriesTimeRange;
  timestampColumnName?: string;
};

export type RuntimeServiceGetTimeRangeSummaryParams = { priority?: number };

export type RuntimeServiceEstimateSmallestTimeGrainParams = {
  priority?: number;
};

export type RuntimeServiceGetRugHistogramParams = { priority?: number };

export type RuntimeServiceGetTableRowsParams = {
  limit?: number;
  priority?: number;
};

export type RuntimeServiceEstimateRollupIntervalBody = {
  columnName?: string;
  priority?: number;
};

export type RuntimeServiceGetNumericHistogramParams = { priority?: number };

export type RuntimeServiceGetNullCountParams = { priority?: number };

export type RuntimeServiceGetDescriptiveStatisticsParams = {
  priority?: number;
};

export type RuntimeServiceProfileColumnsParams = { priority?: number };

export type RuntimeServiceGetCardinalityOfColumnParams = { priority?: number };

export type RuntimeServiceGetTableCardinalityParams = { priority?: number };

export type RuntimeServiceMetricsViewTotalsBody = {
  filter?: V1MetricsViewFilter;
  measureNames?: string[];
  priority?: number;
  timeEnd?: string;
  timeStart?: string;
};

export type RuntimeServiceMetricsViewToplistBody = {
  filter?: V1MetricsViewFilter;
  limit?: string;
  measureNames?: string[];
  offset?: string;
  priority?: number;
  sort?: V1MetricsViewSort[];
  timeEnd?: string;
  timeStart?: string;
};

export type RuntimeServiceMetricsViewTimeSeriesBody = {
  filter?: V1MetricsViewFilter;
  measureNames?: string[];
  priority?: number;
  timeEnd?: string;
  timeGranularity?: string;
  timeStart?: string;
};

export type RuntimeServiceRenameFileBody = {
  fromPath?: string;
  toPath?: string;
};

export type RuntimeServicePutFileBody = {
  blob?: string;
  create?: boolean;
  /** Will cause the operation to fail if the file already exists.
It should only be set when create = true. */
  createOnly?: boolean;
};

export type RuntimeServiceListFilesParams = { glob?: string };

export type RuntimeServiceListCatalogEntriesType =
  typeof RuntimeServiceListCatalogEntriesType[keyof typeof RuntimeServiceListCatalogEntriesType];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const RuntimeServiceListCatalogEntriesType = {
  OBJECT_TYPE_UNSPECIFIED: "OBJECT_TYPE_UNSPECIFIED",
  OBJECT_TYPE_TABLE: "OBJECT_TYPE_TABLE",
  OBJECT_TYPE_SOURCE: "OBJECT_TYPE_SOURCE",
  OBJECT_TYPE_MODEL: "OBJECT_TYPE_MODEL",
  OBJECT_TYPE_METRICS_VIEW: "OBJECT_TYPE_METRICS_VIEW",
} as const;

export type RuntimeServiceListCatalogEntriesParams = {
  type?: RuntimeServiceListCatalogEntriesType;
};

export type RuntimeServiceListInstancesParams = {
  pageSize?: number;
  pageToken?: string;
};

export type V1TypeCode = typeof V1TypeCode[keyof typeof V1TypeCode];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1TypeCode = {
  CODE_UNSPECIFIED: "CODE_UNSPECIFIED",
  CODE_BOOL: "CODE_BOOL",
  CODE_INT8: "CODE_INT8",
  CODE_INT16: "CODE_INT16",
  CODE_INT32: "CODE_INT32",
  CODE_INT64: "CODE_INT64",
  CODE_INT128: "CODE_INT128",
  CODE_UINT8: "CODE_UINT8",
  CODE_UINT16: "CODE_UINT16",
  CODE_UINT32: "CODE_UINT32",
  CODE_UINT64: "CODE_UINT64",
  CODE_UINT128: "CODE_UINT128",
  CODE_FLOAT32: "CODE_FLOAT32",
  CODE_FLOAT64: "CODE_FLOAT64",
  CODE_TIMESTAMP: "CODE_TIMESTAMP",
  CODE_DATE: "CODE_DATE",
  CODE_TIME: "CODE_TIME",
  CODE_STRING: "CODE_STRING",
  CODE_BYTES: "CODE_BYTES",
  CODE_ARRAY: "CODE_ARRAY",
  CODE_STRUCT: "CODE_STRUCT",
  CODE_MAP: "CODE_MAP",
  CODE_DECIMAL: "CODE_DECIMAL",
  CODE_JSON: "CODE_JSON",
  CODE_UUID: "CODE_UUID",
} as const;

export interface V1TriggerSyncResponse {
  objectsAddedCount?: number;
  objectsCount?: number;
  objectsRemovedCount?: number;
  objectsUpdatedCount?: number;
}

export interface V1TriggerRefreshResponse {
  [key: string]: any;
}

export interface V1TopK {
  entries?: TopKEntry[];
}

export type V1TimeSeriesValueRecords = { [key: string]: number };

export interface V1TimeSeriesValue {
  bin?: number;
  records?: V1TimeSeriesValueRecords;
  ts?: string;
}

export interface V1TimeSeriesTimeRange {
  end?: string;
  interval?: V1TimeGrain;
  start?: string;
}

export interface V1TimeSeriesResponse {
  results?: V1TimeSeriesValue[];
  sampleSize?: number;
  spark?: V1TimeSeriesValue[];
  timeRange?: V1TimeSeriesTimeRange;
}

export interface V1TimeRangeSummary {
  interval?: TimeRangeSummaryInterval;
  max?: string;
  min?: string;
}

export type V1TimeGrain = typeof V1TimeGrain[keyof typeof V1TimeGrain];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1TimeGrain = {
  TIME_GRAIN_UNSPECIFIED: "TIME_GRAIN_UNSPECIFIED",
  TIME_GRAIN_MILLISECOND: "TIME_GRAIN_MILLISECOND",
  TIME_GRAIN_SECOND: "TIME_GRAIN_SECOND",
  TIME_GRAIN_MINUTE: "TIME_GRAIN_MINUTE",
  TIME_GRAIN_HOUR: "TIME_GRAIN_HOUR",
  TIME_GRAIN_DAY: "TIME_GRAIN_DAY",
  TIME_GRAIN_WEEK: "TIME_GRAIN_WEEK",
  TIME_GRAIN_MONTH: "TIME_GRAIN_MONTH",
  TIME_GRAIN_YEAR: "TIME_GRAIN_YEAR",
} as const;

export interface V1StructType {
  fields?: StructTypeField[];
}

/**
 * Table represents a table in the OLAP database. These include pre-existing tables discovered by periodically
scanning the database's information schema when the instance is created with exposed=true. Pre-existing tables
have managed = false.
 */
export interface V1Table {
  /** Managed is true if the table was created through a runtime migration, false if it was discovered in by
scanning the database's information schema. */
  managed?: boolean;
  name?: string;
  schema?: V1StructType;
}

export type V1SourceProperties = { [key: string]: any };

export interface V1Source {
  connector?: string;
  name?: string;
  properties?: V1SourceProperties;
  schema?: V1StructType;
}

export interface V1RenameFileResponse {
  [key: string]: any;
}

export interface V1RenameFileAndReconcileResponse {
  /** affected_paths lists all the file artifact paths that were considered while
executing the reconciliation. If changed_paths was empty, this will include all
code artifacts in the repo. */
  affectedPaths?: string[];
  /** Errors encountered during reconciliation. If strict = false, any path in
affected_paths without an error can be assumed to have been reconciled succesfully. */
  errors?: V1ReconcileError[];
}

export interface V1RenameFileAndReconcileRequest {
  /** If true, will save the file and validate it and related file artifacts, but not actually execute any migrations. */
  dry?: boolean;
  fromPath?: string;
  instanceId?: string;
  strict?: boolean;
  toPath?: string;
}

export interface V1RefreshAndReconcileRequest {
  /** If true, will save the file and validate it and related file artifacts, but not actually execute any migrations. */
  dry?: boolean;
  instanceId?: string;
  path?: string;
  strict?: boolean;
}

/**
 * - CODE_UNSPECIFIED: Unspecified error
 - CODE_SYNTAX: Code artifact failed to parse
 - CODE_VALIDATION: Code artifact has internal validation errors
 - CODE_DEPENDENCY: Code artifact is valid, but has invalid dependencies
 - CODE_OLAP: Error returned from the OLAP database
 - CODE_SOURCE: Error encountered during source inspection or ingestion
 */
export type V1ReconcileErrorCode =
  typeof V1ReconcileErrorCode[keyof typeof V1ReconcileErrorCode];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1ReconcileErrorCode = {
  CODE_UNSPECIFIED: "CODE_UNSPECIFIED",
  CODE_SYNTAX: "CODE_SYNTAX",
  CODE_VALIDATION: "CODE_VALIDATION",
  CODE_DEPENDENCY: "CODE_DEPENDENCY",
  CODE_OLAP: "CODE_OLAP",
  CODE_SOURCE: "CODE_SOURCE",
} as const;

/**
 * ReconcileError represents an error encountered while running Reconcile.
 */
export interface V1ReconcileError {
  code?: V1ReconcileErrorCode;
  endLocation?: ReconcileErrorCharLocation;
  filePath?: string;
  message?: string;
  /** Property path of the error in the code artifact (if any).
It's represented as a JS-style property path, e.g. "key0.key1[index2].key3".
It only applies to structured code artifacts (i.e. YAML).
Only applicable if file_path is set. */
  propertyPath?: string[];
  startLocation?: ReconcileErrorCharLocation;
}

export interface V1RefreshAndReconcileResponse {
  /** affected_paths lists all the file artifact paths that were considered while
executing the reconciliation. If changed_paths was empty, this will include all
code artifacts in the repo. */
  affectedPaths?: string[];
  /** Errors encountered during reconciliation. If strict = false, any path in
affected_paths without an error can be assumed to have been reconciled succesfully. */
  errors?: V1ReconcileError[];
}

export interface V1ReconcileResponse {
  /** affected_paths lists all the file artifact paths that were considered while
executing the reconciliation. If changed_paths was empty, this will include all
code artifacts in the repo. */
  affectedPaths?: string[];
  /** Errors encountered during reconciliation. If strict = false, any path in
affected_paths without an error can be assumed to have been reconciled succesfully. */
  errors?: V1ReconcileError[];
}

export type V1QueryResponseDataItem = { [key: string]: any };

export interface V1QueryResponse {
  data?: V1QueryResponseDataItem[];
  meta?: V1StructType;
}

export interface V1PutFileResponse {
  filePath?: string;
}

export interface V1PutFileAndReconcileResponse {
  /** affected_paths lists all the file artifact paths that were considered while
executing the reconciliation. If changed_paths was empty, this will include all
code artifacts in the repo. */
  affectedPaths?: string[];
  /** Errors encountered during reconciliation. If strict = false, any path in
affected_paths without an error can be assumed to have been reconciled succesfully. */
  errors?: V1ReconcileError[];
}

export interface V1PutFileAndReconcileRequest {
  blob?: string;
  create?: boolean;
  /** create_only will cause the operation to fail if a file already exists at path.
It should only be set when create = true. */
  createOnly?: boolean;
  /** If true, will save the file and validate it and related file artifacts, but not actually execute any migrations. */
  dry?: boolean;
  instanceId?: string;
  path?: string;
  strict?: boolean;
}

export interface V1ProfileColumn {
  largestStringLength?: number;
  name?: string;
  type?: string;
}

export interface V1ProfileColumnsResponse {
  profileColumns?: V1ProfileColumn[];
}

export interface V1PingResponse {
  time?: string;
  version?: string;
}

export type V1ObjectType = typeof V1ObjectType[keyof typeof V1ObjectType];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1ObjectType = {
  OBJECT_TYPE_UNSPECIFIED: "OBJECT_TYPE_UNSPECIFIED",
  OBJECT_TYPE_TABLE: "OBJECT_TYPE_TABLE",
  OBJECT_TYPE_SOURCE: "OBJECT_TYPE_SOURCE",
  OBJECT_TYPE_MODEL: "OBJECT_TYPE_MODEL",
  OBJECT_TYPE_METRICS_VIEW: "OBJECT_TYPE_METRICS_VIEW",
} as const;

export interface V1NumericStatistics {
  max?: number;
  mean?: number;
  min?: number;
  q25?: number;
  q50?: number;
  q75?: number;
  sd?: number;
}

export interface V1NumericOutliers {
  outliers?: NumericOutliersOutlier[];
}

export interface V1NumericHistogramBins {
  bins?: NumericHistogramBinsBin[];
}

/**
 * Response for RuntimeService.GetNumericHistogram, RuntimeService.GetDescriptiveStatistics and RuntimeService.GetCardinalityOfColumn.
Message will have either numericHistogramBins, numericStatistics or numericOutliers set.
 */
export interface V1NumericSummary {
  numericHistogramBins?: V1NumericHistogramBins;
  numericOutliers?: V1NumericOutliers;
  numericStatistics?: V1NumericStatistics;
}

export interface V1Model {
  dialect?: ModelDialect;
  name?: string;
  schema?: V1StructType;
  sql?: string;
}

export type V1MetricsViewTotalsResponseData = { [key: string]: any };

export interface V1MetricsViewTotalsResponse {
  data?: V1MetricsViewTotalsResponseData;
  meta?: V1MetricsViewColumn[];
}

export type V1MetricsViewToplistResponseDataItem = { [key: string]: any };

export type V1MetricsViewTimeSeriesResponseDataItem = { [key: string]: any };

export interface V1MetricsViewTimeSeriesResponse {
  data?: V1MetricsViewTimeSeriesResponseDataItem[];
  meta?: V1MetricsViewColumn[];
}

export interface V1MetricsViewSort {
  ascending?: boolean;
  name?: string;
}

export interface V1MetricsViewFilter {
  exclude?: MetricsViewFilterCond[];
  include?: MetricsViewFilterCond[];
  match?: string[];
}

export interface V1MetricsViewDimensionValue {
  in?: unknown[];
  like?: unknown[];
  name?: string;
}

export interface V1MetricsViewRequestFilter {
  exclude?: V1MetricsViewDimensionValue[];
  include?: V1MetricsViewDimensionValue[];
}

export interface V1MetricsViewColumn {
  name?: string;
  nullable?: boolean;
  type?: string;
}

export interface V1MetricsViewToplistResponse {
  data?: V1MetricsViewToplistResponseDataItem[];
  meta?: V1MetricsViewColumn[];
}

export interface V1MetricsView {
  description?: string;
  dimensions?: MetricsViewDimension[];
  from?: string;
  label?: string;
  measures?: MetricsViewMeasure[];
  name?: string;
  timeDimension?: string;
  /** Recommended granularities for rolling up the time dimension.
Should be a valid SQL INTERVAL value. */
  timeGrains?: string[];
}

export interface V1MapType {
  keyType?: Runtimev1Type;
  valueType?: Runtimev1Type;
}

export interface V1ListInstancesResponse {
  instances?: V1Instance[];
  nextPageToken?: string;
}

export interface V1ListFilesResponse {
  paths?: string[];
}

export interface V1ListConnectorsResponse {
  connectors?: V1Connector[];
}

export interface V1ListCatalogEntriesResponse {
  entries?: V1CatalogEntry[];
}

/**
 * Instance represents a single data project, meaning one set of code artifacts,
one connection to an OLAP datastore (DuckDB, Druid), and one catalog of related
metadata (such as reconciliation state). Instances are the unit of isolation within
the runtime. They enable one runtime deployment to serve not only multiple data
projects, but also multiple tenants. On local, the runtime will usually have
just a single instance.
 */
export interface V1Instance {
  /** If true, the runtime will store the instance's catalog in its OLAP store instead
of in the runtime's metadata store. Currently only supported for the duckdb driver. */
  embedCatalog?: boolean;
  instanceId?: string;
  olapDriver?: string;
  olapDsn?: string;
  /** Driver for reading/editing code artifacts (options: file, metastore).
This enables virtualizing a file system in a cloud setting. */
  repoDriver?: string;
  repoDsn?: string;
}

export interface V1GetTopKResponse {
  categoricalSummary?: V1CategoricalSummary;
}

export interface V1GetTimeRangeSummaryResponse {
  timeRangeSummary?: V1TimeRangeSummary;
}

export type V1GetTableRowsResponseDataItem = { [key: string]: any };

export interface V1GetTableRowsResponse {
  data?: V1GetTableRowsResponseDataItem[];
}

export interface V1GetTableCardinalityResponse {
  cardinality?: string;
}

export interface V1GetRugHistogramResponse {
  numericSummary?: V1NumericSummary;
}

export interface V1GetNumericHistogramResponse {
  numericSummary?: V1NumericSummary;
}

export interface V1GetNullCountResponse {
  count?: number;
}

export interface V1GetInstanceResponse {
  instance?: V1Instance;
}

export interface V1GetFileResponse {
  blob?: string;
  updatedOn?: string;
}

export interface V1GetDescriptiveStatisticsResponse {
  numericSummary?: V1NumericSummary;
}

export interface V1GetCatalogEntryResponse {
  entry?: V1CatalogEntry;
}

export interface V1GetCardinalityOfColumnResponse {
  categoricalSummary?: V1CategoricalSummary;
}

export interface V1GenerateTimeSeriesResponse {
  rollup?: V1TimeSeriesResponse;
}

export interface V1EstimateSmallestTimeGrainResponse {
  timeGrain?: V1TimeGrain;
}

export interface V1EstimateRollupIntervalResponse {
  end?: string;
  interval?: V1TimeGrain;
  start?: string;
}

export interface V1DeleteInstanceResponse {
  [key: string]: any;
}

export interface V1DeleteFileResponse {
  [key: string]: any;
}

export interface V1DeleteFileAndReconcileResponse {
  /** affected_paths lists all the file artifact paths that were considered while
executing the reconciliation. If changed_paths was empty, this will include all
code artifacts in the repo. */
  affectedPaths?: string[];
  /** Errors encountered during reconciliation. If strict = false, any path in
affected_paths without an error can be assumed to have been reconciled succesfully. */
  errors?: V1ReconcileError[];
}

export interface V1DeleteFileAndReconcileRequest {
  /** If true, will save the file and validate it and related file artifacts, but not actually execute any migrations. */
  dry?: boolean;
  instanceId?: string;
  path?: string;
  strict?: boolean;
}

export interface V1CreateInstanceResponse {
  instance?: V1Instance;
}

/**
 * Request message for RuntimeService.CreateInstance.
See message Instance for field descriptions.
 */
export interface V1CreateInstanceRequest {
  embedCatalog?: boolean;
  instanceId?: string;
  olapDriver?: string;
  olapDsn?: string;
  repoDriver?: string;
  repoDsn?: string;
}

/**
 * Connector represents a connector available in the runtime.
It should not be confused with a source.
 */
export interface V1Connector {
  description?: string;
  displayName?: string;
  name?: string;
  properties?: ConnectorProperty[];
}

/**
 * Response for RuntimeService.GetTopK and RuntimeService.GetCardinalityOfColumn. Message will have either topK or cardinality set.
 */
export interface V1CategoricalSummary {
  cardinality?: number;
  topK?: V1TopK;
}

export interface V1CatalogEntry {
  createdOn?: string;
  metricsView?: V1MetricsView;
  model?: V1Model;
  name?: string;
  path?: string;
  refreshedOn?: string;
  source?: V1Source;
  table?: V1Table;
  updatedOn?: string;
}

export interface Runtimev1Type {
  arrayElementType?: Runtimev1Type;
  code?: V1TypeCode;
  mapType?: V1MapType;
  nullable?: boolean;
  structType?: V1StructType;
}

export interface RpcStatus {
  code?: number;
  details?: ProtobufAny[];
  message?: string;
}

/**
 * `NullValue` is a singleton enumeration to represent the null value for the
`Value` type union.

 The JSON representation for `NullValue` is JSON `null`.

 - NULL_VALUE: Null value.
 */
export type ProtobufNullValue =
  typeof ProtobufNullValue[keyof typeof ProtobufNullValue];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const ProtobufNullValue = {
  NULL_VALUE: "NULL_VALUE",
} as const;

export interface ProtobufAny {
  "@type"?: string;
  [key: string]: unknown;
}

export interface TopKEntry {
  count?: number;
  value?: unknown;
}

export interface TimeRangeSummaryInterval {
  days?: number;
  micros?: string;
  months?: number;
}

export interface StructTypeField {
  name?: string;
  type?: Runtimev1Type;
}

export interface ReconcileErrorCharLocation {
  column?: number;
  line?: number;
}

export interface NumericOutliersOutlier {
  bucket?: number;
  count?: number;
  high?: number;
  low?: number;
  present?: boolean;
}

export interface NumericHistogramBinsBin {
  bucket?: number;
  count?: number;
  high?: number;
  low?: number;
}

export type ModelDialect = typeof ModelDialect[keyof typeof ModelDialect];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const ModelDialect = {
  DIALECT_UNSPECIFIED: "DIALECT_UNSPECIFIED",
  DIALECT_DUCKDB: "DIALECT_DUCKDB",
} as const;

export interface MetricsViewMeasure {
  description?: string;
  expression?: string;
  format?: string;
  label?: string;
  name?: string;
}

export interface MetricsViewFilterCond {
  in?: unknown[];
  like?: unknown[];
  name?: string;
}

export interface MetricsViewDimension {
  description?: string;
  label?: string;
  name?: string;
}

export interface GenerateTimeSeriesRequestBasicMeasure {
  expression?: string;
  id?: string;
  sqlName?: string;
}

export type ConnectorPropertyType =
  typeof ConnectorPropertyType[keyof typeof ConnectorPropertyType];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const ConnectorPropertyType = {
  TYPE_UNSPECIFIED: "TYPE_UNSPECIFIED",
  TYPE_STRING: "TYPE_STRING",
  TYPE_NUMBER: "TYPE_NUMBER",
  TYPE_BOOLEAN: "TYPE_BOOLEAN",
  TYPE_INFORMATIONAL: "TYPE_INFORMATIONAL",
} as const;

export interface ConnectorProperty {
  description?: string;
  displayName?: string;
  hint?: string;
  href?: string;
  key?: string;
  nullable?: boolean;
  placeholder?: string;
  type?: ConnectorPropertyType;
}
