<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import * as Select from "@rilldata/web-common/components/select";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/shadcn/DashboardVisibilityDropdown.svelte";

  export let metricViewName: string;

  const {
    selectors: {
      measures: {
        filteredSimpleMeasures,
        leaderboardMeasureName,
        getMeasureByName,
      },
      dimensions: { visibleDimensions, allDimensions },
    },
    actions: {
      dimensions: { toggleDimensionVisibility, setVisibleDimensions },

      contextCol: { setContextColumn },
      setLeaderboardMeasureName,
    },
  } = getStateManagers();

  let active = false;

  $: measures = $filteredSimpleMeasures();

  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  $: activeLeaderboardMeasure = $getMeasureByName($leaderboardMeasureName);

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

  $: visibleDimensionsNames = $visibleDimensions
    .map(({ name }) => name)
    .filter(isDefined);
  $: allDimensionNames = $allDimensions
    .map(({ name }) => name)
    .filter(isDefined);

  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }
</script>

<div>
  {#if measures.length && activeLeaderboardMeasure}
    <div
      class="flex flex-row items-center ui-copy-muted gap-x-0.5"
      style:max-width="450px"
    >
      <DashboardVisibilityDropdown
        category="Dimensions"
        tooltipText="Choose dimensions to display"
        onSelect={(name) => toggleDimensionVisibility(name)}
        selectableItems={$allDimensions.map(({ name, label }) => ({
          name,
          label: label ?? name,
        }))}
        selectedItems={visibleDimensionsNames}
        onToggleSelectAll={() => {
          const deselectAll =
            visibleDimensionsNames.length === allDimensionNames.length;
          setVisibleDimensions(
            allDimensionNames.slice(0, deselectAll ? 1 : undefined),
          );
        }}
      />

      <Select.Root
        bind:open={active}
        selected={{ value: activeLeaderboardMeasure.name, label: "" }}
        items={measures.map((measure) => ({
          value: measure.name ?? "",
          label: measure.label ?? measure.name,
        }))}
        onSelectedChange={(newSelection) => {
          if (!newSelection?.value) return;
          setLeaderboardMeasureName(newSelection.value);
        }}
      >
        <Select.Trigger class="outline-none border-none w-fit  px-0 gap-x-0.5">
          <Button type="text" label="Select a measure to filter by">
            <span class="truncate text-gray-700 hover:text-inherit">
              Showing <b>
                {activeLeaderboardMeasure?.label ??
                  activeLeaderboardMeasure.name}
              </b>
            </span>
          </Button>
        </Select.Trigger>

        <Select.Content
          sameWidth={false}
          align="start"
          class="max-h-80 overflow-y-auto min-w-44"
        >
          {#each measures as measure (measure.name)}
            <Select.Item
              value={measure.name}
              label={measure.label ?? measure.name}
              class="text-[12px]"
            >
              <div class="flex flex-col">
                <div class:font-bold={$leaderboardMeasureName === measure.name}>
                  {measure.label ?? measure.name}
                </div>

                <p class="ui-copy-muted" style:font-size="11px">
                  {measure.description}
                </p>
              </div>
            </Select.Item>
          {/each}
        </Select.Content>
      </Select.Root>
    </div>
  {/if}
</div>
