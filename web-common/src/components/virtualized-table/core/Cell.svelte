<script lang="ts">
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import { copyToClipboard, isClipboardApiSupported } from "@rilldata/actions";
  import { modified } from "@rilldata/actions";
  import { STRING_LIKES } from "@rilldata/web-common/lib/duckdb-data-types";
  import { formatDataTypeAsDuckDbQueryString } from "@rilldata/web-common/lib/formatters";
  import { createEventDispatcher, getContext } from "svelte";
  import BarAndLabel from "../../BarAndLabel.svelte";
  import type { VirtualizedTableConfig } from "../types";

  export let row;
  export let column;
  export let value;
  export let formattedValue: string | null = null;
  export let type;
  export let barValue = 0;
  export let rowActive = false;
  export let suppressTooltip = false;
  export let rowSelected = false;
  export let colSelected = false;
  export let atLeastOneSelected = false;
  export let excludeMode = false;
  export let positionStatic = false;
  export let label: string | undefined = undefined;

  const config: VirtualizedTableConfig = getContext("config");
  const isDimensionTable = config.table === "DimensionTable";

  let cellActive = false;
  $: isTextColumn = type === "VARCHAR" || type === "CODE_STRING";

  const dispatch = createEventDispatcher();

  function onFocus() {
    dispatch("inspect", row.index);
    cellActive = true;
  }

  function onSelectItem(e: MouseEvent) {
    dispatch("select-item", { index: row.index, meta: e.ctrlKey || e.metaKey });
  }

  function onBlur() {
    cellActive = false;
  }

  /** Because this table is virtualized,
   * it's a bit harder to get the proper
   * row-based hover highlighting. So let's
   * use javascript to solve this issue.
   */
  let activityStatus;
  $: {
    if (cellActive) {
      activityStatus = "bg-gray-200 dark:bg-gray-600";
    } else if (rowActive && !cellActive) {
      activityStatus = "bg-gray-100 dark:bg-gray-700";
    } else if (colSelected) {
      activityStatus = "bg-gray-50";
    } else {
      activityStatus = "surface";
    }
  }

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneSelected
    ? excludeMode
      ? rowSelected
      : !rowSelected
    : false;

  $: barColor = excluded
    ? "ui-measure-bar-excluded"
    : rowSelected
      ? "ui-measure-bar-included-selected"
      : "ui-measure-bar-included";

  $: tooltipValue =
    value && STRING_LIKES.has(type) && value.length >= TOOLTIP_STRING_LIMIT
      ? value?.slice(0, TOOLTIP_STRING_LIMIT) + "..."
      : value;

  $: formattedDataTypeStyle = excluded
    ? "font-normal ui-copy-disabled-faint"
    : rowSelected
      ? "font-normal ui-copy-strong"
      : "font-normal ui-copy";

  const shiftClick = async () => {
    let exportedValue = formatDataTypeAsDuckDbQueryString(value, type);
    copyToClipboard(exportedValue);
    // update this to set the active animation in the tooltip text
  };
</script>

<Tooltip
  distance={16}
  location="top"
  suppress={suppressTooltip || !isClipboardApiSupported()}
>
  <div
    class="
      {positionStatic ? 'static' : 'absolute'}
      z-9
      text-ellipsis
      whitespace-nowrap
      {isDimensionTable ? '' : 'border-r border-b'}
      {activityStatus}
      "
    on:blur={onBlur}
    on:click={onSelectItem}
    on:focus={onFocus}
    on:keydown
    on:mouseout={onBlur}
    on:mouseover={onFocus}
    role="gridcell"
    style:height="{row.size}px"
    style:left="{column.start}px"
    style:top="{row.start}px"
    style:width="{column.size}px"
    tabindex="0"
  >
    <BarAndLabel
      color={barColor}
      customBackgroundColor="rgba(0,0,0,0)"
      justify="left"
      showBackground={false}
      value={barValue}
    >
      <button
        aria-label={label}
        class="
          {isTextColumn ? 'text-left' : 'text-right'}
          {isDimensionTable ? '' : 'px-4'}
          w-full text-ellipsis overflow-x-hidden whitespace-nowrap
          "
        on:click={modified({
          shift: shiftClick,
        })}
        style:height="{row.size}px"
      >
        <FormattedDataType
          customStyle={formattedDataTypeStyle}
          inTable
          isNull={value === null || value === undefined}
          {type}
          value={formattedValue || value}
        />
      </button>
    </BarAndLabel>
  </div>
  <TooltipContent maxWidth="360px" slot="tooltip-content">
    <TooltipTitle>
      <FormattedDataType dark slot="name" value={tooltipValue} />
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>
        <StackingWord key="shift">Copy</StackingWord> this value to clipboard
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + Click
      </Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
