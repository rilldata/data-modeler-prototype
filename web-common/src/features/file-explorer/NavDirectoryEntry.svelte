<script lang="ts">
  import ContextButton from "@rilldata/web-common/components/button/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
  import type { NavDragData } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { getPaddingFromPath } from "@rilldata/web-common/features/file-explorer/nav-tree-spacing";
  import {
    type Directory,
    getDirectoryHasErrors,
  } from "@rilldata/web-common/features/file-explorer/transform-file-list";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { Folder } from "lucide-svelte";
  import { createRuntimeServiceCreateDirectory } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { removeLeadingSlash } from "../entity-management/entity-mappers";
  import { useDirectoryNamesInDirectory } from "../entity-management/file-selectors";
  import { getName, getTopLevelFolder } from "@rilldata/utils";
  import { PROTECTED_DIRECTORIES } from "./protected-paths";

  export let dir: Directory;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string, isDir: boolean) => void;
  export let onMouseDown: (e: MouseEvent, dragData: NavDragData) => void;

  let contextMenuOpen = false;

  const createFolder = createRuntimeServiceCreateDirectory();

  $: id = `${dir.path}-nav-entry`;
  $: expanded = $directoryState[dir.path];
  $: padding = getPaddingFromPath(dir.path);
  $: instanceId = $runtime.instanceId;
  $: topLevelFolder = getTopLevelFolder(dir.path);
  $: isProtectedDirectory = PROTECTED_DIRECTORIES.includes(topLevelFolder);

  $: hasErrors = getDirectoryHasErrors(queryClient, instanceId, dir);

  $: currentDirectoryDirectoryNamesQuery = useDirectoryNamesInDirectory(
    instanceId,
    dir.path,
  );

  function toggleDirectory(directory: Directory): void {
    directoryState.toggle(directory.path);
  }

  /**
   * Put a folder in the current directory
   */
  async function handleAddFolder() {
    const nextFolderName = getName(
      "untitled_folder",
      $currentDirectoryDirectoryNamesQuery?.data ?? [],
    );

    const path =
      dir.path !== ""
        ? `${removeLeadingSlash(dir.path)}/${nextFolderName}`
        : nextFolderName;

    await $createFolder.mutateAsync({
      instanceId: instanceId,
      data: {
        path: path,
      },
    });

    // Expand the directory to show the new folder
    const pathWithLeadingSlash = `/${path}`;
    directoryState.expand(pathWithLeadingSlash);
  }
</script>

<button
  class="pr-2 w-full h-6 text-left flex justify-between group gap-x-1 items-center
  {isProtectedDirectory
    ? 'text-gray-500'
    : 'text-gray-900 hover:text-gray-900'} 
  font-medium hover:bg-slate-100"
  {id}
  on:click={() => toggleDirectory(dir)}
  on:mousedown={(e) => onMouseDown(e, { id, filePath: dir.path, isDir: true })}
  style:padding-left="{padding}px"
  aria-controls={`nav-${dir.path}`}
  aria-expanded={expanded}
>
  <CaretDownIcon
    className="flex-none text-gray-400 {expanded ? '' : 'transform -rotate-90'}"
    size="14px"
  />
  <span class="truncate w-full" class:text-red-600={$hasErrors}>
    {dir.name}
  </span>
  {#if !isProtectedDirectory}
    <DropdownMenu.Root bind:open={contextMenuOpen}>
      <DropdownMenu.Trigger asChild let:builder>
        <ContextButton
          builders={[builder]}
          id="more-actions-{dir.path}"
          label="{dir.name} actions menu trigger"
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
        <NavigationMenuItem on:click={handleAddFolder}>
          <Folder slot="icon" size="12px" />
          New folder
        </NavigationMenuItem>
        <NavigationMenuItem on:click={() => onRename(dir.path, true)}>
          <EditIcon slot="icon" />
          Rename...
        </NavigationMenuItem>
        <NavigationMenuItem on:click={() => onDelete(dir.path, true)}>
          <Cancel slot="icon" />
          Delete
        </NavigationMenuItem>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</button>
