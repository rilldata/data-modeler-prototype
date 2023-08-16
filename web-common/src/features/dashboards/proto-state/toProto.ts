import {
  NullValue,
  PartialMessage,
  Timestamp,
  Value,
} from "@bufbuild/protobuf";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import {
  DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  MetricsViewFilter,
  MetricsViewFilter_Cond,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/queries_pb";
import {
  TimeGrain,
  TimeGrain as TimeGrainProto,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  DashboardState,
  DashboardState_DashboardLeaderboardContextColumn,
  DashboardTimeRange,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewFilterCond,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

// TODO: make a follow up PR to use the one from the proto directly
const LeaderboardContextColumnMap: Record<
  LeaderboardContextColumn,
  DashboardState_DashboardLeaderboardContextColumn
> = {
  [LeaderboardContextColumn.PERCENT]:
    DashboardState_DashboardLeaderboardContextColumn.PERCENT,
  [LeaderboardContextColumn.DELTA_CHANGE]:
    DashboardState_DashboardLeaderboardContextColumn.DELTA_CHANGE,
  [LeaderboardContextColumn.HIDDEN]:
    DashboardState_DashboardLeaderboardContextColumn.HIDDEN,
};

export function getProtoFromDashboardState(
  metrics: MetricsExplorerEntity
): string {
  if (!metrics) return "";

  const state: PartialMessage<DashboardState> = {};
  if (metrics.filters) {
    state.filters = toFiltersProto(metrics.filters) as any;
  }
  if (metrics.selectedTimeRange) {
    state.timeRange = toTimeRangeProto(metrics.selectedTimeRange);
    if (metrics.selectedTimeRange.interval) {
      state.timeGrain = toTimeGrainProto(metrics.selectedTimeRange.interval);
    }
  }
  if (metrics.selectedComparisonTimeRange) {
    state.compareTimeRange = toTimeRangeProto(
      metrics.selectedComparisonTimeRange
    );
  }
  if (metrics.lastDefinedScrubRange) {
    state.scrubRange = toTimeRangeProto(metrics.lastDefinedScrubRange);
  }
  state.showComparison = Boolean(metrics.showComparison);
  if (metrics.selectedTimezone) {
    state.selectedTimezone = metrics.selectedTimezone;
  }
  if (metrics.leaderboardMeasureName) {
    state.leaderboardMeasure = metrics.leaderboardMeasureName;
  }
  if (metrics.selectedDimensionName) {
    state.selectedDimension = metrics.selectedDimensionName;
  }

  if (metrics.allMeasuresVisible) {
    state.allMeasuresVisible = true;
  } else if (metrics.visibleMeasureKeys) {
    state.visibleMeasures = [...metrics.visibleMeasureKeys];
  }

  if (metrics.allDimensionsVisible) {
    state.allDimensionsVisible = true;
  } else if (metrics.visibleDimensionKeys) {
    state.visibleDimensions = [...metrics.visibleDimensionKeys];
  }

  if (metrics.leaderboardContextColumn) {
    state.leaderboardContextColumn =
      LeaderboardContextColumnMap[metrics.leaderboardContextColumn];
  }

  const message = new DashboardState(state);
  return protoToBase64(message.toBinary());
}

function protoToBase64(proto: Uint8Array) {
  return btoa(String.fromCharCode.apply(null, proto));
}

function toFiltersProto(filters: V1MetricsViewFilter) {
  return new MetricsViewFilter({
    include: toFilterCondProto(filters.include) as any,
    exclude: toFilterCondProto(filters.exclude) as any,
  });
}

function toTimeRangeProto(range: DashboardTimeControls) {
  const timeRangeArgs: PartialMessage<DashboardTimeRange> = {
    name: range.name,
  };
  if (
    range.name === TimeRangePreset.CUSTOM ||
    range.name === TimeComparisonOption.CUSTOM
  ) {
    timeRangeArgs.timeStart = toTimeProto(range.start);
    timeRangeArgs.timeEnd = toTimeProto(range.end);
  }
  return new DashboardTimeRange(timeRangeArgs);
}

function toTimeProto(date: Date) {
  return new Timestamp({
    seconds: BigInt(date.getTime()),
  });
}

function toTimeGrainProto(timeGrain: V1TimeGrain) {
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_UNSPECIFIED:
    default:
      return TimeGrain.UNSPECIFIED;
    case V1TimeGrain.TIME_GRAIN_MILLISECOND:
      return TimeGrain.MILLISECOND;
    case V1TimeGrain.TIME_GRAIN_SECOND:
      return TimeGrain.SECOND;
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return TimeGrainProto.MINUTE;
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return TimeGrainProto.HOUR;
    case V1TimeGrain.TIME_GRAIN_DAY:
      return TimeGrainProto.DAY;
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return TimeGrainProto.WEEK;
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return TimeGrainProto.MONTH;
    case V1TimeGrain.TIME_GRAIN_QUARTER:
      return TimeGrainProto.QUARTER;
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return TimeGrainProto.YEAR;
  }
}

function toFilterCondProto(conds: Array<MetricsViewFilterCond>) {
  return conds.map(
    (include) =>
      new MetricsViewFilter_Cond({
        name: include.name,
        like: include.like,
        in: include.in.map(
          (v) =>
            (v === null
              ? new Value({
                  kind: {
                    case: "nullValue",
                    value: NullValue.NULL_VALUE,
                  },
                })
              : new Value({
                  kind: {
                    case: "stringValue",
                    value: v as string,
                  },
                })) as any
        ),
      })
  );
}
