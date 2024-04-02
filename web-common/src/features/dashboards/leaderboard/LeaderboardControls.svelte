<script lang="ts">
  import SelectMenu from "@rilldata/web-common/components/menu/shadcn/SelectMenu.svelte";
  import SearchableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterButton.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { createShowHideDimensionsStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import { crossfade, fly } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";

  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { useMetricsView } from "../selectors";
  import { getStateManagers } from "../state-managers/state-managers";
  import LeaderboardContextColumnMenu from "./LeaderboardContextColumnMenu.svelte";

  export let metricsViewName: string;
  export let selectedDimensions: boolean[];
  export let dimensions: MetricsViewSpecMeasureV2[];

  const {
    actions: {
      contextCol: { setContextColumn },
      setLeaderboardMeasureName,
    },
  } = getStateManagers();

  $: metricsView = useMetricsView($runtime.instanceId, metricsViewName);

  $: measures = $metricsView.data?.measures;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsViewName];

  function handleMeasureUpdate(event: CustomEvent) {
    setLeaderboardMeasureName(event.detail.key);
  }

  function measureKeyAndMain(measure: MetricsViewSpecMeasureV2) {
    // CAST SAFETY: measure expression must exist!
    const main = (
      measure.label?.length ? measure.label : measure.expression
    ) as string;
    return {
      main,
      // CAST SAFETY: measure expression must exist!
      key: measure.name ?? (measure.expression as string),
    };
  }

  function formatForSelector(
    measure: MetricsViewSpecMeasureV2,
  ): (MetricsViewSpecMeasureV2 & { key: string; main: string }) | undefined {
    if (!measure) return undefined;
    return {
      ...measure,
      ...measureKeyAndMain(measure),
    };
  }

  let [send, receive] = crossfade({
    fallback: (node, _params, _intro) => fly(node),
  });

  /** this should be a single element */
  // reset selections based on the active leaderboard measure
  let activeLeaderboardMeasure: ReturnType<typeof formatForSelector>;

  $: unformattedMeasure =
    measures?.length && metricsExplorer?.leaderboardMeasureName
      ? measures.find(
          (measure) => measure.name === metricsExplorer?.leaderboardMeasureName,
        )
      : undefined;

  $: activeLeaderboardMeasure =
    unformattedMeasure && formatForSelector(unformattedMeasure);

  /** this controls the animation direction */

  $: options =
    measures?.map((measure) => {
      return {
        ...measure,
        ...measureKeyAndMain(measure),
      };
    }) || [];

  /** set the selection only if measures is not undefined */
  $: selection = unformattedMeasure && measureKeyAndMain(unformattedMeasure);

  $: validPercentOfTotal =
    activeLeaderboardMeasure?.validPercentOfTotal || false;

  // if the percent of total is currently being shown,
  // but it is not valid for this measure, then turn it off
  $: if (
    !validPercentOfTotal &&
    metricsExplorer?.leaderboardContextColumn ===
      LeaderboardContextColumn.PERCENT
  ) {
    setContextColumn(LeaderboardContextColumn.HIDDEN);
  }

  $: showHideDimensions = createShowHideDimensionsStore(
    metricsViewName,
    metricsView,
  );

  const setAllDimensionsNotVisible = () => {
    showHideDimensions.setAllToNotVisible();
  };
  const setAllDimensionsVisible = () => {
    showHideDimensions.setAllToVisible();
  };
</script>

<div>
  {#if measures && options.length && selection}
    <div
      class="flex flex-row items-center ui-copy-muted gap-x-0.5"
      in:send|global={{ key: "leaderboard-metric" }}
      style:max-width="450px"
    >
      <SearchableFilterButton
        selectableItems={dimensions}
        selectedItems={selectedDimensions}
        on:item-clicked
        on:deselect-all={setAllDimensionsNotVisible}
        on:select-all={setAllDimensionsVisible}
        label="Dimensions"
        tooltipText="Choose dimensions to display"
      />

      <SelectMenu
        fixedText="Showing"
        {options}
        selections={[selection.key]}
        on:select={handleMeasureUpdate}
        ariaLabel="Select a measure to filter by"
      />

      <LeaderboardContextColumnMenu {validPercentOfTotal} />
    </div>
  {:else}
    <div
      class="flex flex-row items-center"
      style:grid-column-gap=".4rem"
      in:receive|global={{ key: "loading-leaderboard-metric" }}
    >
      pulling leaderboards <Spinner status={EntityStatus.Running} />
    </div>
  {/if}
</div>
