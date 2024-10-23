<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipDescription from "@rilldata/web-common/components/tooltip/TooltipDescription.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import { isClipboardApiSupported } from "@rilldata/actions";

  export let measure: MetricsViewSpecMeasureV2;
  export let isMeasureExpanded = false;
  export let value = "";

  $: description =
    measure?.description || measure?.label || measure?.expression;
  $: name = measure?.label || measure?.expression;
</script>

<TooltipContent maxWidth="280px">
  <TooltipTitle>
    <svelte:fragment slot="name">
      {name}
    </svelte:fragment>
    <svelte:fragment slot="description">
      {value}
    </svelte:fragment>
  </TooltipTitle>
  <TooltipDescription>
    {description}
  </TooltipDescription>

  <TooltipShortcutContainer>
    {#if !isMeasureExpanded}
      <div>Expand measure</div>
      <Shortcut>Click</Shortcut>
    {/if}
    {#if isClipboardApiSupported()}
      <div>
        <StackingWord key="shift">Copy</StackingWord>
        number
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + Click
      </Shortcut>
    {/if}
  </TooltipShortcutContainer>
</TooltipContent>
