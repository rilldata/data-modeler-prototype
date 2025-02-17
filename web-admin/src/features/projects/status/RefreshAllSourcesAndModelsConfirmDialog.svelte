<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";

  export let open = false;
  export let onRefresh: () => void;

  function handleRefresh() {
    try {
      onRefresh();
      open = false;
    } catch (error) {
      console.error("Failed to refresh resource:", error);
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Refresh all sources and models?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          This will refresh all project sources and models.
          <br />
          <br />
          <span class="font-medium">Note:</span> To refresh a single resource, scroll
          to the source or model, click the '...' button, and select the refresh
          option.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="primary" on:click={handleRefresh}>Yes, refresh</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
