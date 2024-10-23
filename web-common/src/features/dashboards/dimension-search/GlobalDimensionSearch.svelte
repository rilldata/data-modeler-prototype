<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import GlobalDimensionSearchResults from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearchResults.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import SearchIcon from "@rilldata/web-common/components/icons/Search.svelte";
  import { slide } from "svelte/transition";

  const { dimensionSearch } = featureFlags;

  let searchBarOpen = false;
  let searchText = "";

  function reset() {
    searchBarOpen = false;
  }

  let submittedSearchText = "";
  let searchResultsOpen = false;
  function onSubmit() {
    submittedSearchText = searchText;
    searchResultsOpen = true;
  }
</script>

{#if $dimensionSearch}
  <div class="relative flex flex-row">
    {#if searchBarOpen}
      <div
        in:slide={{ axis: "x", duration: 200 }}
        class="flex items-center gap-x-2 pr-2 w-60 bg-slate-50 border border-primary-300"
      >
        <Search
          bind:value={searchText}
          on:submit={onSubmit}
          placeholder="Search dimensions"
          autofocus
          border={false}
          background={false}
        />
        <button class="ui-copy-icon" on:click={reset}>
          <Cancel size="16px" />
        </button>
      </div>
    {:else}
      <Button
        class="flex items-center gap-x-2 p-1.5 text-gray-700"
        on:click={() => (searchBarOpen = !searchBarOpen)}
        type="secondary"
        compact
      >
        <SearchIcon size="16px" />
      </Button>
    {/if}

    <GlobalDimensionSearchResults
      searchText={submittedSearchText}
      onSelect={reset}
      bind:open={searchResultsOpen}
    />
  </div>
{/if}
