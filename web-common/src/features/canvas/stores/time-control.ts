import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import {
  calculateComparisonTimeRangePartial,
  calculateTimeRangePartial,
  getComparisonTimeRange,
  getTimeGrain,
  type ComparisonTimeRangeState,
  type TimeRangeState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  TimeRangePreset,
  type DashboardTimeControls,
  type TimeRange,
} from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewTimeRange,
  V1ExploreComparisonMode,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import {
  runtime,
  type Runtime,
} from "@rilldata/web-common/runtime-client/runtime-store";
import {
  derived,
  get,
  writable,
  type Readable,
  type Unsubscriber,
  type Writable,
} from "svelte/store";

type AllTimeRange = TimeRange & { isFetching: boolean };

let lastAllTimeRange: AllTimeRange | undefined;

export class TimeControls {
  /**
   * Writables which can be updated by the user
   */
  selectedTimeRange: Writable<DashboardTimeControls | undefined>;
  selectedComparisonTimeRange: Writable<DashboardTimeControls | undefined>;
  showTimeComparison: Writable<boolean>;
  selectedTimezone: Writable<string>;

  /**
   * Derived stores based on writables and spec
   */
  allTimeRange: Readable<AllTimeRange>;
  isReady: Readable<boolean>;
  minTimeGrain: Readable<V1TimeGrain>;
  hasTimeSeries: Readable<boolean>;
  timeRangeStateStore: Readable<TimeRangeState | undefined>;
  comparisonRangeStateStore: Readable<ComparisonTimeRangeState | undefined>;

  private isInitialStateSet: boolean = false;
  private initialStateSubscriber: Unsubscriber | undefined;

  constructor(specStore: CanvasSpecResponseStore) {
    this.allTimeRange = this.combinedTimeRangeSummaryStore(runtime, specStore);

    this.selectedTimeRange = writable(undefined);
    this.selectedComparisonTimeRange = writable(undefined);
    this.showTimeComparison = writable(false);
    this.selectedTimezone = writable("UTC");

    this.minTimeGrain = derived(specStore, (spec) => {
      const metricsViews = spec?.data?.metricsViews || {};
      const minTimeGrain = Object.keys(metricsViews).reduce<V1TimeGrain>(
        (min: V1TimeGrain, metricView) => {
          const metricsViewSpec = metricsViews[metricView]?.state?.validSpec;
          if (
            !metricsViewSpec?.smallestTimeGrain ||
            metricsViewSpec.smallestTimeGrain ===
              V1TimeGrain.TIME_GRAIN_UNSPECIFIED
          )
            return min;
          const timeGrain = metricsViewSpec.smallestTimeGrain;
          return isGrainBigger(min, timeGrain) ? timeGrain : min;
        },
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      );
      return minTimeGrain;
    });

    this.hasTimeSeries = derived(specStore, (spec) => {
      const metricsViews = spec?.data?.metricsViews || {};
      return Object.keys(metricsViews).some((metricView) => {
        const metricsViewSpec = metricsViews[metricView]?.state?.validSpec;
        return Boolean(metricsViewSpec?.timeDimension);
      });
    });

    this.timeRangeStateStore = derived(
      [
        specStore,
        this.allTimeRange,
        this.selectedTimeRange,
        this.selectedTimezone,
        this.minTimeGrain,
      ],
      ([
        spec,
        allTimeRange,
        selectedTimeRange,
        selectedTimezone,
        minTimeGrain,
      ]) => {
        if (!spec?.data || !selectedTimeRange) {
          return undefined;
        }
        const { defaultPreset } = spec.data?.canvas || {};
        const defaultTimeRange = isoDurationToFullTimeRange(
          defaultPreset?.timeRange,
          allTimeRange.start,
          allTimeRange.end,
          selectedTimezone,
        );

        const timeRangeState = calculateTimeRangePartial(
          allTimeRange,
          selectedTimeRange,
          undefined, // scrub not present in canvas yet
          selectedTimezone,
          defaultTimeRange,
          minTimeGrain,
        );
        if (!timeRangeState) return undefined;
        return { ...timeRangeState };
      },
    );

    this.comparisonRangeStateStore = derived(
      [
        specStore,
        this.allTimeRange,
        this.selectedComparisonTimeRange,
        this.selectedTimezone,
        this.showTimeComparison,
        this.timeRangeStateStore,
      ],
      ([
        spec,
        allTimeRange,
        selectedComparisonTimeRange,
        selectedTimezone,
        showTimeComparison,
        timeRangeState,
      ]) => {
        if (!spec?.data || !timeRangeState) return undefined;
        const timeRanges = spec.data?.canvas?.timeRanges;
        return calculateComparisonTimeRangePartial(
          timeRanges,
          allTimeRange,
          selectedComparisonTimeRange,
          selectedTimezone,
          undefined, // scrub not present in canvas yet
          showTimeComparison,
          timeRangeState,
        );
      },
    );

    this.setInitialState(specStore);
  }

