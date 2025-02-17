<script lang="ts">
  import { page } from "$app/stores";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import { isEmbedPage } from "../layout/navigation/navigation-utils";
  import AlertCircleOutline from "./icons/AlertCircleOutline.svelte";

  export let statusCode: number | null | undefined = undefined;
  export let header: string;
  export let body: string = "";
  export let detail: string | undefined = undefined;
  export let fatal = false;

  let showDetail = false;

  const onEmbedPage = isEmbedPage($page);
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if statusCode}
      <h1 class="status-code">{statusCode}</h1>
    {:else}
      <AlertCircleOutline size="64px" />
    {/if}
    <h2 class="header">{header}</h2>
    <CtaMessage>{body}</CtaMessage>
    {#if !fatal && !onEmbedPage}
      <CtaButton variant="secondary" href="/">Back to home</CtaButton>
    {/if}
    {#if detail}
      <section class="detail-section">
        <button
          class="detail-toggle"
          on:click={() => (showDetail = !showDetail)}
        >
          {#if !showDetail}
            Show details
          {:else}
            Hide details
          {/if}
        </button>
        {#if showDetail}
          <div class="detail-text-wrapper">
            <p class="font-mono">{detail}</p>
          </div>
        {/if}
      </section>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>

<style lang="postcss">
  .status-code {
    @apply text-8xl font-extrabold;
    @apply bg-gradient-to-b from-[#CBD5E1] to-[#E2E8F0] text-transparent bg-clip-text;
  }

  .header {
    @apply text-lg font-semibold;
  }

  .detail-section {
    @apply flex flex-col items-center gap-y-2;
  }

  .detail-toggle {
    @apply text-sm text-slate-600 font-medium;
    @apply flex items-center;
    @apply transition-colors duration-300 ease-in-out;
  }

  .detail-toggle:hover {
    @apply text-primary-700;
  }

  .detail-text-wrapper {
    @apply mt-4;
    @apply w-[700px] px-[26px] py-2;
    @apply border border-slate-200 rounded-sm;
    @apply bg-slate-50;
  }
</style>
