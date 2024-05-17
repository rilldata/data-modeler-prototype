/**
 * Generated by orval v6.12.0 🍺
 * Do not edit manually.
 * rill/runtime/v1/colors.proto
 * OpenAPI spec version: version not set
 */
import { createQuery } from "@tanstack/svelte-query";
import type {
  CreateQueryOptions,
  QueryFunction,
  CreateQueryResult,
  QueryKey,
} from "@tanstack/svelte-query";
import type {
  V1BigQueryListDatasetsResponse,
  RpcStatus,
  ConnectorServiceBigQueryListDatasetsParams,
  V1BigQueryListTablesResponse,
  ConnectorServiceBigQueryListTablesParams,
  V1OLAPGetTableResponse,
  ConnectorServiceOLAPGetTableParams,
  V1GCSListObjectsResponse,
  ConnectorServiceGCSListObjectsParams,
  V1GCSListBucketsResponse,
  ConnectorServiceGCSListBucketsParams,
  V1GCSGetCredentialsInfoResponse,
  ConnectorServiceGCSGetCredentialsInfoParams,
  V1OLAPListTablesResponse,
  ConnectorServiceOLAPListTablesParams,
  V1S3GetBucketMetadataResponse,
  ConnectorServiceS3GetBucketMetadataParams,
  V1S3ListObjectsResponse,
  ConnectorServiceS3ListObjectsParams,
  V1S3ListBucketsResponse,
  ConnectorServiceS3ListBucketsParams,
  V1S3GetCredentialsInfoResponse,
  ConnectorServiceS3GetCredentialsInfoParams,
} from "../index.schemas";
import { httpClient } from "../../http-client";
import type { ErrorType } from "../../http-client";

type AwaitedInput<T> = PromiseLike<T> | T;

type Awaited<O> = O extends AwaitedInput<infer T> ? T : never;

/**
 * @summary BigQueryListDatasets list all datasets in a bigquery project
 */
