<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { fly } from "svelte/transition";
  import Button from "../../button/Button.svelte";
  import CaretDownIcon from "../../icons/CaretDownIcon.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";
  import SearchableMenuContent from "../../searchable-filter-menu/SearchableMenuContent.svelte";

  export let selectableItems: SearchableFilterSelectableItem[];
  export let selectedItems: string[];
  export let tooltipText: string;
  export let category: string;
  export let disabled = false;
  export let onSelect: (name: string) => void;
  export let onToggleSelectAll: () => void;

  let active = false;

  $: numAvailable = selectableItems?.length ?? 0;
  $: numShown = selectedItems?.filter((x) => x).length ?? 0;

  $: numShownString =
    numAvailable === numShown ? "All" : `${numShown} of ${numAvailable}`;
</script>

<DropdownMenu.Root
  closeOnItemClick={false}
  typeahead={false}
  bind:open={active}
>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={active}
    >
      <Button builders={[builder]} type="text" label={tooltipText} on:click>
        <div
          class="flex items-center gap-x-0.5 px-1 text-gray-700 hover:text-inherit"
        >
          <strong>{`${numShownString} ${category}`}</strong>
          <span
            class="transition-transform"
            class:hidden={disabled}
            class:-rotate-180={active}
          >
            <CaretDownIcon />
          </span>
        </div>
      </Button>

      <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
        <TooltipContent maxWidth="400px">
          {tooltipText}
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>

  <SearchableMenuContent
    allowMultiSelect
    showHiddenSelectionsCount
    fadeUnselected
    requireSelection
    {onSelect}
    {onToggleSelectAll}
    selectableGroups={[{ name: "", items: selectableItems }]}
    selectedItems={[selectedItems]}
  />
</DropdownMenu.Root>
