<script lang="ts">
  import Notification from "./Notification.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { NotificationMessage } from "@rilldata/web-common/lib/event-bus/events";
  import { onMount } from "svelte";
  import { NOTIFICATION_TIMEOUT } from "./constants";

  let notifications: NotificationMessage[] = [];
  let currentTimeoutId: number | null = null;

  onMount(() => {
    const unsubscribeNotification = eventBus.on(
      "notification",
      (notification) => {
        // Clear existing notifications before showing new one
        notifications = [notification];

        // Clear any existing timeout
        if (currentTimeoutId) {
          clearTimeout(currentTimeoutId);
          currentTimeoutId = null;
        }

        // Set up auto-dismiss for non-persisted notifications
        if (
          !notification.options?.persisted &&
          notification.type !== "loading"
        ) {
          const timeout = notification.options?.timeout ?? NOTIFICATION_TIMEOUT;
          currentTimeoutId = window.setTimeout(clear, timeout);
        }
      },
    );

    const unsubscribeClear = eventBus.on("clear-all-notifications", () => {
      clear();
    });

    return () => {
      unsubscribeNotification();
      unsubscribeClear();
      if (currentTimeoutId) {
        clearTimeout(currentTimeoutId);
      }
    };
  });

  function clear() {
    notifications = [];
    if (currentTimeoutId) {
      clearTimeout(currentTimeoutId);
      currentTimeoutId = null;
    }
  }
</script>

{#each notifications as notification}
  <Notification {notification} onClose={clear} />
{/each}
