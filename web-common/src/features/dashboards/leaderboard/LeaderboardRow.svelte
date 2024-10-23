<script lang="ts">
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import ExternalLink from "@rilldata/web-common/components/icons/ExternalLink.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import { copyToClipboard } from "@rilldata/actions";
  import { modified } from "@rilldata/actions";
  import { clamp } from "@rilldata/utils";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { slide } from "svelte/transition";
  import { getStateManagers } from "../state-managers/state-managers";
  import type { LeaderboardItemData } from "./leaderboard-utils";
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import LeaderboardTooltipContent from "./LeaderboardTooltipContent.svelte";
  import LongBarZigZag from "./LongBarZigZag.svelte";

  export let itemData: LeaderboardItemData;
  export let dimensionName: string;
  export let isValidPercentOfTotal: boolean;
  export let isTimeComparisonActive: boolean;
  export let tableWidth: number;
  export let borderTop = false;
  export let borderBottom = false;
  export let uri: string | undefined;
  export let firstColumnWidth: number;
  export let columnWidth: number;
  export let gutterWidth: number;

  let hovered = false;

  $: ({
    dimensionValue: label,
    selectedIndex,
    pctOfTotal,
    value: measureValue,
    prevValue: comparisonValue,
  } = itemData);

  $: selected = selectedIndex >= 0;

  const {
    selectors: {
      activeMeasure: { isSummableMeasure },
      numberFormat: { activeMeasureFormatter },
      dimensionFilters: { atLeastOneSelection, isFilterExcludeMode },
      comparison: { isBeingCompared: isBeingComparedReadable },
    },
    actions: {
      dimensionsFilter: { toggleDimensionValueSelection },
    },
  } = getStateManagers();

  $: isBeingCompared = $isBeingComparedReadable(dimensionName);
  $: filterExcludeMode = $isFilterExcludeMode(dimensionName);
  $: atLeastOneActive = $atLeastOneSelection(dimensionName);

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  $: previousValueString =
    comparisonValue !== undefined && comparisonValue !== null
      ? $activeMeasureFormatter(comparisonValue)
      : undefined;

  $: formattedValue = measureValue
    ? $activeMeasureFormatter(measureValue)
    : null;

  $: negativeChange = itemData.deltaAbs !== null && itemData.deltaAbs < 0;

  function shiftClickHandler(label: string) {
    let truncatedLabel = label?.toString();
    if (truncatedLabel?.length > TOOLTIP_STRING_LIMIT) {
      truncatedLabel = `${truncatedLabel.slice(0, TOOLTIP_STRING_LIMIT)}...`;
    }
    copyToClipboard(
      label,
      `copied dimension value "${truncatedLabel}" to clipboard`,
    );
  }

  $: href = makeHref(uri);

  $: percentOfTotal = $isSummableMeasure && pctOfTotal ? pctOfTotal : 0;

  $: barLength = (tableWidth - gutterWidth) * percentOfTotal;

  $: barColor = excluded
    ? "rgb(243 244 246)"
    : selected || hovered
      ? "var(--color-primary-200)"
      : "var(--color-primary-100)";

  // Bar gradient has to be split up across cells because of a bug in Safari
  // This is not necessary in other browsers, but doing it this way ensures consistency

  $: secondCellBarLength = clamp(0, barLength - firstColumnWidth, columnWidth);
  $: thirdCellBarLength = clamp(
    0,
    barLength - (firstColumnWidth + columnWidth),
    columnWidth,
  );
  $: fourthCellBarLength = clamp(
    0,
    Math.max(barLength - (firstColumnWidth + columnWidth * 2), 0),
    columnWidth,
  );

  $: firstCellGradient = `linear-gradient(to right, ${barColor}
    ${barLength}px, transparent ${barLength}px)`;

  $: secondCellGradient = secondCellBarLength
    ? `linear-gradient(to right, ${barColor}
    ${secondCellBarLength}px, transparent ${secondCellBarLength}px)`
    : undefined;

  $: thirdCellGradient = thirdCellBarLength
    ? `linear-gradient(to right, ${barColor}
    ${thirdCellBarLength}px, transparent ${thirdCellBarLength}px)`
    : undefined;

  $: fourthCellGradient = fourthCellBarLength
    ? `linear-gradient(to right, ${barColor}
    ${fourthCellBarLength}px, transparent ${fourthCellBarLength}px)`
    : undefined;

  $: showZigZag = barLength > tableWidth;

  // uri template or "true" string literal or undefined
  function makeHref(uriTemplateOrBoolean: string | undefined) {
    if (!uriTemplateOrBoolean) {
      return undefined;
    }

    const uri =
      uriTemplateOrBoolean === "true"
        ? label
        : uriTemplateOrBoolean
            .replace(/\s/g, "")
            .replace(`{{${dimensionName}}}`, label);

    const hasProtocol = /^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(uri);

    if (!hasProtocol) {
      return "https://" + uri;
    } else {
      return uri;
    }
  }
