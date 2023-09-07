import {
  loadColumnCardinality,
  loadColumnsNullCount,
  loadColumnTopK,
  loadTableCardinality,
} from "@rilldata/web-common/components/column-profile/column-types/common-data";
import {
  loadColumnHistogram,
  loadDescriptiveStatistics,
} from "@rilldata/web-common/components/column-profile/column-types/numeric-profile-data";
import { loadTimeSeries } from "@rilldata/web-common/components/column-profile/column-types/timestamp-profile-data";
import { loadReferencesProfile } from "@rilldata/web-common/components/column-profile/references-profile-data";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import { createThrottler } from "@rilldata/web-common/lib/create-throttler";
import {
  CATEGORICALS,
  INTERVALS,
  isFloat,
  isNested,
  NUMERICS,
  TIMESTAMPS,
} from "@rilldata/web-common/lib/duckdb-data-types";
import type {
  NumericHistogramBinsBin,
  NumericOutliersOutlier,
  TopKEntry,
  V1NumericStatistics,
  V1ProfileColumn,
  V1TableColumnsResponse,
  V1TimeGrain,
  V1TimeSeriesValue,
  V1CatalogEntry,
} from "@rilldata/web-common/runtime-client";
import { BatchedRequest } from "@rilldata/web-common/runtime-client/batched-request";
import { waitUntil } from "@rilldata/web-local/lib/util/waitUtils";
import type { QueryObserverResult } from "@tanstack/query-core";
import { getContext, setContext } from "svelte";
import { Updater, writable } from "svelte/store";
import type { Readable } from "svelte/store";

export type ColumnProfileData = {
  name: string;
  type: string;

  isFetching: boolean;
  nullCount?: number;
  cardinality?: number;

  topK?: Array<TopKEntry>;

  // numeric profile
  rugHistogram?: Array<NumericOutliersOutlier>;
  histogram?: Array<NumericHistogramBinsBin>;
  descriptiveStatistics?: V1NumericStatistics;

  // timestamp profile
  estimatedRollupInterval?: V1TimeGrain;
  smallestTimeGrain?: V1TimeGrain;
  timeSeriesData?: Array<V1TimeSeriesValue>;
  timeSeriesSpark?: Array<V1TimeSeriesValue>;
};
export type ReferenceProfileData = {
  cardinality: number;
  columns: Array<V1ProfileColumn>;
};
export type ColumnsProfileData = {
  isFetching: boolean;
  tableRows: number;
  columnNames: Array<string>;
  profiles: Record<string, ColumnProfileData>;
  references: Array<ReferenceProfileData>;
};
export type ColumnsProfileDataMethods = {
  load: (
    instanceId: string,
    tableName: string,
    profileColumnResponse: QueryObserverResult<V1TableColumnsResponse>,
    references: Array<V1CatalogEntry>
  ) => Promise<void>;
};
export type ColumnsProfileDataStore = Readable<ColumnsProfileData> &
  ColumnsProfileDataMethods;
type StoreUpdater = (state: ColumnsProfileData) => ColumnsProfileData;

export function setColumnsProfileStore(store: ColumnsProfileDataStore) {
  setContext("COLUMNS_PROFILE", store);
}

export function getColumnsProfileStore() {
  return getContext<ColumnsProfileDataStore>("COLUMNS_PROFILE");
}

export function createColumnsProfileData(): ColumnsProfileDataStore {
  const { update, subscribe } = writable<ColumnsProfileData>({
    isFetching: true,
    tableRows: 0,
    columnNames: [],
    profiles: {},
    references: [],
  });

  let batchedRequest: BatchedRequest;

  const throttler = createThrottler(500);
  let updaters = new Array<StoreUpdater>();
  const throttledUpdate = (updater: StoreUpdater) => {
    updaters.push(updater);
    throttler(() => {
      update((state) => {
        for (const up of updaters) {
          up(state);
        }
        return state;
      });
      updaters = [];
    });
  };

  return {
    subscribe,
    load: async (
      instanceId: string,
      tableName: string,
      profileColumnResponse: QueryObserverResult<V1TableColumnsResponse>,
      references: Array<V1CatalogEntry>
    ) => {
      batchedRequest?.cancel();

      resetState(profileColumnResponse, references?.length ?? 0, update);

      batchedRequest = new BatchedRequest();
      loadTableCardinality(
        instanceId,
        tableName,
        batchedRequest,
        throttledUpdate
      );

      loadReferencesProfile(
        instanceId,
        references ?? [],
        batchedRequest,
        throttledUpdate
      );

      for (const column of profileColumnResponse.data.profileColumns) {
        const columnPromises = new Array<Promise<any>>();
        const columnName = column.name;
        columnPromises.push(
          loadColumnsNullCount(
            instanceId,
            tableName,
            columnName,
            batchedRequest,
            throttledUpdate
          ),
          loadColumnCardinality(
            instanceId,
            tableName,
            columnName,
            batchedRequest,
            throttledUpdate
          )
        );

        let type = column.type;
        if (!type) continue;
        if (type.includes("DECIMAL")) type = "DECIMAL";

        if (CATEGORICALS.has(type)) {
          columnPromises.push(
            loadColumnTopK(
              instanceId,
              tableName,
              columnName,
              batchedRequest,
              throttledUpdate
            )
          );
        } else if (NUMERICS.has(type) || INTERVALS.has(type)) {
          columnPromises.push(
            loadColumnHistogram(
              instanceId,
              tableName,
              columnName,
              isFloat(type),
              batchedRequest,
              throttledUpdate
            ),
            loadDescriptiveStatistics(
              instanceId,
              tableName,
              columnName,
              batchedRequest,
              throttledUpdate
            )
          );
        } else if (TIMESTAMPS.has(type)) {
          columnPromises.push(
            loadTimeSeries(
              instanceId,
              tableName,
              columnName,
              batchedRequest,
              throttledUpdate
            )
          );
        } else if (isNested(type)) {
          columnPromises.push(
            loadColumnTopK(
              instanceId,
              tableName,
              columnName,
              batchedRequest,
              throttledUpdate
            )
          );
        }

        Promise.all(columnPromises).then(async () => {
          throttledUpdate((state) => {
            if (!state.profiles[columnName]) return;
            state.profiles[columnName].isFetching = false;
            return state;
          });
        });
      }

      return batchedRequest.send(instanceId);
    },
  };
}

export type ColumnsProfileDataUpdate = (
  this: void,
  updater: Updater<ColumnsProfileData>
) => void;

export function resetState(
  profileColumnResponse: QueryObserverResult<V1TableColumnsResponse>,
  referencesCount: number,
  update: ColumnsProfileDataUpdate
) {
  const columnsMap = getMapFromArray(
    profileColumnResponse.data.profileColumns,
    (entity) => entity.name
  );

  update((state) => {
    state.isFetching = true;

    // remove older columns
    for (const oldColumnName in state.profiles) {
      if (!columnsMap.has(oldColumnName)) {
        delete state.profiles[oldColumnName];
      }
    }

    const columnNames = new Array<string>();

    // mark everything as fetching
    for (const column of profileColumnResponse.data.profileColumns) {
      if (!(column.name in state.profiles)) {
        state.profiles[column.name] = {
          name: column.name,
          type: column.type,
          isFetching: true,
        };
      } else {
        state.profiles[column.name].isFetching = true;
      }
      columnNames.push(column.name);
    }

    state.columnNames = columnNames;
    state.references = new Array(referencesCount).fill(undefined);

    return state;
  });
}
