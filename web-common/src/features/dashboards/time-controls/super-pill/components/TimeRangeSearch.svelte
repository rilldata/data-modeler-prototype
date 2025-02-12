<script lang="ts">
  import SyntaxElement from "./SyntaxElement.svelte";
  import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
  import { Clock } from "lucide-svelte";

  export let context: string;
  export let onSelectRange: (range: string, isSearch: boolean) => void;

  let searchValue = "";
  let searchElement: HTMLInputElement;

  const latestNSearches = localStorageStore(`${context}-recent-searches`, [
    "-45M",
    "-32D",
    "-1Y",
    "-2Q, latest/Q",
  ]);

  export function updateSearch(value: string) {
    searchValue = value;
    searchElement.focus();
  }
</script>

<div
  class="border-b h-fit pt-2.5 py-0 flex p-3 flex-col w-full overflow-y-auto"
>
  <form
    class="mb-2.5"
    on:submit={() => {
      latestNSearches.update((searches) => {
        return Array.from(new Set([searchValue, ...searches].slice(0, 20)));
      });
      onSelectRange(searchValue, true);
      searchValue = "";
    }}
  >
    <span class="mr-1 flex-none">
      <Clock size={15} />
    </span>
    <input
      placeholder="Search"
      type="text"
      class="h-7 border w-full"
      bind:this={searchElement}
      bind:value={searchValue}
    />
  </form>

  <div class="flex gap-x-2 size-full overflow-x-auto pb-2.5">
    {#each $latestNSearches as search, i (i)}
      <SyntaxElement range={search} onClick={updateSearch} />
    {/each}
  </div>
</div>

<style lang="postcss">
  form {
    @apply overflow-hidden;
    @apply flex justify-center gap-x-1 items-center pl-2 pr-0.5;
    @apply bg-background justify-center;
    @apply border border-gray-300 rounded-[2px];
    @apply cursor-pointer;
    @apply h-7 w-full truncate;
  }

  form:focus-within {
    @apply border-primary-500;
  }

  input {
    @apply p-0 bg-transparent;
    @apply size-full;
    @apply outline-none border-0;
    @apply cursor-text;
    vertical-align: middle;
  }
</style>
