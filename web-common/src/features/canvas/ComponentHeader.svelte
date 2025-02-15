<script lang="ts">
  import type { ComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/types";
  import LocalFiltersHeader from "@rilldata/web-common/features/canvas/LocalFiltersHeader.svelte";

  export let title: string | undefined = undefined;
  export let description: string | undefined = undefined;
  export let filters: ComponentFilterProperties;
  export let faint: boolean = false;

  $: atleastOneFilter = Boolean(
    filters?.time_filters || filters?.dimension_filters,
  );
</script>

{#if title || description}
  <div class="w-full h-fit flex flex-col bg-white px-2 pt-2 pb-1 items-start">
    {#if title}
      <div class="flex items-center gap-x-2">
        <h1 class:faint class="title">{title}</h1>
        <LocalFiltersHeader {atleastOneFilter} {filters} />
      </div>
    {/if}
    {#if description}
      <div class="flex items-center gap-x-2">
        <h2 class="description">{description}</h2>
        {#if !title}
          <LocalFiltersHeader {atleastOneFilter} {filters} />
        {/if}
      </div>
    {/if}
  </div>
{:else if atleastOneFilter}
  <div class="absolute top-0 left-0 z-50 pl-1 pt-1">
    <LocalFiltersHeader {atleastOneFilter} {filters} />
  </div>
{/if}

<style lang="postcss">
  .title {
    font-size: 15px;
    line-height: 26px;
    @apply font-medium text-gray-700 truncate;
  }

  .title.faint {
    @apply text-gray-500;
  }

  .description {
    font-size: 13px;
    @apply text-gray-500 font-normal leading-none;
  }
</style>
