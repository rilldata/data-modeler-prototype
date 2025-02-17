<script lang="ts">
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import type { CanvasComponentType } from "./components/types";
  import { dropZone, hoveredDivider, activeDivider } from "./stores/ui-stores";

  export let allowDrop: boolean;
  export let resizeIndex = -1;
  export let dropIndex: number;
  export let onDrop: (row: number, column: number | null) => void;
  export let onRowResizeStart: (e: MouseEvent) => void;
  export let addItem: (type: CanvasComponentType) => void;

  $: dividerId = `row:${resizeIndex}::column:null`;
  $: dropId = `row:${dropIndex}::column:null`;

  $: isActiveDivider = $activeDivider === dividerId;
  $: isDropZone = $dropZone === dropId;
  $: isHoveredDivider = $hoveredDivider === dividerId;

  $: notActiveDivider = !!$activeDivider && !isActiveDivider;

  $: showAddComponent = isHoveredDivider && !isActiveDivider;
  $: notResizable = resizeIndex === -1;

  function focus() {
    activeDivider.set(dividerId);
  }

  function hover() {
    hoveredDivider.set(dividerId);
  }
</script>

<div
  role="presentation"
  style:pointer-events={allowDrop ? "auto" : "none"}
  style:width="calc(100% + 160px)"
  class:top={notResizable}
  class:bottom={!notResizable}
  class="absolute z-10 -left-20 h-20 flex items-center justify-center px-2"
  on:mouseenter={() => {
    if (!allowDrop) return;
    dropZone.set(dropId);
  }}
  on:mouseleave={() => {
    if (!allowDrop) return;
    dropZone.clear();
  }}
  on:mouseup={() => {
    if (!allowDrop) return;
    onDrop(dropIndex, null);
  }}
>
  <button
    data-row={resizeIndex}
    class:cursor-default={notResizable && !allowDrop}
    class:cursor-row-resize={!allowDrop && !notResizable}
    style:pointer-events={notActiveDivider || allowDrop ? "none" : "auto"}
    class="mx-20 w-full h-4 z-40 group flex items-center justify-center"
    on:mousedown={(e) => {
      onRowResizeStart(e);
      focus();
    }}
    on:mouseenter={hover}
    on:mouseleave={hoveredDivider.reset}
  >
    <span
      class:bg-primary-300={isActiveDivider || isHoveredDivider || isDropZone}
      class="w-full h-[3px] rounded-full pointer-events-none"
    />
  </button>

  {#if showAddComponent}
    <span
      role="presentation"
      on:mouseleave={hoveredDivider.reset}
      class:shift-down={notResizable}
      class="flex shadow-sm pointer-events-auto absolute left-1/2 w-fit z-50 bg-white -translate-x-1/2 border rounded-sm"
    >
      <AddComponentDropdown
        onMouseEnter={hover}
        onItemClick={(type) => {
          hoveredDivider.reset();
          addItem(type);
        }}
      />
    </span>
  {/if}
</div>

<style lang="postcss">
  /* div {
    @apply bg-blue-400/20;
  } */

  .top {
    @apply top-0  -translate-y-1/2;
  }

  .bottom {
    @apply bottom-0 translate-y-1/2;
  }

  .shift-down {
    @apply translate-y-[8px];
  }
</style>
