<script lang="ts">
  import { page } from "$app/stores";
  import ContextButton from "@rilldata/web-common/components/button/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { NavDragData } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { getPaddingFromPath } from "@rilldata/web-common/features/file-explorer/nav-tree-spacing";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
    ResourceKindToScreenMap,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import type { V1ResourceName } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { Save } from "lucide-svelte";
  import type { Readable } from "svelte/store";
  import File from "../../components/icons/File.svelte";
  import NavigationMenuSeparator from "../../layout/navigation/NavigationMenuSeparator.svelte";
  import { fileArtifacts } from "../entity-management/file-artifacts";
  import { getTopLevelFolder } from "@rilldata/utils";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "../entity-management/resource-icon-mapping";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import MetricsViewMenuItems from "../metrics-views/MetricsViewMenuItems.svelte";
  import ModelMenuItems from "../models/navigation/ModelMenuItems.svelte";
  import SourceMenuItems from "../sources/navigation/SourceMenuItems.svelte";
  import { PROTECTED_DIRECTORIES, PROTECTED_FILES } from "./protected-paths";
  import { Settings } from "lucide-svelte";

  export let filePath: string;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string, isDir: boolean) => void;
  export let onGenerateChart: (data: {
    table?: string;
    connector?: string;
    metricsView?: string;
  }) => void;
  export let onMouseDown: (e: MouseEvent, dragData: NavDragData) => void;

  let contextMenuOpen = false;
  let resourceName: Readable<V1ResourceName | undefined>;

  $: id = `${filePath}-nav-link`;
  $: fileName = filePath.split("/").pop();
  $: isCurrentFile =
    removeLeadingSlash(filePath) ===
    removeLeadingSlash($page.params.file ?? "");
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: ({
    resourceName,
    hasUnsavedChanges,
    saveLocalContent,
    inferredResourceKind,
  } = fileArtifact);
  $: resourceKind = ($resourceName?.kind ??
    $inferredResourceKind) as ResourceKind;
  $: padding = getPaddingFromPath(filePath);
  $: topLevelFolder = getTopLevelFolder(filePath);
  $: isProtectedDirectory = PROTECTED_DIRECTORIES.includes(topLevelFolder);
  $: isDotFile = fileName && fileName.startsWith(".");
  $: isProtectedFile = PROTECTED_FILES.includes(filePath);

  $: hasErrors = fileArtifact.getHasErrors(queryClient, $runtime.instanceId);

  function fireTelemetry() {
    const previousScreenName = getScreenNameFromPage();
    behaviourEvent
      .fireNavigationEvent(
        filePath,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        previousScreenName,
        ResourceKindToScreenMap[resourceKind] ?? MetricsEventScreenName.Unknown,
      )
      .catch(console.error);
  }

  function handleMouseDown(e: MouseEvent) {
    if (PROTECTED_FILES.includes(filePath)) return;
    onMouseDown(e, { id, filePath, isDir: false, kind: resourceKind });
  }
</script>

<li
  aria-label="{filePath} Nav Entry"
  class="w-full text-left pr-2 h-6 group flex justify-between gap-x-1 items-center hover:bg-slate-100"
  class:bg-slate-100={isCurrentFile}
  class:opacity-50={$hasUnsavedChanges}
>
  <a
    class="w-full truncate flex items-center gap-x-1 font-medium {isProtectedDirectory ||
    isDotFile
      ? 'text-gray-500 hover:text-gray-500'
      : 'text-gray-900 hover:text-gray-900'}"
    href={`/files${filePath}`}
    {id}
    class:italic={$hasUnsavedChanges}
    on:click={fireTelemetry}
    on:mousedown={handleMouseDown}
    style:padding-left="{padding}px"
  >
    <div class="flex-none">
      <svelte:component
        this={resourceKind
          ? resourceIconMapping[resourceKind]
          : filePath === "/.env" || filePath === "/rill.yaml"
            ? Settings
            : File}
        size="14px"
        color={resourceKind ? resourceColorMapping[resourceKind] : "#9CA3AF"}
      />
    </div>
    <span class="truncate w-full" class:text-red-600={$hasErrors}>
      {fileName}
    </span>
  </a>
  {#if !isProtectedDirectory && !isProtectedFile}
    <DropdownMenu.Root bind:open={contextMenuOpen}>
      <DropdownMenu.Trigger asChild let:builder>
        <ContextButton
          builders={[builder]}
          id="more-actions-{filePath}"
          label="{filePath} actions menu trigger"
          suppressTooltip={contextMenuOpen}
          tooltipText="More actions"
        >
          <MoreHorizontal />
        </ContextButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        align="start"
        class="border-none bg-gray-800 text-white min-w-60"
        side="right"
        sideOffset={16}
      >
        {#if resourceKind}
          {#if resourceKind === ResourceKind.Source}
            <SourceMenuItems
              {filePath}
              on:generate-chart={({ detail }) => onGenerateChart(detail)}
            />
            <NavigationMenuSeparator />
          {:else if resourceKind === ResourceKind.Model}
            <ModelMenuItems
              {filePath}
              on:generate-chart={({ detail }) => onGenerateChart(detail)}
            />
            <NavigationMenuSeparator />
          {:else if resourceKind === ResourceKind.MetricsView}
            <MetricsViewMenuItems
              {filePath}
              on:generate-chart={({ detail }) => onGenerateChart(detail)}
            />
            <NavigationMenuSeparator />
          {/if}
        {/if}
        {#if $hasUnsavedChanges}
          <NavigationMenuItem on:click={saveLocalContent}>
            <Save slot="icon" size="12px" />
            Save file
          </NavigationMenuItem>
        {/if}
        <NavigationMenuItem on:click={() => onRename(filePath, false)}>
          <EditIcon slot="icon" />
          Rename...
        </NavigationMenuItem>
        <NavigationMenuItem on:click={() => onDelete(filePath, false)}>
          <Cancel slot="icon" />
          Delete
        </NavigationMenuItem>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</li>
