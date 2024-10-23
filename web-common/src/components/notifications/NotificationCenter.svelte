<script lang="ts">
  import Notification from "./Notification.svelte";
  import { eventBus } from "@rilldata/events";
  import type { NotificationMessage } from "@rilldata/events";
  import { onMount } from "svelte";

  let notification: NotificationMessage | null = null;

  onMount(() => {
    const unsubscribe = eventBus.on("notification", (newNotification) => {
      notification = newNotification;
    });

    return unsubscribe;
  });

  function clear() {
    notification = null;
  }
</script>

{#if notification}
  <Notification {notification} onClose={clear} />
{/if}
