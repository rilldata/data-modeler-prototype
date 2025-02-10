<script lang="ts">
  export let numberOfDots: number;
  export let activeDotIndex: number;
</script>

<div class="progress-dots">
  {#each Array.from({ length: numberOfDots }) as _, index (index)}
    <span class="dot" class:active={activeDotIndex === index} />
  {/each}
</div>

<style lang="postcss">
  .progress-dots {
    @apply mt-2;
    @apply flex justify-center gap-x-3;
  }

  .dot {
    @apply w-2 h-2 rounded-full bg-slate-200;
  }

  .dot.active {
    @apply bg-primary-400;
  }
</style>
