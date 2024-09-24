<script lang="ts">
  import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import { fly } from "svelte/transition";
  import { Chip } from "@rilldata/web-common/components/chip";

  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import SearchableMenuContent from "./SearchableMenuContent.svelte";

  export let selectableItems: SearchableFilterSelectableItem[];
  export let selectedItems: string[];
  export let tooltipText: string;
  export let label: string;
  export let onSelect: (name: string) => void;

  let open = false;
  let searchText = "";
</script>

<DropdownMenu.Root
  bind:open
  typeahead={false}
  onOpenChange={(open) => {
    if (!open) {
      searchText = "";
    }
  }}
>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={open}
    >
      <Chip type="measure" active={open} {label} builders={[builder]}>
        <div class="font-bold truncate">
          {label}
        </div>
      </Chip>
      <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
        <TooltipContent maxWidth="400px">
          {tooltipText}
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>

  <SearchableMenuContent
    bind:searchText
    showSelection
    {onSelect}
    selectedItems={[selectedItems]}
    selectableGroups={[{ name: "", items: selectableItems }]}
  />
</DropdownMenu.Root>
