import type { ActiveValues } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import type { TimeSeriesResponse } from "$common/database-service/DatabaseTimeSeriesActions";
import { updateTimeSeries } from "$lib/redux-store/timeseries/timeseries-slice";
import type { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { prune } from "../../../routes/_surfaces/workspace/explore/utils";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";

export const generateTimeSeriesApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/generateTimeSeries`,
  async (
    {
      metricsDefId,
      measures,
      filters,
      pixels,
      rollupInterval,
    }: {
      metricsDefId: string;
      measures: Array<MeasureDefinitionEntity>;
      filters?: ActiveValues;
      pixels?: number;
      rollupInterval?: RollupInterval;
    },
    thunkAPI
  ) => {
    if (!filters) {
      const metricsLeaderboard: MetricsLeaderboardEntity = (
        thunkAPI.getState() as RillReduxState
      ).metricsLeaderboard.entities[metricsDefId];
      filters = prune(metricsLeaderboard.activeValues);
    }
    const stream = streamingFetchWrapper<TimeSeriesResponse>(
      `metrics/${metricsDefId}/time-series`,
      "POST",
      {
        expressionEntries: measures.map((measure) => [
          measure.id,
          measure.expression,
        ]),
        filters,
        pixels,
        rollupInterval,
      }
    );
    for await (const timeSeriesResponse of stream) {
      thunkAPI.dispatch(
        updateTimeSeries({
          id: timeSeriesResponse.id,
          values: timeSeriesResponse.results,
          rollupInterval: timeSeriesResponse.rollupInterval,
          spark: timeSeriesResponse.spark,
        })
      );
    }
  }
);
