<script lang="ts">
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import Button from "../button/Button.svelte";
  import { Search } from "../search";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { matchSorter } from "match-sorter";

  const voidFn = () => {};

  export let selectableGroups: SearchableFilterSelectableGroup[];
  export let selectedItems: string[][];
  export let allowMultiSelect = false;
  export let requireSelection = false;
  export let fadeUnselected = false;
  export let showXForSelected = false;
  export let showSelection = false;
  export let searchText = "";
  export let showHiddenSelectionsCount = false;
  export let onSelect: (name: string) => void;
  export let onToggleSelectAll: () => void = voidFn;

  $: allSelected = selectableGroups.every((g, i) => {
    return (
      selectedItems[i]?.length && g.items.length === selectedItems[i].length
    );
  });

  $: selectedCount = selectedItems.flat().length;

  $: singleSelection = selectedCount === 1;

  $: numSelectedShown = selectableGroups.reduce(
    (sel, { items }, groupIndex) => {
      return (
        sel +
        items.filter(({ name }) => selectedItems[groupIndex]?.includes(name))
          .length
      );
    },
    0,
  );

  $: numSelectedNotShown = selectedCount - numSelectedShown;

  $: filteredGroups = filterGroups(selectableGroups, searchText);

  function filterGroups(
    selectableGroups: SearchableFilterSelectableGroup[],

    searchText: string,
  ) {
    if (!searchText) return selectableGroups;

    return selectableGroups.map((group) => {
      return {
        ...group,
        items: matchSorter(group.items, searchText, { keys: ["label"] }),
      };
    });
  }
</script>

<DropdownMenu.Content
  align="start"
  class="flex flex-col max-h-96 w-72 overflow-hidden p-0"
>
  <div class="px-3 pt-3 pb-1">
    <Search
      bind:value={searchText}
      label="Search list"
      showBorderOnFocus={false}
    />
  </div>

  <div class="flex flex-col flex-1 overflow-y-auto w-full h-fit pb-1">
    {#each filteredGroups as { name, items }, index (name)}
      <DropdownMenu.Group class="px-1">
        {#if filteredGroups.length > 1}
          <DropdownMenu.Label>
            {name}
          </DropdownMenu.Label>
        {/if}
        {#each items as { name, label } (name)}
          {@const selected = selectedItems[index]?.includes(name)}

          <svelte:component
            this={allowMultiSelect || showSelection
              ? DropdownMenu.CheckboxItem
              : DropdownMenu.Item}
            {...allowMultiSelect || showSelection
              ? { checked: selected, showXForSelected }
              : {}}
            class="text-xs cursor-pointer"
            role="menuitem"
            disabled={requireSelection && singleSelection && selected}
            aria-disabled={requireSelection && singleSelection && selected}
            on:click={() => {
              if (requireSelection && singleSelection && selected) return;

              onSelect(name);
            }}
          >
            <span
              class:ui-copy-disabled={fadeUnselected &&
                !selected &&
                allowMultiSelect}
            >
              {#if label.length > 240}
                {label.slice(0, 240)}...
              {:else}
                {label}
              {/if}
            </span>
          </svelte:component>
        {:else}
          <div class="ui-copy-disabled text-center p-2 w-full">no results</div>
        {/each}

        {#if index !== filteredGroups.length - 1}
          <DropdownMenu.Separator />
        {/if}
      </DropdownMenu.Group>
    {/each}
  </div>

  {#if allowMultiSelect}
    <footer>
      <Button on:click={onToggleSelectAll} type="text">
        {#if allSelected}
          Deselect all
        {:else}
          Select all
        {/if}
      </Button>

      <slot name="action" />
      {#if numSelectedNotShown && showHiddenSelectionsCount}
        <div class="ui-label">
          {numSelectedNotShown} other value{numSelectedNotShown > 1 ? "s" : ""} selected
        </div>
      {/if}
    </footer>
  {/if}
</DropdownMenu.Content>

<style lang="postcss">
  footer {
    height: 42px;
    @apply flex flex-row mt-auto items-center justify-end;
    @apply bg-slate-100;
    @apply border-t border-slate-300;
    @apply bottom-0 sticky;
    @apply gap-x-2 p-2 px-3.5;
  }

  footer:is(.dark) {
    @apply bg-gray-800;
    @apply border-gray-700;
  }
</style>
