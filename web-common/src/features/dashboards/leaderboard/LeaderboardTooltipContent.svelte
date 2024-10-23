<!-- @component
Creates a set of DimensionLeaderboardEntry components. This component makes it easy
to stitch together  chunks of a list. For instance, we can have:
leaderboard values above the fold
divider
leaderboard values not visible but selected
divider
see more button
-->
<script lang="ts">
  import MetaKey from "@rilldata/web-common/components/tooltip/MetaKey.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { isClipboardApiSupported } from "@rilldata/actions";

  export let label: string | number;
  export let selected: boolean;
  export let excluded: boolean;
  // false = include, true = exclude
  export let filterExcludeMode: boolean;
  export let atLeastOneActive;
</script>

<TooltipContent>
  <TooltipTitle>
    <svelte:fragment slot="name">
      {label}
    </svelte:fragment>
  </TooltipTitle>

  <TooltipShortcutContainer>
    {#if atLeastOneActive}
      <div>
        {excluded ? "Include" : "Exclude"}
        this dimension value
      </div>
    {:else}
      <div class="text-ellipsis overflow-hidden whitespace-nowrap">
        Filter {filterExcludeMode ? "out" : "on"}
        this dimension value
      </div>
    {/if}
    <Shortcut>Click</Shortcut>
  </TooltipShortcutContainer>
  {#if isClipboardApiSupported()}
    <TooltipShortcutContainer>
      <div>
        <StackingWord key="shift">Copy</StackingWord>
        this dimension value to clipboard
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + Click
      </Shortcut>
    </TooltipShortcutContainer>
  {/if}
  {#if !selected && atLeastOneActive}
    <TooltipShortcutContainer>
      <div>Exclusively select this dimension value</div>
      <Shortcut>
        <MetaKey />
      </Shortcut>
    </TooltipShortcutContainer>
  {/if}
</TooltipContent>
