import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  buildValidMetricsViewFilter,
  createAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  type V1Expression,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { GridStack } from "gridstack";
import { derived, writable, type Readable, type Writable } from "svelte/store";
import { CanvasComponentState } from "./canvas-component";
import { Filters } from "./filters";
import { CanvasResolvedSpec } from "./spec";
import { TimeControls } from "./time-control";

export class CanvasEntity {
  name: string;

  /** Local state store for canvas components */
  components: Map<string, CanvasComponentState>;

  /**
   * Time controls for the canvas entity containing various
   * time related writables
   */
  timeControls: TimeControls;

  /**
   * Dimension and measure filters for the canvas entity
   */
  filters: Filters;

  /**
   * Spec store containing selectors derived from ResolveCanvas query
   */
  spec: CanvasResolvedSpec;

  gridstack?: GridStack | null;

  /**
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponentIndex: Writable<number | null>;

  private specStore: CanvasSpecResponseStore;

  constructor(name: string) {
    this.specStore = derived(runtime, (r, set) =>
      useCanvas(r.instanceId, name, { queryClient }).subscribe(set),
    );

    this.name = name;

    this.components = new Map();
    this.selectedComponentIndex = writable(null);
    this.spec = new CanvasResolvedSpec(this.specStore);
    this.timeControls = new TimeControls(this.specStore);
    this.filters = new Filters(this.spec);
  }

  setSelectedComponentIndex = (index: number | null) => {
    this.selectedComponentIndex.set(index);
  };

  useComponent = (componentName: string): CanvasComponentState => {
    let componentEntity = this.components.get(componentName);

    if (!componentEntity) {
      componentEntity = new CanvasComponentState(
        componentName,
        this.specStore,
        this.spec,
      );
      this.components.set(componentName, componentEntity);
    }
    return componentEntity;
  };

  removeComponent = (componentName: string) => {
    this.components.delete(componentName);
  };

  setGridstack(gridstack: GridStack | null) {
    this.gridstack = gridstack;
  }

  /**
   * Helper method to get the time range and where clause for a given metrics view
   * with the ability to override the time range and filter
   */
  componentTimeAndFilterStore = (
    componentName: string,
  ): Readable<TimeAndFilterStore> => {
    const { timeControls, filters, spec, useComponent } = this;

    const componentSpecStore = spec.getComponentResourceFromName(componentName);

    return derived(componentSpecStore, (componentSpec, set) => {
      const metricsViewName = componentSpec?.rendererProperties
        ?.metrics_view as string;

      if (!metricsViewName) {
        throw new Error("Metrics view name is not set for component");
      }

      const component = useComponent(componentName);
      const dimensionsStore = spec.getDimensionsForMetricView(metricsViewName);
      const measuresStore = spec.getMeasuresForMetricView(metricsViewName);

      return derived(
        [
          timeControls.timeRangeStateStore,
          component.localTimeControls.timeRangeStateStore,
          timeControls.comparisonRangeStateStore,
          component.localTimeControls.comparisonRangeStateStore,
          timeControls.selectedTimezone,
          component.localTimeControls.showTimeComparison,
          filters.whereFilter,
          filters.dimensionThresholdFilters,
          dimensionsStore,
          measuresStore,
        ],
        ([
          globalTimeRangeState,
          localTimeRangeState,
          globalComparisonRangeState,
          localComparisonRangeState,
          timeZone,
          showLocalTimeComparison,
          whereFilter,
          dtf,
          dimensions,
          measures,
        ]) => {
          // Time Filters
          let timeRange: V1TimeRange = {
            start: globalTimeRangeState?.timeStart,
            end: globalTimeRangeState?.timeEnd,
            timeZone,
          };

          let timeGrain = globalTimeRangeState?.selectedTimeRange?.interval;

          let comparisonTimeRange: V1TimeRange | undefined = {
            start: globalComparisonRangeState?.comparisonTimeStart,
            end: globalComparisonRangeState?.comparisonTimeEnd,
            timeZone,
          };
          if (componentSpec?.rendererProperties?.time_filters) {
            timeRange = {
              start: localTimeRangeState?.timeStart,
              end: localTimeRangeState?.timeEnd,
              timeZone,
            };

            comparisonTimeRange = {
              start: localComparisonRangeState?.comparisonTimeStart,
              end: localComparisonRangeState?.comparisonTimeEnd,
              timeZone,
            };

            if (!showLocalTimeComparison) comparisonTimeRange = undefined;
            timeGrain = localTimeRangeState?.selectedTimeRange?.interval;
          }

          // Dimension Filters
          const globalWhere =
            buildValidMetricsViewFilter(
              whereFilter,
              dtf,
              dimensions,
              measures,
            ) ?? createAndExpression([]);

          let where: V1Expression | undefined = globalWhere;

          if (componentSpec?.rendererProperties?.dimension_filters) {
            const componentWhere = component.localFilters.getFiltersFromText(
              componentSpec.rendererProperties.dimension_filters as string,
            );
            where = mergeFilters(globalWhere, componentWhere);
          }

          return { timeRange, comparisonTimeRange, where, timeGrain };
        },
      ).subscribe(set);
    });
  };
}
