<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import { createPivotDataStore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
  import {
    PivotChipType,
    type PivotDashboardContext,
    type PivotDataStore,
    type PivotDataStoreConfig,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { readable, type Readable, writable } from "svelte/store";
  import { getTableConfig, validateTableSchema } from "./selector";
  import TableRenderer from "./TableRenderer.svelte";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;

  const ctx = getCanvasStateManagers();

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

  let pivotDataStore: PivotDataStore | undefined = undefined;
  let pivotConfig: Readable<PivotDataStoreConfig> | undefined = undefined;

  // TODO: Consider moving to a memoized store
  $: if ($schema.isValid) {
    const pivotDashboardContext: PivotDashboardContext = {
      metricsViewName: readable(tableSpec.metrics_view),
      queryClient: ctx.queryClient,
      enabled: !!ctx.canvasEntity.spec.canvasSpec,
    };
    pivotConfig = getTableConfig(
      ctx,
      tableSpec,
      $pivotState,
      timeAndFilterStore,
    );
    pivotDataStore = createPivotDataStore(pivotDashboardContext, pivotConfig);
  }
</script>

<div class="overflow-y-auto h-full">
  {#if !$schema.isValid}
    <ComponentError error={$schema.error} />
  {:else if pivotDataStore && pivotConfig && $pivotConfig}
    <TableRenderer
      {pivotDataStore}
      config={$pivotConfig}
      pivotDashboardStore={pivotState}
    />
  {:else}
    <div>Loading...</div>
  {/if}
</div>
