<script lang="ts">
  import BooleanType from "@rilldata/web-common/components/icons/BooleanType.svelte";
  import FloatType from "@rilldata/web-common/components/icons/FloatType.svelte";
  import IntegerType from "@rilldata/web-common/components/icons/IntegerType.svelte";
  import StringlikeType from "@rilldata/web-common/components/icons/StringlikeType.svelte";
  import TimestampType from "@rilldata/web-common/components/icons/TimestampType.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { copyToClipboard, isClipboardApiSupported } from "@rilldata/actions";
  import {
    BOOLEANS,
    FLOATS,
    INTEGERS,
    INTERVALS,
    STRING_LIKES,
    TIMESTAMPS,
    isList,
    isNested,
    isStruct,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import ListType from "../icons/ListType.svelte";
  import StructType from "../icons/StructType.svelte";
  import ShiftKey from "../tooltip/ShiftKey.svelte";
  import Shortcut from "../tooltip/Shortcut.svelte";
  import StackingWord from "../tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../tooltip/TooltipTitle.svelte";
  import { modified } from "@rilldata/actions";

  export let color = "text-gray-400";
  export let type: string;
  export let suppressTooltip = false;

  function typeToSymbol(fieldType: string) {
    if (INTEGERS.has(fieldType)) {
      return IntegerType;
    } else if (FLOATS.has(fieldType)) {
      return FloatType;
    } else if (STRING_LIKES.has(fieldType)) {
      return StringlikeType;
    } else if (TIMESTAMPS.has(fieldType) || INTERVALS.has(type)) {
      return TimestampType;
    } else if (BOOLEANS.has(fieldType)) {
      return BooleanType;
    } else if (isStruct(fieldType)) {
      return StructType;
    } else if (isList(fieldType)) {
      return ListType;
    } else if (isNested(fieldType)) {
      return StructType;
    }
  }
</script>

<Tooltip
  location="left"
  distance={16}
  suppress={suppressTooltip || !isClipboardApiSupported()}
>
  <button
    title={type}
    class="
    {color}
    grid place-items-center rounded"
    style="width: 16px; height: 16px;"
    on:click={modified({
      shift: () => copyToClipboard(type),
    })}
  >
    <div>
      <svelte:component this={typeToSymbol(type)} size="16px" />
    </div>
  </button>
  <TooltipContent maxWidth="300px" slot="tooltip-content">
    <TooltipTitle>
      <div slot="name" class="truncate">
        {type}
      </div>
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>
        <StackingWord key="shift">Copy</StackingWord> type to clipboard
      </div>
      <Shortcut>
        <ShiftKey /> + Click
      </Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
