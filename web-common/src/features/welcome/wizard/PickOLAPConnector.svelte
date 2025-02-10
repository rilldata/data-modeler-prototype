<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { logoIconMapping } from "../../connectors/connector-icon-mapping";
  import type { OnboardingState } from "./onboarding-state";
  import "./wizard.css";

  export let onboardingState: OnboardingState;
  export let continueHref: string;

  const { managementType, olapDriver } = onboardingState;

  const RILL_MANAGED_OLAP_OPTIONS = [
    {
      name: "duckdb",
      copy: "Ideal for projects up to 10GB",
    },
    {
      name: "clickhouse",
      copy: "Great for projects up to 100GB",
    },
  ];

  const SELF_MANAGED_OLAP_OPTIONS = [
    {
      name: "clickhouse",
      copy: "Great for projects up to 100GB",
    },
    {
      name: "druid",
      copy: "Connect to an existing cluster",
    },
  ];

  $: olapOptions =
    $managementType === "rill-managed"
      ? RILL_MANAGED_OLAP_OPTIONS
      : SELF_MANAGED_OLAP_OPTIONS;
</script>

<!-- For now, Rill-managed OLAP will always be DuckDB, so we don't need to show the options -->
{#if $managementType === "self-managed"}
  <section class="flex flex-col gap-y-4 items-center">
    <div class="olap-cards">
      {#each olapOptions as option (option.name)}
        {@const { component, width, height } = logoIconMapping[option.name]}
        <button
          class="option"
          class:selected={$olapDriver === option.name}
          on:click={() => onboardingState.selectOLAP(option.name)}
        >
          <svelte:component
            this={component}
            {width}
            {height}
            className="shrink-0"
          />
          <small class="description">{option.copy}</small>
        </button>
      {/each}
    </div>

    <Button wide type="primary" disabled={!$olapDriver} href={continueHref}>
      Continue
    </Button>
  </section>
{/if}

<style lang="postcss">
  .olap-cards {
    @apply flex justify-center gap-x-4;
  }

  button {
    @apply w-[196px] h-[64px] px-2 py-3;
    @apply flex flex-col gap-y-1 items-center justify-center;
  }

  .description {
    @apply text-slate-500 text-xs;
  }
</style>
