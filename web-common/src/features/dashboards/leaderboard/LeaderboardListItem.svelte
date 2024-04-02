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

  import ContextColumnValue from "./ContextColumnValue.svelte";
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import type { LeaderboardItemData } from "./leaderboard-utils";

  import { page } from "$app/stores";

  export let dimensionName: string;
  export let formatter = (value: number) => value.toString();
  export let itemData: LeaderboardItemData;

  $: label = itemData.dimensionValue;
  $: measureValue = formatter(itemData.value);

  $: comparisonValue = itemData.prevValue;
  $: pctOfTotal = itemData.pctOfTotal;

  // const {
  //   selectors: {
  //     numberFormat: { activeMeasureFormatter },
  //     activeMeasure: { isSummableMeasure },
  //     dimensionFilters: { atLeastOneSelection, isFilterExcludeMode },
  //     comparison: { isBeingCompared: isBeingComparedReadable },
  //   },
  //   actions: {
  //     dimensionsFilter: { toggleDimensionValueSelection },
  //   },
  // } = getStateManagers();

  $: isBeingCompared = $page.url.searchParams.get("compare") === dimensionName;
  $: filterExcludeMode =
    $page.url.searchParams.get(dimensionName)?.split(",")[0] === "!";
  $: atLeastOneActive =
    $page.url.searchParams.get(dimensionName)?.split(",")?.length > 0;
  /** for summable measures, this is the value we use to calculate the bar % to fill */

  $: formattedValue = measureValue ? measureValue : null;

  $: previousValueString =
    comparisonValue !== undefined && comparisonValue !== null
      ? comparisonValue.toString()
      : undefined;

  $: showPreviousTimeValue = hovered && previousValueString !== undefined;
  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;
  let isSummableMeasure = true;
  $: renderedBarValue = isSummableMeasure && pctOfTotal ? pctOfTotal : 0;

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

  import { goto } from "$app/navigation";

  $: includedValues = searchParams.get(dimensionName)?.split(",") ?? [];

  $: searchParams = $page.url.searchParams;
  $: selected = Boolean(includedValues.includes(itemData.dimensionValue));

  function createDimensionLink(dimensionName: string, dimensionValue: string) {
    const newParams = new URLSearchParams(searchParams);

    const existing = searchParams.get(dimensionName);
    const values = existing ? existing.split(",") : [];
    if (!values.includes(dimensionValue)) values.push(dimensionValue);
    else values.splice(values.indexOf(dimensionValue), 1);

    if (values.length === 0) newParams.delete(dimensionName);
    else newParams.set(dimensionName, values.join(","));

    return `?${newParams.toString()}`;
  }
</script>

<Tooltip location="right">
  <button
    class="flex flex-row items-center w-full text-left transition-color"
    on:blur={onLeave}
    on:click={(e) => {
      if (e.shiftKey) return;
      // toggleDimensionValueSelection(
      //   dimensionName,
      //   label,
      //   false,
      //   e.ctrlKey || e.metaKey,
      // );

      goto(createDimensionLink(dimensionName, itemData.dimensionValue));
    }}
    on:focus={onHover}
    on:keydown
    on:mouseleave={onLeave}
    on:mouseover={onHover}
    on:shift-click={() => shiftClickHandler(label)}
    transition:slide={{ duration: 200 }}
    use:shiftClickAction
  >
    <LeaderboardItemFilterIcon
      {excluded}
      {isBeingCompared}
      selectionIndex={includedValues.findIndex(
        (value) => value === itemData.dimensionValue,
      )}
    />
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
          <!-- <ContextColumnValue {itemData} /> -->
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
