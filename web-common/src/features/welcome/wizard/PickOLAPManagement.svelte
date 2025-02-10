<script lang="ts">
  import type { OnboardingState } from "@rilldata/web-common/features/welcome/wizard/onboarding-state";
  import "./wizard.css";

  export let onboardingState: OnboardingState;

  const { managementType } = onboardingState;

  const managementTypeOptions = ["rill-managed", "self-managed"] as const;
</script>

<div class="management-toggle">
  {#each managementTypeOptions as option (option)}
    <button
      class="option"
      class:selected={$managementType === option}
      on:click={() => onboardingState.selectManagementType(option)}
    >
      {option.charAt(0).toUpperCase() + option.slice(1)}
    </button>
  {/each}
</div>

<style lang="postcss">
  .management-toggle {
    @apply flex flex-row gap-x-4 justify-center;
  }

  button {
    @apply w-[196px] h-[40px];
    @apply rounded;
    @apply text-base text-slate-800;
  }

  button:not(.selected) {
    @apply border border-slate-200;
  }
</style>
