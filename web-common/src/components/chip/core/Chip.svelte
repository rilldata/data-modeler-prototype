<script lang="ts">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import { createEventDispatcher, getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import CancelCircle from "../../icons/CancelCircle.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";
  import CaretDownIcon from "../../icons/CaretDownIcon.svelte";
  import { slide } from "svelte/transition";

  export let removable = false;
  export let active = false;
  export let readOnly = false;
  export let type: "measure" | "dimension" | "time" | "special" = "dimension";
  export let exclude = false;
  export let grab = false;
  export let builders: Builder[] = [];
  export let caret = builders.length > 0;
  export let slideDuration = 150;
  export let supressTooltip = false;
  export let label: string | undefined = undefined;
  export let removeTooltipText: string | undefined = undefined;

  const dispatch = createEventDispatcher();

  const tooltipSuppression = getContext<Writable<boolean>>(
    "rill:app:childRequestedTooltipSuppression",
  );

  function focusOnRemove() {
    if (tooltipSuppression) tooltipSuppression.set(true);
  }
  function blurOnRemove() {
    if (tooltipSuppression) tooltipSuppression.set(false);
  }
</script>

<div in:slide={{ axis: "x", duration: slideDuration }}>
  <div
    class="chip {type}"
    class:active
    class:grab
    class:exclude
    class:pointer-events-none={readOnly}
    aria-label={label}
    {...getAttrs(builders)}
    use:builderActions={{ builders }}
  >
    {#if removable && !readOnly}
      <Tooltip
        alignment="start"
        distance={12}
        suppress={supressTooltip || !removeTooltipText}
      >
        <button
          class="text-inherit mr-0.5"
          aria-label="Remove"
          on:mouseover={focusOnRemove}
          on:focus={focusOnRemove}
          on:mouseleave={blurOnRemove}
          on:blur={blurOnRemove}
          on:click|stopPropagation={() => dispatch("remove")}
        >
          <CancelCircle size="16px" />
        </button>

        <TooltipContent maxWidth="300px" slot="tooltip-content">
          {removeTooltipText}
        </TooltipContent>
      </Tooltip>
    {/if}

    {#if $$slots.body}
      <button
        on:click
        on:mousedown
        aria-label={label}
        class="text-inherit w-full select-none flex items-center gap-x-1 px-0.5"
      >
        <slot name="body" />

        {#if caret}
          <span class="transition-transform -mr-0.5" class:-rotate-180={active}>
            <CaretDownIcon size="10px" />
          </span>
        {/if}
      </button>
    {/if}
  </div>
</div>

<style lang="postcss">
  .grab {
    @apply cursor-grab;
  }

  .chip {
    @apply flex flex-none gap-x-1;
    @apply items-center justify-center;
    @apply px-2 py-[3px] border w-fit;
  }

  .dimension {
    @apply rounded-2xl;
    @apply bg-primary-50 border-primary-200 text-primary-800;
  }

  .dimension:hover,
  .dimension:active,
  .dimension.active {
    @apply bg-primary-100;
  }

  .dimension:active,
  .dimension.active {
    @apply border-primary-400;
  }

  .measure {
    @apply rounded-sm;
    @apply bg-secondary-50 border-secondary-200 text-secondary-800;
  }

  .measure:hover,
  .measure:active,
  .measure.active {
    @apply bg-secondary-100;
  }

  .measure:active,
  .measure.active {
    @apply border-secondary-400;
  }

  .exclude {
    @apply bg-gray-50 border-gray-200 text-gray-600;
  }

  .exclude:hover,
  .exclude:active,
  .exclude.active {
    @apply bg-gray-100;
  }

  .exclude:active,
  .exclude.active {
    @apply border-gray-400;
  }

  .time {
    @apply rounded-2xl;
    @apply bg-white border-slate-200 text-slate-800;
  }

  .time:hover,
  .time.active,
  .time:active {
    @apply bg-slate-50;
  }

  .time.active,
  .time:active {
    @apply border-slate-400;
  }
</style>
