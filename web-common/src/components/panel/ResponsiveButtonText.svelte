<!-- @component
This component will unmount the slotted text if the inspector width is less than
the threshold specified in collapseInspectorCTAButton, causing a cute
horizontal slide animation
-->
<script lang="ts">
  import { slideRight } from "@rilldata/web-common/lib/transitions";

  export let collapse = false;
</script>

{#if !collapse}
  <div
    transition:slideRight={{
      duration: 400,
      rightOffset: 8,
    }}
  >
    <slot />
  </div>
{/if}
