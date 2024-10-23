<script lang="ts">
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import { copyToClipboard } from "@rilldata/actions";

  export let command: string;
  export let className = "";

  let copied = false;

  function copyCommand() {
    copyToClipboard(command);
    copied = true;
    setTimeout(() => (copied = false), 2500);
  }
</script>

<div class="flex flex-row text-gray-800 my-1 {className}">
  <div
    class="p-0.5 border border-gray-200 rounded-bl-sm rounded-tl-sm bg-gray-50 command-text"
  >
    {command}
  </div>
  <div
    role="button"
    tabindex="0"
    class="p-1 border border-gray-200 rounded-br-sm rounded-tr-sm bg-gray-50"
    class:cursor-pointer={copied}
    on:click={copyCommand}
    on:keydown={copyCommand}
  >
    {#if copied}
      <Check size="14px" />
    {:else}
      <CopyIcon color="gray-400" size="14px" />
    {/if}
  </div>
</div>

<style>
  .command-text {
    font-family: "Source Code Variable", monospace;
    padding-left: 7px;
    padding-right: 7px;
    border-right: 0;
  }
</style>
