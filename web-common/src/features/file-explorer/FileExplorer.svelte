<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import GenerateChartYAMLPrompt from "@rilldata/web-common/features/canvas-components/prompt/GenerateChartYAMLPrompt.svelte";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import {
    deleteFileArtifact,
    renameFileArtifact,
  } from "@rilldata/web-common/features/entity-management/actions";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getTopLevelFolder, splitFolderAndFileName } from "@rilldata/utils";
  import ForceDeleteConfirmation from "@rilldata/web-common/features/file-explorer/ForceDeleteConfirmationDialog.svelte";
  import NavEntryPortal from "@rilldata/web-common/features/file-explorer/NavEntryPortal.svelte";
  import { navEntryDragDropStore } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { PROTECTED_DIRECTORIES } from "@rilldata/web-common/features/file-explorer/protected-paths";
  import { isCurrentActivePage } from "@rilldata/web-common/features/file-explorer/utils";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { eventBus } from "@rilldata/events";
  import { fileArtifacts } from "../entity-management/file-artifacts";
  import NavDirectory from "./NavDirectory.svelte";
  import { findDirectory, transformFileList } from "./transform-file-list";

  export let hasUnsaved: boolean;

  $: instanceId = $runtime.instanceId;
  $: getFileTree = createRuntimeServiceListFiles(instanceId, undefined, {
    query: {
      select: (data) => {
        if (!data || !data.files?.length) return;

        const files = data.files
          // Sort alphabetically case-insensitive
          .sort(
            (a, b) =>
              a.path?.localeCompare(b.path ?? "", undefined, {
                sensitivity: "base",
              }) ?? 0,
          )
          // Hide dot directories
          .filter(
            (file) =>
              !(
                file.isDir &&
                // Check both the top-level directory and subdirectories
                (file.path?.startsWith(".") || file.path?.includes("/."))
              ),
          )
          // Hide the `tmp` directory
          .filter((file) => !file.path?.startsWith("/tmp"));

        return transformFileList(files);
      },
    },
  });

  $: ({ data: fileTree } = $getFileTree);

  let showRenameModelModal = false;
  let renameFilePath: string;
  let renameIsDir: boolean;

  function onRename(filePath: string, isDir: boolean) {
    showRenameModelModal = true;
    renameFilePath = filePath;
    renameIsDir = isDir;
  }

  let forceDeletePath: string;
  let showForceDelete = false;

  async function onDelete(filePath: string, isDir: boolean) {
    if (!$getFileTree.data) return;

    if (isDir) {
      const dir = findDirectory($getFileTree.data, filePath);
      if (dir?.directories?.length || dir?.files?.length) {
        forceDeletePath = filePath;
        showForceDelete = true;
        return;
      }
    }
    await deleteFileArtifact(instanceId, filePath);
    if (isCurrentActivePage(filePath, isDir)) {
      await goto("/");
    }
  }

  async function onForceDelete() {
    await deleteFileArtifact(instanceId, forceDeletePath, true);
    // onForceDelete is only called on folders, so isDir is always true
    if (isCurrentActivePage(forceDeletePath, true)) {
      await goto("/");
    }
  }

  let showGenerateChartModal = false;
  let generateChartTable: string;
  let generateChartConnector: string;
  let generateChartMetricsView: string;

  function onGenerateChart({
    table,
    connector,
    metricsView,
  }: {
    table?: string;
    connector?: string;
    metricsView?: string;
  }) {
    showGenerateChartModal = true;
    generateChartTable = table ?? "";
    generateChartConnector = connector ?? "";
    generateChartMetricsView = metricsView ?? "";
  }

  const { dragData, position } = navEntryDragDropStore;

  async function handleDropSuccess(fromPath: string, toDir: string) {
    const isCurrentFile =
      $page.params.file && // handle case when user is on home page
      removeLeadingSlash(fromPath) === removeLeadingSlash($page.params.file);
    const [, srcFile] = splitFolderAndFileName(fromPath);
    const newFilePath = `${toDir === "/" ? toDir : toDir + "/"}${srcFile}`;

    if (fromPath !== newFilePath) {
      const newTopLevelPath = getTopLevelFolder(newFilePath);
      if (PROTECTED_DIRECTORIES.includes(newTopLevelPath)) {
        eventBus.emit("notification", {
          message: "cannot move to restricted directories",
        });
        return;
      }
      await renameFileArtifact(instanceId, fromPath, newFilePath);

      if (isCurrentFile) {
        await goto(`/files${newFilePath}`);
      }
    }
  }

  async function saveAll(e: KeyboardEvent) {
    if (e.code === "KeyS" && e.metaKey && e.altKey) {
      e.preventDefault();
      await fileArtifacts.saveAll();
    }
  }
</script>

<svelte:window
  on:beforeunload={(event) => {
    if (hasUnsaved) {
      event.preventDefault();
      return confirm(
        "Are you sure you want to leave? Unsaved changes will be lost.",
      );
    }
  }}
  on:mousemove={(e) => navEntryDragDropStore.onMouseMove(e)}
  on:mouseup={(e) => navEntryDragDropStore.onMouseUp(e, handleDropSuccess)}
  on:keydown={saveAll}
/>

<!-- File tree -->
<ul class="flex flex-col w-full items-start justify-start overflow-auto">
  {#if fileTree}
    <NavDirectory
      directory={fileTree}
      {onRename}
      {onDelete}
      {onGenerateChart}
      onMouseDown={(e, dragData) =>
        navEntryDragDropStore.onMouseDown(e, dragData)}
    />
  {/if}
</ul>

{#if showRenameModelModal}
  <RenameAssetModal
    closeModal={() => (showRenameModelModal = false)}
    filePath={renameFilePath}
    isDir={renameIsDir}
  />
{/if}

<GenerateChartYAMLPrompt
  bind:open={showGenerateChartModal}
  connector={generateChartConnector}
  metricsView={generateChartMetricsView}
  table={generateChartTable}
/>

{#if $dragData}
  <NavEntryPortal position={$position} dragData={$dragData} />
{/if}

<ForceDeleteConfirmation bind:open={showForceDelete} onDelete={onForceDelete} />
