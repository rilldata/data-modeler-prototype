<script lang="ts">
  import PickOlapConnector from "@rilldata/web-common/features/welcome/wizard/PickOLAPConnector.svelte";
  import PickOlapManagement from "@rilldata/web-common/features/welcome/wizard/PickOLAPManagement.svelte";
  import PickSourceConnector from "@rilldata/web-common/features/welcome/wizard/PickSourceConnector.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  const { onboardingState } = data;
  const { managementType } = onboardingState;
</script>

<div class="wrapper">
  <div class="flex flex-col gap-y-2">
    <h1 class="text-gray-800">Let's set up your project</h1>
    <h2 class="text-gray-500">
      Choose an OLAP database for modeling data and serving dashboards.
      <a
        href="https://docs.rilldata.com/concepts/OLAP"
        target="_blank"
        rel="noopener noreferrer">Learn more</a
      >
    </h2>
  </div>
  <div class="flex flex-col gap-y-4">
    <PickOlapManagement {onboardingState} />
    <PickOlapConnector
      {onboardingState}
      continueHref={`/welcome/add-credentials`}
    />
  </div>
  {#if $managementType === "rill-managed"}
    <PickSourceConnector
      {onboardingState}
      continueHref={`/welcome/add-credentials`}
    />
  {/if}
</div>

<style lang="postcss">
  h1 {
    @apply text-lead text-slate-800;
  }

  h2 {
    @apply text-subheading;
  }

  .wrapper {
    @apply flex flex-col gap-y-4 items-center;
  }
</style>
