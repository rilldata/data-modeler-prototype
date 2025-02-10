<script lang="ts">
  import LoadingCircleOutline from "../icons/LoadingCircleOutline.svelte";

  export let disabled = false;
  export let isLoading = false;
  export let imageUrl = "";
</script>

<button
  class:gradient={!imageUrl}
  on:click
  on:keydown={(e) => e.key === "Enter" && e.currentTarget.click()}
  aria-disabled={disabled}
  class:loading={isLoading}
  style:background-image={imageUrl ? `url('${imageUrl}')` : ""}
>
  {#if isLoading}
    <div
      class="absolute z-10 inset-0 flex items-center justify-center backdrop-blur-sm"
    >
      <LoadingCircleOutline size="48px" color="var(--color-primary-600)" />
    </div>
  {/if}
  <slot />
</button>

<style lang="postcss">
  button {
    @apply bg-no-repeat bg-center bg-cover;
    @apply relative select-none;
    @apply size-60 rounded-md;
    @apply flex flex-col items-center justify-center gap-y-2;
    @apply transition duration-300 ease-out;
    @apply cursor-pointer overflow-hidden;

    box-shadow:
      0px 2px 3px rgba(15, 23, 42, 0.06),
      0px 1px 3px rgba(15, 23, 42, 0.08),
      0px 0px 0px 1px rgba(15, 23, 42, 0.12);
  }

  .gradient {
    @apply bg-gradient-to-b from-white to-slate-50;
  }

  button[aria-disabled="true"] {
    cursor: not-allowed;
    pointer-events: none;
  }

  button[aria-disabled="true"]:not(.loading) {
    opacity: 0.4;
  }

  button:hover {
    box-shadow:
      0px 2px 3px rgba(99, 102, 241, 0.2),
      0px 1px 3px rgba(15, 23, 42, 0.08),
      0px 0px 0px 1px rgba(15, 23, 42, 0.12),
      0px 4px 6px rgba(15, 23, 42, 0.12);
  }
</style>
