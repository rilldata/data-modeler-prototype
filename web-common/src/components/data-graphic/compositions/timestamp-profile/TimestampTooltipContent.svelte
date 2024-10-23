<script lang="ts">
  /**
   * The TimestampTooltipContent is used in the TimestampDetail component.
   * The goal is to provide user a quick & easy onboarding for the basic TimestampDetail
   * actions of zooming and panning. This component is a bit extra.
   */
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "@rilldata/web-common/lib/formatters";
  import { isClipboardApiSupported } from "@rilldata/actions";
  import TimestampSpark from "./TimestampSpark.svelte";

  export let xAccessor: string;
  export let yAccessor: string;
  export let data;
  // FIXME: document meaning of these special looking numbers
  // e.g. something like width = y* CHAR_HEIGHT, height = CHAR_HEIGHT?
  export let width = 84;
  export let height = 12;

  export let totalRows: number;
  export let zoomedRows: number;

  // these flags change the text in the tooltip.
  export let zoomed = false;
  export let zooming = false;
  // this determines the "shake" of the pan label when panning.
  export let tooltipPanShakeAmount = 0;
  // the window bounds for the spark within the zoom row of the tooltip.
  export let zoomWindowXMin: Date | undefined = undefined;
  export let zoomWindowXMax: Date | undefined = undefined;
</script>

<TooltipContent>
  <div class="pt-1 pb-1 italic font-semibold">
    {#if zoomed}
      <div
        class="grid space-between w-full"
        style="grid-template-columns: auto max-content;"
      >
        <div>
          {#if zooming}<span>Zoomed</span>{:else}<span>Zooming</span>{/if}
          to {formatInteger(zoomedRows)} row{#if zoomedRows !== 1}s{/if}
        </div>
        <div class="text-right text-gray-300 font-normal not-italic">
          {formatBigNumberPercentage(zoomedRows / totalRows)}
        </div>
      </div>
    {:else}
      Showing all {formatInteger(totalRows)} rows
    {/if}
  </div>
  <TooltipShortcutContainer>
    {#if isClipboardApiSupported()}
      <div>
        <StackingWord key="shift">Copy</StackingWord> to clipboard
      </div>
      <Shortcut>
        <span
          style="
          font-family: var(--system); 
          font-size: 11.5px;
        ">⇧</span
        > + Click
      </Shortcut>
    {/if}
    <div>
      <div style:transform="translateX({tooltipPanShakeAmount}px)">Pan</div>
    </div>
    <Shortcut>Click + Drag</Shortcut>
    <div>
      Zoom
      <div style:display="inline-grid">
        <TimestampSpark
          {data}
          {xAccessor}
          {yAccessor}
          {width}
          {height}
          left={0}
          right={0}
          top={0}
          bottom={0}
          color="hsla(217,1%,99%, .5)"
          zoomWindowColor="hsla(217, 70%, 60%, .6)"
          zoomWindowBoundaryColor="hsla(217, 10%, 90%, .9)"
          {zoomWindowXMin}
          {zoomWindowXMax}
        />
      </div>
    </div>
    <Shortcut>Ctrl + Click + Drag</Shortcut>
  </TooltipShortcutContainer>
</TooltipContent>
