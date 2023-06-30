import type { Timestamp } from "@bufbuild/protobuf";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/catalog_pb";
import type { MetricsViewFilter_Cond } from "@rilldata/web-common/proto/gen/rill/runtime/v1/queries_pb";
import {
  DashboardState,
  DashboardTimeRange,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1MetricsView,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

export function getDashboardStateFromUrl(
  urlState: string,
  metricsView: V1MetricsView
): Partial<MetricsExplorerEntity> {
  return getDashboardStateFromProto(
    base64ToProto(decodeURIComponent(urlState)),
    metricsView
  );
}

export function getDashboardStateFromProto(
  binary: Uint8Array,
  metricsView: V1MetricsView
): Partial<MetricsExplorerEntity> {
  const dashboard = DashboardState.fromBinary(binary);
  const entity: Partial<MetricsExplorerEntity> = {
    filters: {
      include: [],
      exclude: [],
    },
  };

  if (dashboard.filters) {
    entity.filters.include = fromFiltersProto(dashboard.filters.include);
    entity.filters.exclude = fromFiltersProto(dashboard.filters.exclude);
  }
  if (dashboard.compareTimeRange) {
    entity.selectedComparisonTimeRange = fromTimeRangeProto(
      dashboard.compareTimeRange
    );
  }
  entity.showComparison = dashboard.showComparison ?? true;

  entity.selectedTimeRange = dashboard.timeRange
    ? fromTimeRangeProto(dashboard.timeRange)
    : undefined;
  if (dashboard.timeGrain && dashboard.timeRange) {
    entity.selectedTimeRange.interval = fromTimeGrainProto(dashboard.timeGrain);
  }

  if (dashboard.leaderboardMeasure) {
    entity.leaderboardMeasureName = dashboard.leaderboardMeasure;
  }
  if (dashboard.selectedDimension) {
    entity.selectedDimensionName = dashboard.selectedDimension;
  }

  if (dashboard.allMeasuresVisible) {
    entity.allMeasuresVisible = true;
    entity.visibleMeasureKeys = new Set(
      metricsView.measures.map((measure) => measure.name)
    );
  } else if (dashboard.visibleMeasures) {
    entity.allMeasuresVisible = false;
    entity.visibleMeasureKeys = new Set(dashboard.visibleMeasures);
  }

  if (dashboard.allDimensionsVisible) {
    entity.allDimensionsVisible = true;
    entity.visibleDimensionKeys = new Set(
      metricsView.dimensions.map((measure) => measure.name)
    );
  } else if (dashboard.visibleDimensions) {
    entity.allDimensionsVisible = false;
    entity.visibleDimensionKeys = new Set(dashboard.visibleDimensions);
  }

  return entity;
}

export function base64ToProto(message: string) {
  return new Uint8Array(
    atob(message)
      .split("")
      .map(function (c) {
        return c.charCodeAt(0);
      })
  );
}

function fromFiltersProto(conditions: Array<MetricsViewFilter_Cond>) {
  return conditions.map((condition) => {
    return {
      name: condition.name,
      ...(condition.like?.length ? { like: condition.like } : {}),
      ...(condition.in?.length
        ? {
            in: condition.in.map((v) =>
              v.kind.case === "nullValue" ? null : v.kind.value
            ),
          }
        : {}),
    };
  });
}

function fromTimeRangeProto(timeRange: DashboardTimeRange) {
  const selectedTimeRange: DashboardTimeControls = {
    name: timeRange.name,
  } as DashboardTimeControls;

  selectedTimeRange.name = timeRange.name;
  if (timeRange.timeStart) {
    selectedTimeRange.start = fromTimeProto(timeRange.timeStart);
  }
  if (timeRange.timeEnd) {
    selectedTimeRange.end = fromTimeProto(timeRange.timeEnd);
  }

  return selectedTimeRange;
}

function fromTimeProto(timestamp: Timestamp) {
  return new Date(Number(timestamp.seconds));
}

function fromTimeGrainProto(timeGrain: TimeGrain): V1TimeGrain {
  switch (timeGrain) {
    case TimeGrain.UNSPECIFIED:
    default:
      return V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
    case TimeGrain.MILLISECOND:
      return V1TimeGrain.TIME_GRAIN_MILLISECOND;
    case TimeGrain.SECOND:
      return V1TimeGrain.TIME_GRAIN_SECOND;
    case TimeGrain.MINUTE:
      return V1TimeGrain.TIME_GRAIN_MINUTE;
    case TimeGrain.HOUR:
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case TimeGrain.DAY:
      return V1TimeGrain.TIME_GRAIN_DAY;
    case TimeGrain.WEEK:
      return V1TimeGrain.TIME_GRAIN_WEEK;
    case TimeGrain.MONTH:
      return V1TimeGrain.TIME_GRAIN_MONTH;
    case TimeGrain.QUARTER:
      return V1TimeGrain.TIME_GRAIN_QUARTER;
    case TimeGrain.YEAR:
      return V1TimeGrain.TIME_GRAIN_YEAR;
  }
}
