<script lang="ts">
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
</script>

<div class="flex size-full overflow-hidden">
  <Navigation />

  <section class="size-full overflow-hidden">
    <slot />
  </section>
</div>
