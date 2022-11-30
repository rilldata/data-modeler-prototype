import {
  useRuntimeServiceEstimateRollupInterval,
  useRuntimeServiceEstimateSmallestTimeGrain,
  useRuntimeServiceGenerateTimeSeries,
  useRuntimeServiceGetCardinalityOfColumn,
  useRuntimeServiceGetNullCount,
  useRuntimeServiceGetNumericHistogram,
  useRuntimeServiceGetRugHistogram,
  useRuntimeServiceGetTableCardinality,
  useRuntimeServiceGetTopK,
} from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";
import { convertTimestampPreview } from "../../util/convertTimestampPreview";

export function getNullPercentage(instanceId, objectName, columnName) {
  const nullQuery = useRuntimeServiceGetNullCount(
    instanceId,
    objectName,
    columnName
  );
  const totalRowsQuery = useRuntimeServiceGetTableCardinality(
    instanceId,
    objectName
  );
  return derived([nullQuery, totalRowsQuery], ([nulls, totalRows]) => {
    return {
      nullCount: nulls?.data?.count,
      totalRows: +totalRows?.data?.cardinality,
    };
  });
}

export function getCountDistinct(instanceId, objectName, columnName) {
  const cardinalityQuery = useRuntimeServiceGetCardinalityOfColumn(
    instanceId,
    objectName,
    columnName
  );

  const totalRowsQuery = useRuntimeServiceGetTableCardinality(
    instanceId,
    objectName
  );

  return derived(
    [cardinalityQuery, totalRowsQuery],
    ([cardinality, totalRows]) => {
      return {
        cardinality: cardinality?.data?.categoricalSummary?.cardinality,
        totalRows: +totalRows?.data?.cardinality,
      };
    }
  );
}

export function getTopK(instanceId, objectName, columnName) {
  const topKQuery = useRuntimeServiceGetTopK(
    instanceId,
    objectName,
    columnName,
    {
      agg: "count(*)",
      k: 75,
    }
  );
  return derived(topKQuery, ($topKQuery) => {
    return $topKQuery?.data?.categoricalSummary?.topK?.entries;
  });
}

export function getTimeSeriesAndSpark(instanceId, objectName, columnName) {
  const query = useRuntimeServiceGenerateTimeSeries(
    instanceId,
    // FIXME: convert pixel back to number once the API
    {
      tableName: objectName,
      timestampColumnName: columnName,
      pixels: 92,
    }
  );
  const estimatedInterval = useRuntimeServiceEstimateRollupInterval(
    instanceId,
    objectName,
    { columnName }
  );

  const smallestTimeGrain = useRuntimeServiceEstimateSmallestTimeGrain(
    instanceId,
    objectName,
    columnName
  );

  return derived(
    [query, estimatedInterval, smallestTimeGrain],
    ([$query, $estimatedInterval, $smallestTimeGrain]) => {
      return {
        estimatedRollupInterval: $estimatedInterval?.data,
        smallestTimegrain: $smallestTimeGrain?.data?.timeGrain,
        data: convertTimestampPreview(
          $query?.data?.rollup?.results?.map((di) => {
            const next = { ...di };
            next.count = next.records.count;
            return next;
          }) || [],
          "ts"
        ),
        spark: convertTimestampPreview(
          $query?.data?.rollup?.spark?.map((di) => {
            const next = { ...di };
            next.count = next.records.count;
            return next;
          }) || [],
          "ts"
        ),
      };
    }
  );
}

export function getNumericHistogram(instanceId, objectName, columnName) {
  const histogramQuery = useRuntimeServiceGetNumericHistogram(
    instanceId,
    objectName,
    columnName
  );
  return derived(histogramQuery, ($query) => {
    return $query?.data?.numericSummary?.numericHistogramBins?.bins;
  });
}

export function getRugPlotData(instanceId, objectName, columnName) {
  const outliersQuery = useRuntimeServiceGetRugHistogram(
    instanceId,
    objectName,
    columnName
  );
  return derived(outliersQuery, ($query) => {
    return $query?.data?.numericSummary?.numericOutliers?.outliers;
  });
}
