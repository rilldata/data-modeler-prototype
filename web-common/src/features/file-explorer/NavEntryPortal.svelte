<script lang="ts">
  import File from "@rilldata/web-common/components/icons/File.svelte";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import type { NavDragData } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { portal } from "@rilldata/actions";
  import { Folder } from "lucide-svelte";

  export let position = { left: 0, top: 0 };
  export let dragData: NavDragData;
</script>

<div
  class="portal-item"
  style:left="{position.left}px"
  style:top="{position.top}px"
  use:portal
>
  <div class="flex flex-row gap-x-1">
    {#if dragData.isDir}
      <Folder className="text-gray-400" size="14px" />
    {:else}
      <svelte:component
        this={dragData.kind ? resourceIconMapping[dragData.kind] : File}
        className="text-gray-400"
        size="14px"
      />
    {/if}
    <span>{dragData.fileName ?? ""}</span>
  </div>
</div>

<style lang="postcss">
  .portal-item {
    @apply shadow-lg shadow-slate-300;
    @apply z-50;
    @apply absolute pointer-events-none;
  }
</style>
