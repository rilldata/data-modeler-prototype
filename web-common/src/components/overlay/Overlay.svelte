<script>
  import { fade } from "svelte/transition";
  import { portal } from "@rilldata/actions";

  export let bg = "rgba(0,0,0,.8)";

  // FIXME: when an element pops up before the last one has dismounted,
  // the focus trapping won't work.
  // We'll need a better solution than this!
  function captureKeydown(event) {
    // capture all events

    // FIXME: `.blur()` doesn't exist on activeElement --
    // What was the intent here @djbarnwal?
    // document.activeElement.blur();
    event.preventDefault();
  }

  let classes =
    "fixed top-0 left-0 right-0 w-screen h-screen grid place-content-center text-lg z-[120]";
</script>

<svelte:window on:keydown={captureKeydown} />

<div use:portal>
  {#key bg}
    <div
      transition:fade|global={{ duration: 200 }}
      style:background={bg}
      class={classes}
    />
  {/key}
  <div transition:fade|global={{ duration: 300 }} class={classes}>
    <slot />
  </div>
</div>

<style lang="postcss">
  :global(.body) {
    transition: transform 400ms;
    transform-origin: center;
  }
  :global(.big-process-overlay) {
    transform: scale(0.95) translateY(2.5vh);
  }
</style>
