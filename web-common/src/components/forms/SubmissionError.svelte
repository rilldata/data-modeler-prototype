<script lang="ts">
  import AlertTriangle from "../icons/AlertTriangle.svelte";

  export let message: string;
</script>

<div>
  <AlertTriangle size="16px" className="shrink-0" />
  <p class="ml-2">
    {message}
  </p>
</div>

<style lang="postcss">
  div {
    @apply text-red-800 bg-red-100 border-red-300;
    @apply p-2 flex border-2 rounded-md;
  }
</style>
