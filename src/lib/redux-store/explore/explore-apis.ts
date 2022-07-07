import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { prune } from "../../../routes/_surfaces/workspace/explore/utils";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import {
  clearSelectedLeaderboardValues,
  initMetricsExplore,
  MetricsExploreEntity,
  setLeaderboardDimensionValues,
  setMeasureId,
  toggleExploreMeasure,
  toggleLeaderboardActiveValue,
} from "$lib/redux-store/explore/explore-slice";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { generateTimeSeriesApi } from "$lib/redux-store/timeseries/timeseries-apis";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { generateBigNumbersApi } from "$lib/redux-store/big-number/big-number-apis";

const updateExploreWrapper = (dispatch, id: string) => {
  dispatch(updateLeaderboardValuesApi(id));
  dispatch(generateTimeSeriesApi({ id }));
  dispatch(generateBigNumbersApi({ id }));
};

export const initAndUpdateExplore = (
  dispatch,
  id: string,
  dimensions: Array<DimensionDefinitionEntity>,
  measures: Array<MeasureDefinitionEntity>
) => {
  dispatch(initMetricsExplore(id, dimensions, measures));
  updateExploreWrapper(dispatch, id);
};

export const toggleExploreMeasureAndUpdate = (
  dispatch,
  id: string,
  measureId: string
) => {
  dispatch(toggleExploreMeasure(id, measureId));
  dispatch(generateTimeSeriesApi({ id }));
  dispatch(generateBigNumbersApi({ id }));
};

export const setMeasureIdAndUpdateLeaderboard = (
  dispatch,
  id: string,
  measureId: string
) => {
  dispatch(setMeasureId(id, measureId));
  dispatch(updateLeaderboardValuesApi(id));
};

export const toggleSelectedLeaderboardValueAndUpdate = (
  dispatch,
  id: string,
  dimensionName: string,
  dimensionValue: unknown,
  include: boolean
) => {
  dispatch(
    toggleLeaderboardActiveValue(id, dimensionName, dimensionValue, include)
  );
  updateExploreWrapper(dispatch, id);
};

export const clearSelectedLeaderboardValuesAndUpdate = (
  dispatch,
  id: string
) => {
  dispatch(clearSelectedLeaderboardValues(id));
  updateExploreWrapper(dispatch, id);
};

export const updateLeaderboardValuesApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/updateLeaderboard`,
  async (id: string, thunkAPI) => {
    const metricsLeaderboard: MetricsExploreEntity = (
      thunkAPI.getState() as RillReduxState
    ).metricsLeaderboard.entities[id];
    const filters = prune(metricsLeaderboard.activeValues);
    const requestBody = {
      measureId: metricsLeaderboard.leaderboardMeasureId,
      filters,
    };

    const stream = streamingFetchWrapper<{
      dimensionName: string;
      values: Array<unknown>;
    }>(`metrics/${metricsLeaderboard.id}/leaderboards`, "POST", requestBody);
    for await (const dimensionData of stream) {
      thunkAPI.dispatch(
        setLeaderboardDimensionValues(
          metricsLeaderboard.id,
          dimensionData.dimensionName,
          dimensionData.values
        )
      );
    }
  }
);
