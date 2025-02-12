<script lang="ts">
  export let range: string | undefined;
  export let onClick: ((range: string | undefined) => void) | undefined =
    undefined;
</script>

<svelte:element
  this={onClick ? "button" : "span"}
  role={onClick ? "button" : undefined}
  class="element"
  on:click={() => onClick?.(range)}
>
  {range}
</svelte:element>

<style lang="postcss">
  .element {
    @apply bg-slate-100 text-slate-500 rounded-[2px] px-1 line-clamp-1 truncate flex-none h-5 flex items-center;
    font-family: "Source Code Variable", monospace;
    @apply font-medium w-fit;
  }

  .element:hover {
    @apply bg-slate-200;
  }
</style>
