<script lang="ts">
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";

  export let error: string | undefined;
</script>

<div class="component-error">
  <div class="error-icon">
    <CrossIcon size="12px" />
  </div>
  <div class="error-message">
    {error ?? "Component Error"}
  </div>
</div>

<style lang="postcss">
  .component-error {
    @apply flex flex-col gap-y-3 w-full h-full items-center justify-center p-4;
    @apply text-red-600 bg-red-50/50;
  }

  .error-icon {
    @apply p-2 bg-red-100 rounded-full;
  }

  .error-message {
    @apply text-base font-medium text-center max-w-md;
  }
</style>
