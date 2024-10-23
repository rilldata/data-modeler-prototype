<script lang="ts">
  import type { BannerMessage } from "@rilldata/events";
  import AlertCircleIcon from "../icons/AlertCircleOutline.svelte";
  import CheckCircleOutline from "../icons/CheckCircleOutline.svelte";
  import LoadingCircleOutline from "../icons/LoadingCircleOutline.svelte";
  import MoonCircleOutline from "../icons/MoonCircleOutline.svelte";

  export let banner: BannerMessage;

  const IconMap = {
    alert: AlertCircleIcon,
    check: CheckCircleOutline,
    sleep: MoonCircleOutline,
    loading: LoadingCircleOutline,
  };

  function dummyClickHandler() {}
</script>

<header class="{banner.type} app-banner">
  <div class="banner-content">
    {#if banner.iconType in IconMap}
      <svelte:component this={IconMap[banner.iconType]} size="14px" />
    {/if}
    {#if banner.includesHtml}
      <p class="banner-message">{@html banner.message}</p>
    {:else}
      <p class="banner-message">{banner.message}</p>
    {/if}
    {#if banner.cta}
      {#if banner.cta.type === "link"}
        <a href={banner.cta.url} target={banner.cta.target} class="banner-cta">
          {banner.cta.text}
        </a>
      {:else if banner.cta.type === "button"}
        <button
          on:click={banner.cta.onClick ?? dummyClickHandler}
          class="banner-cta"
        >
          {banner.cta.text}
        </button>
      {/if}
    {/if}
  </div>
</header>

<style lang="postcss">
  .app-banner {
    @apply h-7 bg-secondary-100 border-b;
  }

  .banner-content {
    @apply px-2 py-1.5;
    @apply flex justify-center items-center gap-x-2;
  }

  .app-banner :global(svg) {
    @apply text-secondary-900;
  }

  .banner-message {
    @apply text-xs text-secondary-900 font-medium;
  }

  .banner-message :global(a) {
    @apply text-primary-600;
  }

  .banner-cta {
    @apply text-primary-600 cursor-pointer;
  }

  .banner-message :global(a:hover) {
    @apply text-primary-700;
  }

  .success.app-banner {
    @apply bg-primary-100;
  }
  .success .banner-message,
  .success :global(svg) {
    @apply text-primary-800;
  }

  .info.app-banner {
    @apply bg-slate-100;
  }
  .info .banner-message,
  .info :global(svg) {
    @apply text-slate-700;
  }

  .warning.app-banner {
    @apply bg-yellow-100;
  }
  .warning .banner-message,
  .warning :global(svg) {
    @apply text-yellow-700;
  }

  .error.app-banner {
    @apply bg-red-100;
  }
  .error .banner-message,
  .error :global(svg) {
    @apply text-red-700;
  }
</style>
