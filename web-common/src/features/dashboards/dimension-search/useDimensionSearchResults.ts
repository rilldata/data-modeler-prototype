import { createBatches } from "@rilldata/utils";
import {
  createQueryServiceMetricsViewSearch,
  type V1MetricsViewSpec,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";

export type DimensionSearchResult = {
  dimension: string;
  values: any[];
};
export type DimensionSearchResults = {
  responses: (DimensionSearchResult | undefined)[];
  errors: Error[];
  completed: boolean;
  progress: number;
};

const BatchSize = 5;

export function useDimensionSearchResults(
  instanceId: string,
  metricsViewName: string,
  metricsView: V1MetricsViewSpec,
  timeRangeSummary: V1TimeRangeSummary,
  searchText: string,
) {
  const dimensions = metricsView.dimensions ?? [];
  const batches = createBatches(dimensions, BatchSize);
  return derived(
    batches.map((batch) =>
      createQueryServiceMetricsViewSearch(instanceId, metricsViewName, {
        dimensions: batch.map((d) => d.name ?? ""),
        search: searchText,
        limit: 100,
        timeRange: {
          start: timeRangeSummary.min,
          end: timeRangeSummary.max,
        },
      }),
    ),
    (searchResults) => {
      const results: DimensionSearchResults = {
        responses: [],
        errors: [],
        completed: false,
        progress: 0,
      };
      const dimensionResultsMap = new Map<string, DimensionSearchResult>();

      let completedCount = 0;
      searchResults.forEach((searchResult, index) => {
        if (searchResult.error) {
          results.errors.push(
            new Error(
              searchResult.error.response?.data?.message ??
                searchResult.error.message ??
                "Unknown error",
            ),
          );
        } else if (searchResult.data?.results) {
          searchResult.data.results.forEach((dr) => {
            const dim = dr.dimension ?? "";
            if (!dimensionResultsMap.has(dim)) {
              const dsr = {
                dimension: dim,
                values: [],
              };
              dimensionResultsMap.set(dim, dsr);
              results.responses.push(dsr);
            }

            dimensionResultsMap.get(dim)?.values.push(dr.value);
          });
        }
        if (!searchResult.isFetching) {
          completedCount += batches[index].length;
        }
      });

      results.completed = completedCount === dimensions.length;
      results.progress = Math.round((completedCount * 100) / dimensions.length);

      return results;
    },
  );
}
