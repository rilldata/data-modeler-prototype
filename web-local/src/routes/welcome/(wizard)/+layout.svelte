<script lang="ts">
  import ProgressDots from "@rilldata/web-common/features/welcome/wizard/ProgressDots.svelte";
  import type { LayoutData } from "./$types";

  export let data: LayoutData;

  const {
    onboardingState: { getNumberOfSteps, getCurrentStep },
  } = data;

  const numberOfSteps = getNumberOfSteps();
  const currentStep = getCurrentStep();
</script>

<main>
  <slot />
</main>
<footer>
  <ProgressDots
    numberOfDots={$numberOfSteps}
    activeDotIndex={$currentStep - 1}
  />
</footer>

<style lang="postcss">
  main {
    @apply w-full pt-[100px];
    @apply flex-grow flex flex-col;
    @apply justify-start items-center text-center overflow-auto;
  }

  footer {
    @apply h-14 shrink-0;
  }
</style>
