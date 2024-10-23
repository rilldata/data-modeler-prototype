<script lang="ts">
  import PivotChip from "./PivotChip.svelte";
  import { type PivotChipData, PivotChipType } from "./types";
  import { portal } from "@rilldata/actions";
  import { dragDataStore } from "./DragList.svelte";

  export let item: PivotChipData;
  export let removable: boolean;
  export let offset: { x: number; y: number };
  export let position = { left: 0, top: 0 };

  function trackDragItem(e: MouseEvent) {
    requestAnimationFrame(() => {
      position = {
        left: e.clientX - offset.x,
        top: e.clientY - offset.y,
      };
    });
  }

  function onDragRelease() {
    dragDataStore.set(null);
  }
</script>

<svelte:window on:mousemove={trackDragItem} on:mouseup={onDragRelease} />

<div
  class="portal-item"
  class:rounded-full={item.type !== PivotChipType.Measure}
  style:left="{position.left}px"
  style:top="{position.top}px"
  use:portal
>
  <PivotChip
    active
    slideDuration={0}
    grab
    {item}
    {removable}
    on:mousedown
    on:remove
  />
</div>

<style lang="postcss">
  .portal-item {
    @apply shadow-lg shadow-slate-300;
    @apply z-50;
    @apply absolute pointer-events-none;
  }
</style>