export const connectorServiceBigQueryListDatasets = (
  params?: ConnectorServiceBigQueryListDatasetsParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1BigQueryListDatasetsResponse>({
    url: `/v1/bigquery/datasets`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceBigQueryListDatasetsQueryKey = (
  params?: ConnectorServiceBigQueryListDatasetsParams,
) => [`/v1/bigquery/datasets`, ...(params ? [params] : [])];

export type ConnectorServiceBigQueryListDatasetsQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceBigQueryListDatasets>>
>;
export type ConnectorServiceBigQueryListDatasetsQueryError =
  ErrorType<RpcStatus>;

export const createConnectorServiceBigQueryListDatasets = <
  TData = Awaited<ReturnType<typeof connectorServiceBigQueryListDatasets>>,
  TError = ErrorType<RpcStatus>,
>(
  params?: ConnectorServiceBigQueryListDatasetsParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceBigQueryListDatasets>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getConnectorServiceBigQueryListDatasetsQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceBigQueryListDatasets>>
  > = ({ signal }) => connectorServiceBigQueryListDatasets(params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceBigQueryListDatasets>>,
    TError,
    TData
  >({ queryKey, queryFn, ...queryOptions }) as CreateQueryResult<
    TData,
    TError
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary BigQueryListTables list all tables in a bigquery project:dataset
 */
export const connectorServiceBigQueryListTables = (
  params?: ConnectorServiceBigQueryListTablesParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1BigQueryListTablesResponse>({
    url: `/v1/bigquery/tables`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceBigQueryListTablesQueryKey = (
  params?: ConnectorServiceBigQueryListTablesParams,
) => [`/v1/bigquery/tables`, ...(params ? [params] : [])];

export type ConnectorServiceBigQueryListTablesQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceBigQueryListTables>>
>;
export type ConnectorServiceBigQueryListTablesQueryError = ErrorType<RpcStatus>;

export const createConnectorServiceBigQueryListTables = <
  TData = Awaited<ReturnType<typeof connectorServiceBigQueryListTables>>,
  TError = ErrorType<RpcStatus>,
>(
  params?: ConnectorServiceBigQueryListTablesParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceBigQueryListTables>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getConnectorServiceBigQueryListTablesQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceBigQueryListTables>>
  > = ({ signal }) => connectorServiceBigQueryListTables(params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceBigQueryListTables>>,
    TError,
    TData
  >({ queryKey, queryFn, ...queryOptions }) as CreateQueryResult<
    TData,
    TError
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary OLAPGetTable returns metadata about a table or view in an OLAP
 */
export const connectorServiceOLAPGetTable = (
  params?: ConnectorServiceOLAPGetTableParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1OLAPGetTableResponse>({
    url: `/v1/connectors/olap/table`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceOLAPGetTableQueryKey = (
  params?: ConnectorServiceOLAPGetTableParams,
) => [`/v1/connectors/olap/table`, ...(params ? [params] : [])];

export type ConnectorServiceOLAPGetTableQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceOLAPGetTable>>
>;
export type ConnectorServiceOLAPGetTableQueryError = ErrorType<RpcStatus>;

export const createConnectorServiceOLAPGetTable = <
  TData = Awaited<ReturnType<typeof connectorServiceOLAPGetTable>>,
  TError = ErrorType<RpcStatus>,
>(
  params?: ConnectorServiceOLAPGetTableParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceOLAPGetTable>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getConnectorServiceOLAPGetTableQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceOLAPGetTable>>
  > = ({ signal }) => connectorServiceOLAPGetTable(params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceOLAPGetTable>>,
    TError,
    TData
  >({ queryKey, queryFn, ...queryOptions }) as CreateQueryResult<
    TData,
    TError
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary GCSListObjects lists objects for the given bucket.
 */
export const connectorServiceGCSListObjects = (
  bucket: string,
  params?: ConnectorServiceGCSListObjectsParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1GCSListObjectsResponse>({
    url: `/v1/gcs/bucket/${bucket}/objects`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceGCSListObjectsQueryKey = (
  bucket: string,
  params?: ConnectorServiceGCSListObjectsParams,
) => [`/v1/gcs/bucket/${bucket}/objects`, ...(params ? [params] : [])];

export type ConnectorServiceGCSListObjectsQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceGCSListObjects>>
>;
export type ConnectorServiceGCSListObjectsQueryError = ErrorType<RpcStatus>;

export const createConnectorServiceGCSListObjects = <
  TData = Awaited<ReturnType<typeof connectorServiceGCSListObjects>>,
  TError = ErrorType<RpcStatus>,
>(
  bucket: string,
  params?: ConnectorServiceGCSListObjectsParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceGCSListObjects>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getConnectorServiceGCSListObjectsQueryKey(bucket, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceGCSListObjects>>
  > = ({ signal }) => connectorServiceGCSListObjects(bucket, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceGCSListObjects>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!bucket,
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary GCSListBuckets lists buckets accessible with the configured credentials.
 */
export const connectorServiceGCSListBuckets = (
  params?: ConnectorServiceGCSListBucketsParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1GCSListBucketsResponse>({
    url: `/v1/gcs/buckets`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceGCSListBucketsQueryKey = (
  params?: ConnectorServiceGCSListBucketsParams,
) => [`/v1/gcs/buckets`, ...(params ? [params] : [])];

export type ConnectorServiceGCSListBucketsQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceGCSListBuckets>>
>;
export type ConnectorServiceGCSListBucketsQueryError = ErrorType<RpcStatus>;

export const createConnectorServiceGCSListBuckets = <
  TData = Awaited<ReturnType<typeof connectorServiceGCSListBuckets>>,
  TError = ErrorType<RpcStatus>,
>(
  params?: ConnectorServiceGCSListBucketsParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceGCSListBuckets>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getConnectorServiceGCSListBucketsQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceGCSListBuckets>>
  > = ({ signal }) => connectorServiceGCSListBuckets(params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceGCSListBuckets>>,
    TError,
    TData
  >({ queryKey, queryFn, ...queryOptions }) as CreateQueryResult<
    TData,
    TError
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary GCSGetCredentialsInfo returns metadata for the given bucket.
 */
export const connectorServiceGCSGetCredentialsInfo = (
  params?: ConnectorServiceGCSGetCredentialsInfoParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1GCSGetCredentialsInfoResponse>({
    url: `/v1/gcs/credentials_info`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceGCSGetCredentialsInfoQueryKey = (
  params?: ConnectorServiceGCSGetCredentialsInfoParams,
) => [`/v1/gcs/credentials_info`, ...(params ? [params] : [])];

export type ConnectorServiceGCSGetCredentialsInfoQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceGCSGetCredentialsInfo>>
>;
export type ConnectorServiceGCSGetCredentialsInfoQueryError =
  ErrorType<RpcStatus>;

export const createConnectorServiceGCSGetCredentialsInfo = <
  TData = Awaited<ReturnType<typeof connectorServiceGCSGetCredentialsInfo>>,
  TError = ErrorType<RpcStatus>,
>(
  params?: ConnectorServiceGCSGetCredentialsInfoParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceGCSGetCredentialsInfo>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getConnectorServiceGCSGetCredentialsInfoQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceGCSGetCredentialsInfo>>
  > = ({ signal }) => connectorServiceGCSGetCredentialsInfo(params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceGCSGetCredentialsInfo>>,
    TError,
    TData
  >({ queryKey, queryFn, ...queryOptions }) as CreateQueryResult<
    TData,
    TError
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary OLAPListTables list all tables across all databases on motherduck
 */
export const connectorServiceOLAPListTables = (
  params?: ConnectorServiceOLAPListTablesParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1OLAPListTablesResponse>({
    url: `/v1/olap/tables`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceOLAPListTablesQueryKey = (
  params?: ConnectorServiceOLAPListTablesParams,
) => [`/v1/olap/tables`, ...(params ? [params] : [])];

export type ConnectorServiceOLAPListTablesQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceOLAPListTables>>
>;
export type ConnectorServiceOLAPListTablesQueryError = ErrorType<RpcStatus>;

export const createConnectorServiceOLAPListTables = <
  TData = Awaited<ReturnType<typeof connectorServiceOLAPListTables>>,
  TError = ErrorType<RpcStatus>,
>(
  params?: ConnectorServiceOLAPListTablesParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceOLAPListTables>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getConnectorServiceOLAPListTablesQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceOLAPListTables>>
  > = ({ signal }) => connectorServiceOLAPListTables(params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceOLAPListTables>>,
    TError,
    TData
  >({ queryKey, queryFn, ...queryOptions }) as CreateQueryResult<
    TData,
    TError
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary S3GetBucketMetadata returns metadata for the given bucket.
 */
export const connectorServiceS3GetBucketMetadata = (
  bucket: string,
  params?: ConnectorServiceS3GetBucketMetadataParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1S3GetBucketMetadataResponse>({
    url: `/v1/s3/bucket/${bucket}/metadata`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceS3GetBucketMetadataQueryKey = (
  bucket: string,
  params?: ConnectorServiceS3GetBucketMetadataParams,
) => [`/v1/s3/bucket/${bucket}/metadata`, ...(params ? [params] : [])];

export type ConnectorServiceS3GetBucketMetadataQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceS3GetBucketMetadata>>
>;
export type ConnectorServiceS3GetBucketMetadataQueryError =
  ErrorType<RpcStatus>;

export const createConnectorServiceS3GetBucketMetadata = <
  TData = Awaited<ReturnType<typeof connectorServiceS3GetBucketMetadata>>,
  TError = ErrorType<RpcStatus>,
>(
  bucket: string,
  params?: ConnectorServiceS3GetBucketMetadataParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceS3GetBucketMetadata>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getConnectorServiceS3GetBucketMetadataQueryKey(bucket, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceS3GetBucketMetadata>>
  > = ({ signal }) =>
    connectorServiceS3GetBucketMetadata(bucket, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceS3GetBucketMetadata>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!bucket,
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary S3ListBuckets lists objects for the given bucket.
 */
export const connectorServiceS3ListObjects = (
  bucket: string,
  params?: ConnectorServiceS3ListObjectsParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1S3ListObjectsResponse>({
    url: `/v1/s3/bucket/${bucket}/objects`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceS3ListObjectsQueryKey = (
  bucket: string,
  params?: ConnectorServiceS3ListObjectsParams,
) => [`/v1/s3/bucket/${bucket}/objects`, ...(params ? [params] : [])];

export type ConnectorServiceS3ListObjectsQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceS3ListObjects>>
>;
export type ConnectorServiceS3ListObjectsQueryError = ErrorType<RpcStatus>;

export const createConnectorServiceS3ListObjects = <
  TData = Awaited<ReturnType<typeof connectorServiceS3ListObjects>>,
  TError = ErrorType<RpcStatus>,
>(
  bucket: string,
  params?: ConnectorServiceS3ListObjectsParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceS3ListObjects>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getConnectorServiceS3ListObjectsQueryKey(bucket, params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceS3ListObjects>>
  > = ({ signal }) => connectorServiceS3ListObjects(bucket, params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceS3ListObjects>>,
    TError,
    TData
  >({
    queryKey,
    queryFn,
    enabled: !!bucket,
    ...queryOptions,
  }) as CreateQueryResult<TData, TError> & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary S3ListBuckets lists buckets accessible with the configured credentials.
 */
export const connectorServiceS3ListBuckets = (
  params?: ConnectorServiceS3ListBucketsParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1S3ListBucketsResponse>({
    url: `/v1/s3/buckets`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceS3ListBucketsQueryKey = (
  params?: ConnectorServiceS3ListBucketsParams,
) => [`/v1/s3/buckets`, ...(params ? [params] : [])];

export type ConnectorServiceS3ListBucketsQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceS3ListBuckets>>
>;
export type ConnectorServiceS3ListBucketsQueryError = ErrorType<RpcStatus>;

export const createConnectorServiceS3ListBuckets = <
  TData = Awaited<ReturnType<typeof connectorServiceS3ListBuckets>>,
  TError = ErrorType<RpcStatus>,
>(
  params?: ConnectorServiceS3ListBucketsParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceS3ListBuckets>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getConnectorServiceS3ListBucketsQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceS3ListBuckets>>
  > = ({ signal }) => connectorServiceS3ListBuckets(params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceS3ListBuckets>>,
    TError,
    TData
  >({ queryKey, queryFn, ...queryOptions }) as CreateQueryResult<
    TData,
    TError
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};

/**
 * @summary S3GetCredentialsInfo returns metadata for the given bucket.
 */
export const connectorServiceS3GetCredentialsInfo = (
  params?: ConnectorServiceS3GetCredentialsInfoParams,
  signal?: AbortSignal,
) => {
  return httpClient<V1S3GetCredentialsInfoResponse>({
    url: `/v1/s3/credentials_info`,
    method: "get",
    params,
    signal,
  });
};

export const getConnectorServiceS3GetCredentialsInfoQueryKey = (
  params?: ConnectorServiceS3GetCredentialsInfoParams,
) => [`/v1/s3/credentials_info`, ...(params ? [params] : [])];

export type ConnectorServiceS3GetCredentialsInfoQueryResult = NonNullable<
  Awaited<ReturnType<typeof connectorServiceS3GetCredentialsInfo>>
>;
export type ConnectorServiceS3GetCredentialsInfoQueryError =
  ErrorType<RpcStatus>;

export const createConnectorServiceS3GetCredentialsInfo = <
  TData = Awaited<ReturnType<typeof connectorServiceS3GetCredentialsInfo>>,
  TError = ErrorType<RpcStatus>,
>(
  params?: ConnectorServiceS3GetCredentialsInfoParams,
  options?: {
    query?: CreateQueryOptions<
      Awaited<ReturnType<typeof connectorServiceS3GetCredentialsInfo>>,
      TError,
      TData
    >;
  },
): CreateQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ??
    getConnectorServiceS3GetCredentialsInfoQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof connectorServiceS3GetCredentialsInfo>>
  > = ({ signal }) => connectorServiceS3GetCredentialsInfo(params, signal);

  const query = createQuery<
    Awaited<ReturnType<typeof connectorServiceS3GetCredentialsInfo>>,
    TError,
    TData
  >({ queryKey, queryFn, ...queryOptions }) as CreateQueryResult<
    TData,
    TError
  > & { queryKey: QueryKey };

  query.queryKey = queryKey;

  return query;
};
