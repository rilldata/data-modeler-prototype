import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getMapFromArray } from "@rilldata/utils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecMeasureV2,
  type V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

/**
 * Single use class to correct incorrect use of advanced measures.
 * Reason to use a class here is to avoid too many arguments in methods (especially measureIsValidForComponent).
 * NOTE: this doesnt have to deal with V1ExploreSpec since it is assumed measures/dimensions not present are already removed
 * TODO: this should not be necessary once we use V1ExplorePreset for everything
 */
export class AdvancedMeasureCorrector {
  private measuresMap: Map<string, MetricsViewSpecMeasureV2>;
  private measuresGrains: Map<string, V1TimeGrain>;

  private constructor(
    private readonly dashboard: MetricsExplorerEntity,
    private readonly metricsViewSpec: V1MetricsViewSpec,
  ) {
    this.measuresMap = getMapFromArray(
      metricsViewSpec.measures ?? [],
      (m) => m.name ?? "",
    );
    this.measuresGrains = getMapFromArray(
      metricsViewSpec.measures ?? [],
      (m) => m.name ?? "",
      (m) => {
        const d = m.requiredDimensions?.find(
          (d) =>
            d.timeGrain && d.timeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        );
        return d?.timeGrain ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
      },
    );
  }

  public static correct(
    dashboard: MetricsExplorerEntity,
    metricsViewSpec: V1MetricsViewSpec,
  ) {
    new AdvancedMeasureCorrector(dashboard, metricsViewSpec).correct();
  }

  private correct() {
    this.correctFilters();
    this.correctLeaderboards();
    this.correctTimeDimensionDetails();
    this.correctPivot();
  }

  private correctFilters() {
    this.dashboard.dimensionThresholdFilters.forEach((dimensionThreshold) => {
      dimensionThreshold.filters = dimensionThreshold.filters.filter(
        (dtf) => !this.measureIsValidForComponent(dtf.measure, false, false),
      );
    });
    this.dashboard.dimensionThresholdFilters =
      this.dashboard.dimensionThresholdFilters.filter(
        (dt) => dt.filters.length,
      );
  }

  private correctLeaderboards() {
    if (
      this.dashboard.leaderboardMeasureName &&
      !this.measureIsValidForComponent(
        this.dashboard.leaderboardMeasureName,
        true,
        false,
      )
    ) {
      return;
    }

    this.dashboard.leaderboardMeasureName = "";
    for (const measure of this.metricsViewSpec.measures ?? []) {
      if (!this.measureIsValidForComponent(measure.name ?? "", true, false)) {
        this.dashboard.leaderboardMeasureName = measure.name ?? "";
        break;
      }
    }
  }

  private correctTimeDimensionDetails() {
    if (
      !this.measureIsValidForComponent(
        this.dashboard.tdd.expandedMeasureName ?? "",
        true,
        false,
      )
    ) {
      return;
    }

    this.dashboard.tdd.expandedMeasureName = "";
    if (
      this.dashboard.activePage ===
      DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL
    ) {
      this.dashboard.activePage = DashboardState_ActivePage.DEFAULT;
    }
  }

  private correctPivot() {
    this.dashboard.pivot.columns.measure =
      this.dashboard.pivot.columns.measure.filter(
        (m) => !this.measureIsValidForComponent(m.id, true, false),
      );
    this.dashboard.pivot.sorting = this.dashboard.pivot.sorting.filter(
      (s) =>
        !this.measuresMap.has(s.id) ||
        !this.measureIsValidForComponent(s.id, true, false),
    );
  }

  /**
   * Checks if a measure is valid for the component based on dashboard selections and component support.
   * Additional arguments indicate whether certain types of advanced measures are supported in the component or not.
   */
  private measureIsValidForComponent(
    measureName: string,
    supportsComparisonMeasure: boolean,
    supportsWindowedMeasure: boolean,
  ) {
    const measure = this.measuresMap.get(measureName);
    if (!measure) return true;
    const grain =
      this.measuresGrains.get(measureName) ??
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED;

    switch (true) {
      // selected grain and measure's grain mismatch
      case grain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED &&
        grain !== this.dashboard.selectedTimeRange?.interval:
        return true;

      // for comparison measures,
      // if the component supports it and no time comparison is enabled
      // or if the component does not support it
      case measure.type ===
        MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON &&
        ((supportsComparisonMeasure &&
          (!this.dashboard.showTimeComparison ||
            !this.dashboard.selectedComparisonTimeRange)) ||
          !supportsComparisonMeasure):
        return true;

      // for measures with window operations, if the component doesnt support it
      case !!measure.window && !supportsWindowedMeasure:
        return true;

      default:
        return false;
    }
  }
}
