<script lang="ts">
  import { page } from "$app/stores";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";

  let showOtherFiles = true;
</script>

<div class="h-fit flex flex-col">
  <NavigationHeader bind:show={showOtherFiles}>Others</NavigationHeader>

  {#if showOtherFiles}
    <ol>
      <li aria-label={"rill.yaml"}>
        <NavigationEntry
          showContextMenu={false}
          name={"rill.yaml"}
          open={$page.url.pathname === `/rill.yaml`}
        />
      </li>
    </ol>
  {/if}
</div>
