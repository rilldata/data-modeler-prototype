<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import UndoIcon from "@rilldata/web-common/components/icons/UndoIcon.svelte";
  import type { Extension } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import { debounce } from "@rilldata/utils";
  import { FILE_SAVE_DEBOUNCE_TIME } from "./config";
  import { FileArtifact } from "../entity-management/file-artifact";
  import Codespace from "./Codespace.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import MetaKey from "@rilldata/web-common/components/tooltip/MetaKey.svelte";

  export let fileArtifact: FileArtifact;
  export let extensions: Extension[] = [];
  export let autoSave = true;
  export let editor: EditorView;
  export let forceLocalUpdates = false;
  export let forceDisableAutoSave = false;
  export let showSaveBar = true;
  export let refetchOnWindowFocus = true;
  export let onSave: (content: string) => void = () => {};
  export let onRevert: () => void = () => {};

  $: ({
    hasUnsavedChanges,
    saveLocalContent,
    revert,
    localContent,
    disableAutoSave,
  } = fileArtifact);

  $: debounceSave = debounce(save, FILE_SAVE_DEBOUNCE_TIME);

  async function handleKeydown(e: KeyboardEvent) {
    if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      await save();
    }
  }

  async function save() {
    const local = $localContent;
    if (local === null) return;
    onSave(local);
    await saveLocalContent();
  }

  function revertContent() {
    revert(); // Revert fileArtifact to remote content
    onRevert(); // Call revert callback
  }

  async function handleRefocus() {
    if (refetchOnWindowFocus) await fileArtifact.fetchContent(true);
  }
</script>

<svelte:window on:keydown={handleKeydown} on:focus={handleRefocus} />

<section>
  <div class="editor-container">
    {#key fileArtifact}
      <Codespace
        {extensions}
        {debounceSave}
        {forceLocalUpdates}
        {fileArtifact}
        autoSave={!forceDisableAutoSave && !disableAutoSave && autoSave}
        bind:editor
      />
    {/key}
  </div>

  {#if showSaveBar}
    <footer>
      <div class="flex gap-x-3">
        {#if !autoSave || disableAutoSave || forceDisableAutoSave}
          <Tooltip distance={8} activeDelay={300}>
            <Button
              type="subtle"
              disabled={!$hasUnsavedChanges}
              on:click={save}
            >
              <Check size="14px" />
              Save
            </Button>
            <TooltipContent slot="tooltip-content">
              <TooltipShortcutContainer pad={false}>
                Save
                <Shortcut>
                  <MetaKey action="S" />
                </Shortcut>
              </TooltipShortcutContainer>
            </TooltipContent>
          </Tooltip>

          <Button
            type="text"
            disabled={!$hasUnsavedChanges}
            on:click={revertContent}
          >
            <UndoIcon size="14px" />
            Revert changes
          </Button>
        {/if}
      </div>
      <div
        class="flex gap-x-1 items-center h-full bg-white rounded-full"
        class:hidden={disableAutoSave || forceDisableAutoSave}
      >
        <Switch
          bind:checked={autoSave}
          id="auto-save"
          small
          on:click={() => {
            if (!autoSave) debounceSave();
          }}
        />
        <Label class="font-normal text-xs" for="auto-save">Auto-save</Label>
      </div>
    </footer>
  {/if}
</section>

<style lang="postcss">
  .editor-container {
    @apply size-full overflow-auto p-2 pb-0 flex flex-col;
  }

  footer {
    @apply justify-between items-center flex flex-none;
    @apply h-10 p-2 w-full rounded-b-sm border-t bg-white;
  }

  section {
    @apply size-full flex-col bg-white flex overflow-hidden relative;
  }
</style>
