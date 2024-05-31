import {
  MetricsViewSpecDimensionSelector,
  MetricsViewSpecMeasureV2,
  V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

export const allMeasures = ({
  metricsSpecQueryResult,
}: DashboardDataSources): MetricsViewSpecMeasureV2[] => {
  const measures = metricsSpecQueryResult.data?.measures;
  return measures === undefined ? [] : measures;
};

export const visibleMeasures = ({
  metricsSpecQueryResult,
  dashboard,
}: DashboardDataSources): MetricsViewSpecMeasureV2[] => {
  const measures = metricsSpecQueryResult.data?.measures?.filter(
    (d) => d.name && dashboard.visibleMeasureKeys.has(d.name),
  );
  return measures === undefined ? [] : measures;
};

export const getMeasureByName = (
  dashData: DashboardDataSources,
): ((name: string) => MetricsViewSpecMeasureV2 | undefined) => {
  return (name: string) => {
    return allMeasures(dashData)?.find((measure) => measure.name === name);
  };
};

export const measureLabel = ({
  metricsSpecQueryResult,
}: DashboardDataSources): ((m: string) => string) => {
  return (measureName) => {
    const measure = metricsSpecQueryResult.data?.measures?.find(
      (d) => d.name === measureName,
    );
    return measure?.label ?? measureName;
  };
};
export const isMeasureValidPercentOfTotal = ({
  metricsSpecQueryResult,
}: DashboardDataSources): ((measureName: string) => boolean) => {
  return (measureName: string) => {
    const measures = metricsSpecQueryResult.data?.measures;
    const selectedMeasure = measures?.find((m) => m.name === measureName);
    return selectedMeasure?.validPercentOfTotal ?? false;
  };
};

export const filterBasicMeasures = (
  measures: MetricsViewSpecMeasureV2[] | undefined,
) => measures?.filter((m) => !m.window) ?? [];

export const getMeasuresAndDimensions = ({
  dashboard,
}: Pick<DashboardDataSources, "dashboard">) => {
  return (
    metricsViewSpec: V1MetricsViewSpec,
    measureNames: string[],
  ): {
    measures: string[];
    dimensions: MetricsViewSpecDimensionSelector[];
  } => {
    const dimensions = new Map<string, V1TimeGrain>();
    const measures = new Set<string>();
    measureNames.forEach((measureName) => {
      const measure = metricsViewSpec.measures?.find(
        (m) => m.name === measureName,
      );
      if (!measure) return;

      let skipMeasure = false;
      measure.requiredDimensions?.forEach((reqDim) => {
        if (
          reqDim.timeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED &&
          reqDim.timeGrain !== dashboard.selectedTimeRange?.interval
        ) {
          // filter out measures with dependant dimensions not matching the selected grain
          skipMeasure = true;
          return;
        }
        if (!reqDim.name) return;

        const existingEntry = dimensions.get(reqDim.name);
        if (existingEntry) {
          if (existingEntry === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) {
            dimensions.set(
              reqDim.name,
              reqDim.timeGrain ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
            );
          } else {
            // mismatching measures are requested
            skipMeasure = true;
          }
          return;
        }

        dimensions.set(
          reqDim.name,
          reqDim.timeGrain ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        );
      });
      if (skipMeasure) return;
      measures.add(measureName);
    });
    return {
      measures: [...measures],
      dimensions: [...dimensions.entries()].map(([name, timeGrain]) => ({
        name,
        timeGrain,
      })),
    };
  };
};

export const measureSelectors = {
  /**
   * Get all measures in the dashboard.
   */
  allMeasures,

  /**
   * Returns a function that can be used to get a MetricsViewSpecMeasureV2
   * by name; this fn returns undefined if the dashboard has no measure with that name.
   */
  getMeasureByName,

  /**
   * Gets all visible measures in the dashboard.
   */
  visibleMeasures,
  /**
   * Get label for a measure by name
   */
  measureLabel,
  /**
   * Checks if the provided measure is a valid percent of total
   */
  isMeasureValidPercentOfTotal,
};
