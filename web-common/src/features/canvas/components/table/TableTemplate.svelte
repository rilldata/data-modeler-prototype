<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { createPivotDataStore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
  import PivotEmpty from "@rilldata/web-common/features/dashboards/pivot/PivotEmpty.svelte";
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import PivotTable from "@rilldata/web-common/features/dashboards/pivot/PivotTable.svelte";
  import {
    PivotChipType,
    type PivotDashboardContext,
    type PivotDataStore,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { readable, writable } from "svelte/store";
  import { getTableConfig, validateTableSchema } from "./selector";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  const ctx = getCanvasStateManagers();
  let pivotDataStore: PivotDataStore;
  let isFetching = false;
  let assembled = false;

  // Cache for pivot data stores
  const pivotStoreCache = new Map<string, PivotDataStore>();

  $: tableSpec = rendererProperties as TableSpec;

  $: measures = tableSpec.measures || [];
  $: colDimensions = tableSpec.col_dimensions || [];
  $: rowDimensions = tableSpec.row_dimensions || [];

  $: schema = validateTableSchema(ctx, tableSpec);

  // TODO: Should we move this to canvas entity store?
  $: pivotState = writable<PivotState>({
    active: true,
    columns: {
      measure: measures.map((measure) => ({
        id: measure,
        title: measure,
        type: PivotChipType.Measure,
      })),
      dimension: colDimensions.map((dimension) => ({
        id: dimension,
        title: dimension,
        type: PivotChipType.Dimension,
      })),
    },
    rows: {
      dimension: rowDimensions.map((dimension) => ({
        id: dimension,
        title: dimension,
        type: PivotChipType.Dimension,
      })),
    },
    expanded: {},
    sorting: [],
    columnPage: 1,
    rowPage: 1,
    enableComparison: false,
    rowJoinType: "nest",
    activeCell: null,
  });

  $: pivotConfig = getTableConfig(ctx, tableSpec, $pivotState);

  $: if ($schema.isValid && tableSpec.metrics_view) {
    const cacheKey = tableSpec.metrics_view;
    let store = pivotStoreCache.get(cacheKey);
    if (!store) {
      const pivotDashboardContext: PivotDashboardContext = {
        metricsViewName: readable(tableSpec.metrics_view),
        queryClient: ctx.queryClient,
        enabled: !!ctx.canvasEntity.spec.canvasSpec,
      };
      store = createPivotDataStore(pivotDashboardContext, pivotConfig);
      pivotStoreCache.set(cacheKey, store);
    }
    pivotDataStore = store;
    ({ isFetching, assembled } = $pivotDataStore);
  }

  $: hasColumnAndNoMeasure =
    $pivotState.columns.dimension.length > 0 &&
    $pivotState.columns.measure.length === 0;
</script>

<div class="overflow-y-auto h-full">
  {#if !$schema.isValid}
    <ComponentError error={$schema.error} />
  {:else if pivotDataStore && $pivotDataStore && pivotConfig && $pivotConfig}
    {#if $pivotDataStore?.error?.length}
      <PivotError errors={$pivotDataStore.error} />
    {:else if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
      <PivotEmpty {assembled} {isFetching} {hasColumnAndNoMeasure} />
    {:else}
      <PivotTable
        {pivotDataStore}
        config={pivotConfig}
        {pivotState}
        setPivotExpanded={(expanded) => {
          pivotState.update((state) => ({
            ...state,
            expanded,
          }));
        }}
        setPivotSort={(sorting) => {
          pivotState.update((state) => ({
            ...state,
            sorting,
          }));
        }}
        setPivotRowPage={(page) => {
          pivotState.update((state) => ({
            ...state,
            rowPage: page,
          }));
        }}
      />
    {/if}
  {:else}
    <div class="flex items-center justify-center w-full h-full">
      <Spinner status={EntityStatus.Running} />
    </div>
  {/if}
</div>
