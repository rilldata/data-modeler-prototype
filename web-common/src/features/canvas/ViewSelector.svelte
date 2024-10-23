<script lang="ts">
  import Viz from "@rilldata/web-common/components/icons/Viz.svelte";
  import Split from "@rilldata/web-common/components/icons/Split.svelte";
  import Code from "@rilldata/web-common/components/icons/Code.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { portal } from "@rilldata/actions";

  const viewOptions = [
    { view: "code", icon: Code },
    { view: "split", icon: Split },
    { view: "viz", icon: Viz },
  ];

  export let selectedView: "code" | "split" | "viz" = "split";
  export let allowSplit = true;

  let tooltip: { view: string; top: number; left: number } | null = null;
</script>

<div class="radio" role="radiogroup">
  {#each viewOptions as { view, icon: Icon } (view)}
    {#if (view === "split" && allowSplit) || view !== "split"}
      <input
        tabindex="0"
        type="radio"
        id={view}
        name="view"
        class="screenreader-only"
        value={view}
        checked={view === selectedView}
        bind:group={selectedView}
        on:click={() => {
          tooltip = null;
        }}
      />

      <label
        aria-label={view}
        for={view}
        on:pointerleave={() => {
          tooltip = null;
        }}
        on:pointerenter={(e) => {
          const { top, left, width, height } =
            e.currentTarget.getBoundingClientRect();
          tooltip = {
            view,
            top: top + height + 8,
            left: left + width / 2,
          };
        }}
      >
        <Icon size="15px" class="fill-primary-600" />
      </label>
    {/if}
  {/each}
</div>

{#if tooltip}
  <div
    use:portal
    class="absolute -translate-x-1/2"
    style:top="{tooltip.top}px"
    style:left="{tooltip.left}px"
  >
    <TooltipContent slot="tooltip-content">{tooltip.view}</TooltipContent>
  </div>
{/if}

<style lang="postcss">
  label {
    @apply w-7 aspect-square;
    @apply border-r border-primary-300 cursor-pointer;
    @apply flex items-center justify-center;
  }

  .screenreader-only {
    position: absolute;
    clip: rect(1px, 1px, 1px, 1px);
    padding: 0;
    border: 0;
    height: 1px;
    width: 1px;
    overflow: hidden;
  }

  label:last-of-type {
    @apply border-0;
  }

  label:hover {
    @apply bg-primary-50;
  }

  input:checked + label {
    @apply bg-primary-200;
  }

  .radio {
    @apply flex border-primary-300 rounded-[2px] border w-fit h-7 items-center justify-center overflow-hidden;
  }
</style>
