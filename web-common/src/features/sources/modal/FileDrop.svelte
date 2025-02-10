<script lang="ts">
  import { goto } from "$app/navigation";
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    createRuntimeServiceUnpackEmpty,
    runtimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFileAPIPathFromNameAndType } from "../../entity-management/entity-mappers";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { isProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileLocalFileSourceYAML } from "../sourceUtils";
  import { uploadTableFiles } from "./file-upload";

  export let showDropOverlay: boolean;

  $: ({ instanceId } = $runtime);

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  const handleSourceDrop = async (e: DragEvent) => {
    showDropOverlay = false;

    const files = e?.dataTransfer?.files;

    // no-op if no files are dropped
    if (files === undefined) return;

    const uploadedFiles = uploadTableFiles(Array.from(files), instanceId);

    const initialized = await isProjectInitialized(instanceId);
    for await (const { tableName, filePath } of uploadedFiles) {
      try {
        // If project is uninitialized, initialize an empty project
        if (!initialized) {
          $unpackEmptyProject.mutate({
            instanceId,
            data: {
              displayName: EMPTY_PROJECT_TITLE,
            },
          });
        }

        const newFilePath = getFileAPIPathFromNameAndType(
          tableName,
          EntityType.Table,
        );
        await runtimeServicePutFile(instanceId, {
          path: newFilePath,
          blob: compileLocalFileSourceYAML(filePath),
          createOnly: false,
        });
        await goto(`/files/${newFilePath}`);
      } catch (err) {
        console.error(err);
      }
    }
  };
</script>

<Overlay bg="rgba(0,0,0,.6)">
  <div
    class="w-screen h-screen grid place-content-center"
    on:dragenter|preventDefault|stopPropagation
    on:dragleave|preventDefault|stopPropagation
    on:dragover|preventDefault|stopPropagation
    on:drag|preventDefault|stopPropagation
    on:drop|preventDefault|stopPropagation={handleSourceDrop}
    on:mouseup|preventDefault|stopPropagation={() => {
      showDropOverlay = false;
    }}
    role="presentation"
  >
    <div
      class="grid place-content-center grid-gap-2 text-white m-auto p-6 break-all text-3xl"
    >
      <span class="place-content-center">
        drop your files to add new source
      </span>
    </div>
  </div>
</Overlay>
