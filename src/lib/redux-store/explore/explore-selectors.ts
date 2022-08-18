import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type {
  BasicMeasureDefinition,
  MeasureDefinitionEntity,
} from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type {
  ActiveValues,
  MetricsExplorerEntity,
} from "$lib/redux-store/explore/explore-slice";
import {
  selectMeasureById,
  selectValidMeasures,
} from "$lib/redux-store/measure-definition/measure-definition-selectors";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { generateEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import { prune } from "../../../routes/_surfaces/workspace/explore/utils";
import {
  getDefaultTimeGrain,
  getDefaultTimeRangeName,
} from "../../../routes/_surfaces/workspace/explore/time-controls/time-range-utils";
import type { TimeRangeName } from "$common/database-service/DatabaseTimeSeriesActions";

export const {
  manySelector: selectMetricsExplorers,
  singleSelector: selectMetricsExplorerById,
} = generateEntitySelectors<MetricsExplorerEntity, "metricsExplorer">(
  "metricsExplorer"
);

/**
 * Common code to fetch metrics explore and some params from state
 * Fetches 'measures' from state if not passed.
 * Fetches 'filters' from state if not passed. Also prunes empty filters.
 */
export const selectMetricsExplorerParams = (
  state: RillReduxState,
  id: string,
  {
    measures,
    filters,
    dimensions,
  }: {
    measures?: Array<MeasureDefinitionEntity>;
    filters?: ActiveValues;
    dimensions: Record<string, DimensionDefinitionEntity>;
  }
) => {
  const metricsExplorer = selectMetricsExplorerById(state, id);

  if (!filters) {
    filters = metricsExplorer.activeValues;
  }
  filters = prune(filters, dimensions);

  if (!measures) {
    measures = metricsExplorer.selectedMeasureIds.map((measureId) =>
      selectMeasureById(state, measureId)
    );
  }
  measures = selectValidMeasures(measures);

  return {
    metricsExplorer,
    prunedFilters: filters,
    normalisedMeasures: measures.map(
      (measure) =>
        ({
          id: measure.id,
          expression: measure.expression,
          sqlName: measure.sqlName,
        } as BasicMeasureDefinition)
    ),
  };
};

/**
 * Selects time range name from metrics explorer.
 * If there is none selected or if the selected time range name is not in the selectable ranges
 * returns {@link getDefaultTimeRangeName}()
 */
export const selectMetricsExploreSelectedTimeRangeName = (
  state: RillReduxState,
  id: string
) => {
  const metricsExplorer = selectMetricsExplorerById(state, id);
  if (metricsExplorer.selectedTimeRange) {
    const selectedTimeRange = metricsExplorer.selectableTimeRanges.find(
      (selectableTimeRange) =>
        selectableTimeRange.name === metricsExplorer.selectedTimeRange.name
    );
    if (selectedTimeRange) return selectedTimeRange.name;
  }
  return getDefaultTimeRangeName();
};

/**
 * Selects time grain from metrics explorer.
 * If there is none selected or if the selected time grain is not in the selectable grains
 * returns {@link getDefaultTimeGrain}()
 */
export const selectMetricsExplorerSelectedTimeGrain = (
  state: RillReduxState,
  id: string,
  timeRangeName: TimeRangeName
) => {
  const metricsExplorer = selectMetricsExplorerById(state, id);
  if (metricsExplorer.selectedTimeGrain) {
    const selectedTimeGrain = metricsExplorer.selectableTimeGrains.find(
      (selectableTimeGrain) =>
        selectableTimeGrain.timeGrain === metricsExplorer.selectedTimeGrain &&
        selectableTimeGrain.enabled === true
    );
    if (selectedTimeGrain) return selectedTimeGrain.timeGrain;
  }
  return getDefaultTimeGrain(timeRangeName, metricsExplorer.allTimeRange);
};
