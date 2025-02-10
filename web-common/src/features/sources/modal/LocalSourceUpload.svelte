<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    openFileUploadDialog,
    uploadTableFiles,
  } from "@rilldata/web-common/features/sources/modal/file-upload";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import {
    createRuntimeServiceUnpackEmpty,
    runtimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
  import { isProjectInitialized } from "../../welcome/is-project-initialized";
  import { compileLocalFileSourceYAML } from "../sourceUtils";

  export let onSuccess: (newFilePath: string) => Promise<void>;

  $: ({ instanceId } = $runtime);

  const unpackEmptyProject = createRuntimeServiceUnpackEmpty();

  async function handleOpenFileDialog() {
    return handleUpload(await openFileUploadDialog());
  }

  async function handleUpload(files: Array<File>) {
    const uploadedFiles = uploadTableFiles(files, instanceId, false);
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

        await onSuccess(newFilePath);
      } catch (err) {
        console.error(err);
        overlay.set(null);
      }
    }
  }
</script>

<div class="local-source-upload">
  <Button on:click={handleOpenFileDialog} type="primary"
    >Upload a CSV, JSON or Parquet file
  </Button>
</div>
<slot name="actions" />

<style lang="postcss">
  .local-source-upload {
    @apply h-44 w-96 grid place-items-center mx-auto my-6;
    @apply border border-gray-300 rounded;
    @apply bg-gray-50;
  }
</style>
