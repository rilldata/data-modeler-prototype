<script lang="ts">
  import { eventBus } from "@rilldata/events";
  import type { BannerMessage } from "@rilldata/events";
  import { onMount } from "svelte";
  import Banner from "./Banner.svelte";

  let banner: BannerMessage | null = null;

  onMount(() => {
    const unsubscribe = eventBus.on("banner", (newBanner) => {
      banner = newBanner;
    });

    return unsubscribe;
  });
</script>

{#if banner}
  <Banner {banner} />
{/if}
