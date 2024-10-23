<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import GenerateVegaSpecPrompt from "@rilldata/web-common/features/canvas-components/prompt/GenerateVegaSpecPrompt.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { extractFileName, splitFolderAndFileName } from "@rilldata/utils";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { WorkspaceHeader } from "@rilldata/web-common/layout/workspace";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let filePath: string;
  export let hasUnsavedChanges: boolean;

  let fileName: string;
  $: [, fileName] = splitFolderAndFileName(filePath);
  $: runtimeInstanceId = $runtime.instanceId;
  $: componentName = extractFileName(filePath);

  async function handleNameChange(newTitle: string) {
    const newRoute = await handleEntityRename(
      runtimeInstanceId,
      newTitle,
      filePath,
      componentName,
      fileArtifacts.getNamesForKind(ResourceKind.Component),
    );

    if (newRoute) await goto(newRoute);
  }

  let generateOpen = false;
</script>

<WorkspaceHeader
  resourceKind={ResourceKind.Component}
  onTitleChange={handleNameChange}
  titleInput={fileName}
  {hasUnsavedChanges}
  {filePath}
  showInspectorToggle={false}
>
  <svelte:fragment slot="cta">
    <PanelCTA side="right">
      <Button type="secondary" on:click={() => (generateOpen = true)}
        >Generate using AI</Button
      >
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>

<GenerateVegaSpecPrompt
  bind:open={generateOpen}
  chart={componentName}
  {filePath}
/>
