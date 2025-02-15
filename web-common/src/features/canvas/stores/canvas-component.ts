import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { get } from "svelte/store";
import { Filters } from "./filters";
import type { CanvasResolvedSpec } from "./spec";
import { TimeControls } from "./time-control";

export class CanvasComponentState {
  name: string;

  metricsViewName: string | undefined;
  filterText: string | undefined;
  timeFilterText: string | undefined;

  localFilters: Filters;
  localTimeControls: TimeControls;

  constructor(
    name: string,
    specStore: CanvasSpecResponseStore,
    spec: CanvasResolvedSpec,
  ) {
    const componentResourceStore = spec.getComponentResourceFromName(name);
    const componentResource = get(componentResourceStore);

    this.metricsViewName = componentResource?.rendererProperties
      ?.metrics_view as string | undefined;
    this.filterText = componentResource?.rendererProperties
      ?.dimension_filters as string | undefined;
    this.timeFilterText = componentResource?.rendererProperties
      ?.time_filters as string | undefined;
    this.name = name;
    this.localFilters = new Filters(spec);
    this.localTimeControls = new TimeControls(specStore, name);

    if (
      componentResource &&
      componentResource.rendererProperties?.dimension_filters
    ) {
      this.localFilters.setFiltersFromText(
        componentResource.rendererProperties?.dimension_filters as string,
      );
    }

    if (
      componentResource &&
      componentResource.rendererProperties?.time_filters
    ) {
      this.localTimeControls.setTimeFiltersFromText(
        componentResource.rendererProperties?.time_filters as string,
      );
    }
  }
}
