<script lang="ts">
  import { onMount } from "svelte";
  import { eventBus } from "@rilldata/events";

  export let key: "command" | "shift";

  let span: HTMLSpanElement;

  onMount(() => {
    const unsubscribe = eventBus.on(`${key}-click`, () => {
      if (!span) return;
      span.animate(
        [
          {
            transform: "translateY(2px) translateX(2px)",
            boxShadow:
              "-1px -1px 0px rgba(100, 100, 100, 1), -2px -2px 0px rgba(75, 75, 75, 1), -3px -3px 0px rgba(50, 50, 50, 1)",
          },
        ],
        { duration: 250, easing: "ease-in-out" },
      );
    });

    return unsubscribe;
  });
</script>

<span bind:this={span}><slot /></span>

<style lang="postcss">
  span {
    @apply inline-block;
    border-radius: 2px;
    position: relative;
    mix-blend-mode: screen;
    background-blend-mode: screen;
  }
</style>
