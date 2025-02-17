<script lang="ts">
  export let zIndex = 50;
  export let maxWidth: number;
  export let gridTemplate: string;
</script>

<section
  role="presentation"
  class="size-full grid min-h-fit canvas-row relative"
  style:z-index={zIndex}
  style:max-width="{maxWidth}px"
  style:grid-template-columns={gridTemplate}
>
  <slot />
</section>

<style lang="postcss">
  .canvas-row {
    /* height: var(--row-height); */
    grid-auto-rows: max-content;
  }
</style>
