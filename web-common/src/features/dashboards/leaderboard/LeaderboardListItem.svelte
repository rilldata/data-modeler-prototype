<script lang="ts">
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import { fly, slide } from "svelte/transition";
  import BarAndLabel from "../../../components/BarAndLabel.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";
  import LeaderboardTooltipContent from "./LeaderboardTooltipContent.svelte";
  import { getStateManagers } from "../state-managers/state-managers";
  import ContextColumnValue from "./ContextColumnValue.svelte";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import type { LeaderboardItemData } from "./leaderboard-utils";
  import { colorGetter } from "../filters/colorGetter";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";

  export let dimensionName: string;
  export let itemData: LeaderboardItemData;

  const {
    metricsViewName,
    selectors: {
      numberFormat: { activeMeasureFormatter },
      activeMeasure: { isSummableMeasure },
      dimensionFilters: { atLeastOneSelection, isFilterExcludeMode },
      comparison: { isBeingCompared: isBeingComparedReadable },
    },
    actions: {
      dimensionsFilter: { toggleDimensionValueSelection },
    },
  } = getStateManagers();

  $: label = itemData.dimensionValue;
  $: measureValue = itemData.value;
  $: selected = itemData.selectedIndex >= 0;
  $: comparisonValue = itemData.prevValue;
  $: pctOfTotal = itemData.pctOfTotal;

  $: isBeingCompared = $isBeingComparedReadable(dimensionName);
  $: filterExcludeMode = $isFilterExcludeMode(dimensionName);
  $: atLeastOneActive = $atLeastOneSelection(dimensionName);
  /** for summable measures, this is the value we use to calculate the bar % to fill */

  $: formattedValue = measureValue
    ? $activeMeasureFormatter(measureValue)
    : null;

  $: previousValueString =
    comparisonValue !== undefined && comparisonValue !== null
      ? $activeMeasureFormatter(comparisonValue)
      : undefined;
  $: showPreviousTimeValue = hovered && previousValueString !== undefined;
  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  $: renderedBarValue = $isSummableMeasure && pctOfTotal ? pctOfTotal : 0;

  $: color = excluded
    ? "ui-measure-bar-excluded"
    : selected
      ? "ui-measure-bar-included-selected"
      : "ui-measure-bar-included";

  const { shiftClickAction } = createShiftClickAction();
  async function shiftClickHandler(label) {
    await navigator.clipboard.writeText(label);
    let truncatedLabel = label?.toString();
    if (truncatedLabel?.length > TOOLTIP_STRING_LIMIT) {
      truncatedLabel = `${truncatedLabel.slice(0, TOOLTIP_STRING_LIMIT)}...`;
    }
    notifications.send({
      message: `copied dimension value "${truncatedLabel}" to clipboard`,
    });
  }

  let hovered = false;
  const onHover = () => {
    hovered = true;
  };
  const onLeave = () => {
    hovered = false;
  };
</script>

<Tooltip location="right">
  <button
    class="flex flex-row items-center w-full text-left transition-color"
    on:blur={onLeave}
    on:click={(e) => {
      if (e.shiftKey) return;
      toggleDimensionValueSelection(
        dimensionName,
        label,
        false,
        e.ctrlKey || e.metaKey,
      );
    }}
    on:focus={onHover}
    on:keydown
    on:mouseleave={onLeave}
    on:mouseover={onHover}
    on:shift-click={() => shiftClickHandler(label)}
    transition:slide={{ duration: 200 }}
    use:shiftClickAction
  >
    <div style:width="22px" class="grid place-items-center">
      {#if selected && !excluded && isBeingCompared}
        {@const color = colorGetter.get($metricsViewName, dimensionName, label)}
        <CheckCircle
          className="fill-{itemData.selectedIndex >= 7 ? 'gray-300' : color}"
          size="18px"
        />
      {:else if selected && !excluded}
        <Check size="20px" />
      {:else if selected && excluded}
        <Cancel size="20px" />
      {/if}
    </div>

    <BarAndLabel
      {color}
      justify={false}
      showBackground={false}
      showHover
      tweenParameters={{ duration: 200 }}
      value={renderedBarValue}
    >
      <div
        class="grid leaderboard-entry items-center gap-x-3"
        style:height="22px"
      >
        <!-- NOTE: empty class leaderboard-label is used to locate this elt in e2e tests -->
        <div
          class="leaderboard-label justify-self-start text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
          class:ui-copy={!atLeastOneActive}
          class:ui-copy-disabled={excluded}
          class:ui-copy-strong={!excluded && selected}
        >
          <FormattedDataType value={label} />
        </div>

        <div
          class="justify-self-end overflow-hidden ui-copy-number flex gap-x-4 items-baseline"
        >
          <!--
            FIXME: "local" default in svelte 4.0, remove after upgrading
            https://github.com/sveltejs/svelte/issues/6812#issuecomment-1593551644
          -->
          <div
            class="flex items-baseline gap-x-1"
            in:fly|local={{ duration: 200, y: 4 }}
          >
            {#if showPreviousTimeValue}
              <!--
              FIXME: "local" default in svelte 4.0, remove after upgrading
              https://github.com/sveltejs/svelte/issues/6812#issuecomment-1593551644
            -->
              <span
                class="inline-block opacity-50"
                transition:slideRight|local={{ duration: LIST_SLIDE_DURATION }}
              >
                {previousValueString}
                →
              </span>
            {/if}
            <FormattedDataType
              type="INTEGER"
              value={formattedValue || measureValue}
            />
          </div>
          <ContextColumnValue {itemData} />
        </div>
      </div>
    </BarAndLabel>
  </button>
  <!-- if the value is greater than 100%, we should add this little serration -->
  {#if renderedBarValue > 1.001}
    <LongBarZigZag />
  {/if}

  <LeaderboardTooltipContent
    {atLeastOneActive}
    {excluded}
    {filterExcludeMode}
    {label}
    {selected}
    slot="tooltip-content"
  />
</Tooltip>

<style>
  .leaderboard-entry {
    grid-template-columns: auto max-content;
  }
</style>