  setInitialState = (specStore: CanvasSpecResponseStore) => {
    const defaultStore = derived(
      [this.allTimeRange, specStore],
      ([allTimeRange, spec]) => {
        if (!spec?.data || allTimeRange.isFetching || this.isInitialStateSet) {
          return;
        }

        const selectedTimezone = get(this.selectedTimezone);
        const comparisonTimeRange = get(this.selectedComparisonTimeRange);

        const { defaultPreset } = spec.data?.canvas || {};
        const timeRanges = spec?.data?.canvas?.timeRanges;

        const defaultTimeRange = isoDurationToFullTimeRange(
          defaultPreset?.timeRange,
          allTimeRange.start,
          allTimeRange.end,
          selectedTimezone,
        );

        const newTimeRange: DashboardTimeControls = {
          name: defaultTimeRange.name,
          start: defaultTimeRange.start,
          end: defaultTimeRange.end,
        };

        newTimeRange.interval = getTimeGrain(
          undefined,
          newTimeRange,
          V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        );

        if (
          defaultPreset?.comparisonMode ===
          V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
        ) {
          const newComparisonRange = getComparisonTimeRange(
            timeRanges,
            allTimeRange,
            newTimeRange,
            comparisonTimeRange,
          );
          this.selectedComparisonTimeRange.set(newComparisonRange);
          this.showTimeComparison.set(true);
        }

        this.selectedTimeRange.set(newTimeRange);
        this.isInitialStateSet = true;
      },
    );

    // Subscribe to ensure the derived code runs
    this.initialStateSubscriber = defaultStore.subscribe(() => {});
  };

  destroy = () => {
    this.initialStateSubscriber?.();
  };

  combinedTimeRangeSummaryStore = (
    runtime: Writable<Runtime>,
    specStore: CanvasSpecResponseStore,
  ): Readable<AllTimeRange> => {
    return derived([runtime, specStore], ([r, spec], set) => {
      const metricsReferred = Object.keys(spec?.data?.metricsViews || {});
      if (!metricsReferred.length) {
        return set({
          name: TimeRangePreset.ALL_TIME,
          start: new Date(0),
          end: new Date(),
          isFetching: false,
        });
      }

      const timeRangeQueries = [...metricsReferred].map((metricView) => {
        return createQueryServiceMetricsViewTimeRange(
          r.instanceId,
          metricView,
          {},
          {
            query: {
              queryClient: queryClient,
              staleTime: Infinity,
              cacheTime: Infinity,
            },
          },
        );
      });

      return derived(timeRangeQueries, (timeRanges, querySet) => {
        const isFetching = timeRanges.some((q) => q.isFetching);
        let start = new Date();
        let end = new Date(0);
        timeRanges.forEach((timeRange) => {
          const metricsStart = timeRange.data?.timeRangeSummary?.min;
          const metricsEnd = timeRange.data?.timeRangeSummary?.max;
          if (metricsStart) {
            const metricsStartDate = new Date(metricsStart);
            start = new Date(
              Math.min(start.getTime(), metricsStartDate.getTime()),
            );
          }
          if (metricsEnd) {
            const metricsEndDate = new Date(metricsEnd);
            end = new Date(Math.max(end.getTime(), metricsEndDate.getTime()));
          }
        });
        if (start.getTime() >= end.getTime()) {
          start = new Date(0);
          end = new Date();
        }
        let newTimeRange: AllTimeRange = {
          name: TimeRangePreset.ALL_TIME,
          start,
          end,
          isFetching,
        };
        if (isFetching && lastAllTimeRange) {
          newTimeRange = { ...lastAllTimeRange, isFetching };
        }
        const noChange =
          lastAllTimeRange &&
          lastAllTimeRange.start.getTime() === newTimeRange.start.getTime() &&
          lastAllTimeRange.end.getTime() === newTimeRange.end.getTime();

        /**
         * TODO: We want to avoid updating the store when there is no change
         * to avoid any downstream updates which depend on allTimeRange
         * Returning without setting any value leads to undefined value in the store.
         */
        if (noChange) {
          querySet(lastAllTimeRange);
          return;
        }

        if (!isFetching) {
          lastAllTimeRange = newTimeRange;
        }
        querySet(newTimeRange);
      }).subscribe(set);
    });
  };

  setTimeZone = (timezone: string) => {
    this.selectedTimezone.set(timezone);
  };

  selectTimeRange = (
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) => {
    if (!timeRange.name) return;

    if (timeRange.name === TimeRangePreset.ALL_TIME) {
      this.showTimeComparison.set(false);
    }

    this.selectedTimeRange.set({
      ...timeRange,
      interval: timeGrain,
    });

    this.selectedComparisonTimeRange.set(comparisonTimeRange);
  };

  setSelectedComparisonRange = (comparisonTimeRange: DashboardTimeControls) => {
    this.selectedComparisonTimeRange.set(comparisonTimeRange);
  };

  displayTimeComparison = (showTimeComparison: boolean) => {
    this.showTimeComparison.set(showTimeComparison);
  };
}
