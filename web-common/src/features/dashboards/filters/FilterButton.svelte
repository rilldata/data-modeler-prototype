<script lang="ts">
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import { getStateManagers } from "../state-managers/state-managers";
  import { getMeasureDisplayName } from "./getDisplayName";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  const {
    selectors: {
      dimensions: { allDimensions },
      dimensionFilters: { dimensionHasFilter },
      measures: { filteredSimpleMeasures },
      measureFilters: { measureHasFilter },
    },
    actions: {
      filters: { setTemporaryFilterName },
    },
  } = getStateManagers();

  $: selectableGroups = [
    <SearchableFilterSelectableGroup>{
      name: "MEASURES",
      items:
        $filteredSimpleMeasures()
          ?.map((m) => ({
            name: m.name as string,
            label: getMeasureDisplayName(m),
          }))
          .filter((m) => !$measureHasFilter(m.name)) ?? [],
    },
    <SearchableFilterSelectableGroup>{
      name: "DIMENSIONS",
      items:
        $allDimensions
          ?.map((d) => ({
            name: (d.name || d.column) as string,
            label: getDimensionDisplayName(d),
          }))
          .filter((d) => !$dimensionHasFilter(d.name)) ?? [],
    },
  ];

  let open = false;
</script>

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip distance={8} suppress={open}>
      <button class:active={open} use:builder.action {...builder}>
        <Add size="17px" />
      </button>
      <TooltipContent slot="tooltip-content">Add filter</TooltipContent>
    </Tooltip>
  </DropdownMenu.Trigger>

  <SearchableMenuContent
    allowMultiSelect={false}
    onSelect={(name) => {
      setTemporaryFilterName(name);
    }}
    {selectableGroups}
    selectedItems={[]}
  />
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply flex items-center justify-center;
    @apply border border-dashed border-slate-300;
    @apply bg-white;
  }

  button:hover {
    @apply bg-slate-100;
  }

  button:active,
  .active {
    @apply bg-slate-200;
  }
</style>
