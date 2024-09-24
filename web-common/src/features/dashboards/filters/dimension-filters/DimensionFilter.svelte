<script lang="ts">
  import { getStateManagers } from "../../state-managers/state-managers";
  import { fly } from "svelte/transition";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import RemovableListBody from "@rilldata/web-common/components/chip/removable-list-chip/RemovableListBody.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useTimeControlStore } from "../../time-controls/time-control-store";
  import { getFilterSearchListQuery } from "../../selectors/index";
  import { V1MetricsViewComparisonRow } from "@rilldata/web-common/runtime-client";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";

  export let name: string;
  export let label: string;
  export let selectedValues: string[];
  export let excludeMode: boolean;
  export let openOnMount: boolean;
  export let readOnly: boolean = false;
  export let onRemove: () => void;
  export let onSelect: (value: string) => void;
  export let onToggleFilterMode: () => void;

  const StateManagers = getStateManagers();
  const timeControls = useTimeControlStore(StateManagers);

  let open = openOnMount && !selectedValues.length;
  let searchText = "";
  let allValues: V1MetricsViewComparisonRow[] = [];

  $: ({ metricsViewName, dashboardStore } = StateManagers);

  $: ({ instanceId } = $runtime);

  $: ({ timeStart, timeEnd, ready: timeControlsReady } = $timeControls);

  $: ({ leaderboardMeasureName } = $dashboardStore);

  $: addNull = searchText.length !== 0 && "null".includes(searchText);

  $: searchQuery = getFilterSearchListQuery(searchText, {
    instanceId,
    metricsViewName: $metricsViewName,
    dimensionName: name,
    leaderboardMeasureName,
    timeStart,
    timeEnd,
    addNull,
    enabled: Boolean(timeControlsReady && open),
  });

  $: allValues = $searchQuery?.data?.rows ?? allValues;

  $: allSelected = Boolean(
    selectedValues.length && allValues?.length === selectedValues.length,
  );

  function onToggleSelectAll() {
    allValues?.forEach(({ dimensionValue }) => {
      if (!allSelected && selectedValues.includes(dimensionValue)) return;

      onSelect(dimensionValue);
    });
  }
</script>

<DropdownMenu.Root
  bind:open
  typeahead={false}
  closeOnItemClick={false}
  onOpenChange={(open) => {
    if (open) {
      searchText = "";
    } else {
      if (selectedValues.length === 0) {
        onRemove();
      }
    }
  }}
>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={open || readOnly}
    >
      <Chip
        builders={[builder]}
        type="dimension"
        active={open}
        label="View filter"
        on:remove={onRemove}
        removable={!readOnly}
        {readOnly}
      >
        <svelte:fragment slot="remove-tooltip">
          <slot name="remove-tooltip-content">
            remove {selectedValues.length}
            value{#if selectedValues.length !== 1}s{/if} for {name}</slot
          >
        </svelte:fragment>

        <RemovableListBody
          slot="body"
          label={excludeMode ? `Exclude ${label}` : label}
          show={1}
          values={selectedValues}
        />
      </Chip>
      <div slot="tooltip-content" transition:fly={{ duration: 100, y: 4 }}>
        <TooltipContent maxWidth="400px">
          <TooltipTitle>
            <svelte:fragment slot="name">{name}</svelte:fragment>
            <svelte:fragment slot="description">dimension</svelte:fragment>
          </TooltipTitle>
          Click to edit the the filters in this dimension
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>

  <SearchableMenuContent
    {onSelect}
    {onToggleSelectAll}
    bind:searchText
    showXForSelected={excludeMode}
    selectedItems={[selectedValues]}
    allowMultiSelect={true}
    selectableGroups={[
      {
        name: "DIMENSIONS",
        items: allValues.map(({ dimensionValue }) => ({
          name: dimensionValue,
          label: dimensionValue,
        })),
      },
    ]}
  >
    <Button slot="action" on:click={onToggleFilterMode} type="secondary">
      {#if excludeMode}
        Include
      {:else}
        Exclude
      {/if}
    </Button>
  </SearchableMenuContent>
</DropdownMenu.Root>
