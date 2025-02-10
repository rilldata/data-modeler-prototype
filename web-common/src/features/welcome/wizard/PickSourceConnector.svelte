<script lang="ts">
  import { Button } from "../../../components/button";
  import {
    CLICKHOUSE_SOURCE_CONNECTORS,
    DUCKDB_SOURCE_CONNECTORS,
  } from "../../connectors/connector-availability";
  import { logoIconMapping } from "../../connectors/connector-icon-mapping";
  import type { OnboardingState } from "./onboarding-state";

  export let onboardingState: OnboardingState;
  export let continueHref: string;

  const { olapDriver, firstDataSource } = onboardingState;

  $: dataSources =
    $olapDriver === "duckdb"
      ? DUCKDB_SOURCE_CONNECTORS
      : CLICKHOUSE_SOURCE_CONNECTORS;

  let isLoading = false;

  async function onSkip() {
    isLoading = true;
    await onboardingState.skipFirstSource();
    isLoading = false;
  }
</script>

<div class="data-sources">
  <h2 class="text-subheading text-gray-500">
    Choose a first data source to add.
  </h2>
  <div class="source-grid">
    {#each dataSources as source (source)}
      {@const { component, width, height } = logoIconMapping[source]}
      <button
        aria-label={source}
        class="source-button"
        class:active={$firstDataSource === source}
        on:click={() => onboardingState.toggleFirstDataSource(source)}
      >
        <svelte:component this={component} {width} {height} />
      </button>
    {/each}
  </div>
</div>

{#if $firstDataSource}
  <Button wide type="primary" href={continueHref}>Continue</Button>
{:else}
  <Button wide type="secondary" on:click={onSkip} disabled={isLoading}>
    Or, start with a blank project
  </Button>
{/if}

<!-- <div class="help-text">
      Don't see what you're looking for? <a href="#">Request a new connector</a>
    </div> -->

<style lang="postcss">
  .data-sources {
    @apply pt-5;
    @apply flex flex-col gap-y-2;
  }

  .source-grid {
    @apply grid grid-cols-[repeat(5,160px)];
    @apply gap-2 justify-center;
    @apply my-2;
  }

  .source-button {
    @apply w-40 h-20;
    @apply rounded-lg;
    @apply flex flex-col gap-2 items-center justify-center;
    @apply cursor-pointer;
  }

  .source-button:not(.active) {
    @apply border border-slate-200;
  }

  .source-button:hover {
    @apply bg-slate-50;
  }

  .source-button.active {
    @apply outline outline-2 outline-primary-300 bg-slate-50;
  }
</style>
