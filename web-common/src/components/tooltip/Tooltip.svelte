<!-- @component 
The Tooltip component is a workhorse generalized-render component wrapper.
By default, it will take the first DOM element in the <slot /> and 
listen to mouseover changes to it to trigger the tooltip.

You can set the following parameters:
- location: left, right, top, bottom.
- alignment: start, middle, end. This will align using flexbox-like language.
  - start on left or right means that the top of the tooltip aligns with the top of the target.
  - start on top or  bottom means that the left side of the tooltip aligns with the left side of the target.

The suppress prop gives programmatic access to suppressing the tooltip, e.g. if a menu is open or you just need
to stop all tooltips.

This component also passes down a boolean writable store to its children, so a child can suppress the tooltip
without the consumer needing to figure out the logic to suppress in the parent component.

FIXME: In the future, we should also be listening to focus events from the child.
-->
<script lang="ts">
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import FloatingElement from "../floating-element/FloatingElement.svelte";
  import type {
    Alignment,
    Location,
  } from "@rilldata/web-common/lib/place-element";

  export let location: Location = "bottom";
  export let alignment: Alignment = "middle";
  export let distance = 8;
  export let pad = 8;
  // provide a programmatic guard to suppressing the tooltip.
  export let suppress = false;
  /** the delay in miliseconds before rendering the tooltip once mouse has entered */
  export let activeDelay = 120;
  /** the delay in miliseconds before unrendering the tooltip once mouse has left */
  export let nonActiveDelay = 0;
  export let active = false;

  let parent: HTMLDivElement;
  let waitUntilTimer: ReturnType<typeof setTimeout>;

  function waitUntil(callback, time = 120) {
    if (waitUntilTimer) clearTimeout(waitUntilTimer);
    waitUntilTimer = setTimeout(callback, time);
  }

  /** create child-supported suppression.
   * If a child changes this context store to true, we should
   * suppress the tooltip.
   * This enables us to disentangle the tooltip state in certain cases
   * where it doesn't make sense to have the user deal with the logic of suppression.
   */
  const childRequestedTooltipSuppression = writable(false);
  setContext(
    "rill:app:childRequestedTooltipSuppression",
    childRequestedTooltipSuppression,
  );
</script>

<div
  role="tooltip"
  class="contents"
  bind:this={parent}
  on:mouseenter={() => {
    waitUntil(() => {
      active = true;
    }, activeDelay);
  }}
  on:mouseleave={() => {
    waitUntil(() => {
      active = false;
    }, nonActiveDelay);
  }}
>
  <slot />
  {#if active && !suppress && !$childRequestedTooltipSuppression}
    <FloatingElement target={parent} {location} {alignment} {distance} {pad}>
      <slot name="tooltip-content" />
    </FloatingElement>
  {/if}
</div>