</script>

<tr
  class:border-b={borderBottom}
  class:border-t={borderTop}
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
  on:click={modified({
    shift: () => shiftClickHandler(label),
    click: (e) =>
      toggleDimensionValueSelection(
        dimensionName,
        label,
        false,
        e.ctrlKey || e.metaKey,
      ),
  })}
>
  <td>
    <LeaderboardItemFilterIcon
      {excluded}
      {isBeingCompared}
      selectionIndex={itemData?.selectedIndex}
    />
  </td>
  <td
    style:background={firstCellGradient}
    class:ui-copy={!atLeastOneActive}
    class:ui-copy-disabled={excluded}
    class:ui-copy-strong={!excluded && selected}
    class="relative size-full flex flex-none justify-between items-center leaderboard-label"
  >
    <Tooltip location="left" distance={20}>
      <FormattedDataType value={label} truncate />

      {#if previousValueString && hovered}
        <span
          class="opacity-50 whitespace-nowrap font-normal"
          transition:slide={{ axis: "x", duration: 200 }}
        >
          {previousValueString} →
        </span>
      {/if}

      {#if hovered && href}
        <a
          target="_blank"
          rel="noopener noreferrer"
          {href}
          title={href}
          on:click|stopPropagation
        >
          <ExternalLink className="fill-primary-600" />
        </a>
      {/if}

      <LeaderboardTooltipContent
        slot="tooltip-content"
        {atLeastOneActive}
        {excluded}
        {filterExcludeMode}
        {label}
        {selected}
      />
    </Tooltip>
  </td>
  <td style:background={secondCellGradient}>
    <FormattedDataType type="INTEGER" value={formattedValue || measureValue} />
    {#if showZigZag && !isTimeComparisonActive && !isValidPercentOfTotal}
      <LongBarZigZag />
    {/if}
  </td>
  {#if isTimeComparisonActive}
    <td style:background={thirdCellGradient}>
      <FormattedDataType
        type="INTEGER"
        value={itemData.deltaAbs
          ? $activeMeasureFormatter(itemData.deltaAbs)
          : null}
        customStyle={negativeChange ? "text-red-500" : ""}
      />
    </td>
    <td style:background={fourthCellGradient}>
      <PercentageChange
        value={itemData.deltaRel
          ? formatMeasurePercentageDifference(itemData.deltaRel)
          : null}
      />
      {#if showZigZag}
        <LongBarZigZag />
      {/if}
    </td>
  {:else if isValidPercentOfTotal}
    <td style:background={thirdCellGradient}>
      <PercentageChange value={itemData.pctOfTotal} />
      {#if showZigZag}
        <LongBarZigZag />
      {/if}
    </td>
  {/if}
</tr>

<style lang="postcss">
  td {
    @apply text-right p-0;
    @apply px-2  relative;
    height: 22px;
  }

  tr {
    @apply cursor-pointer;
  }

  tr:hover {
    @apply bg-gray-100;
  }

  td:first-of-type {
    @apply p-0 bg-background;
  }

  a {
    @apply absolute right-0 z-50  h-[22px] w-[32px];
    @apply bg-white flex items-center justify-center shadow-md rounded-sm;
  }

  a:hover {
    @apply bg-primary-100;
  }
</style>
