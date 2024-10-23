<script lang="ts">
  import TimestampDetail from "@rilldata/web-common/components/data-graphic/compositions/timestamp-profile/TimestampDetail.svelte";
  import TimestampSpark from "@rilldata/web-common/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import WithParentClientRect from "@rilldata/web-common/components/data-graphic/functional-components/WithParentClientRect.svelte";
  import { copyToClipboard } from "@rilldata/actions";
  import { TIMESTAMP_TOKENS } from "@rilldata/web-common/lib/duckdb-data-types";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import {
    getNullPercentage,
    getTimeSeriesAndSpark,
    isFetching,
  } from "../queries";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";

  export let connector: string;
  export let database: string;
  export let databaseSchema: string;
  export let objectName: string;
  export let columnName: string;
  export let type: string;
  export let mode = "summaries";
  export let example: any;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;

  export let enableProfiling = true;

  let timestampDetailHeight = 160;

  let active = false;

  /** queries used to power the different plots */
  $: nullPercentage = getNullPercentage(
    $runtime?.instanceId,
    connector,
    database,
    databaseSchema,
    objectName,
    columnName,
    enableProfiling,
  );

  $: timeSeries = getTimeSeriesAndSpark(
    $runtime?.instanceId,
    connector,
    database,
    databaseSchema,
    objectName,
    columnName,
    enableProfiling,
    active,
  );

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  $: fetchingSummaries = isFetching($timeSeries, $nullPercentage);
</script>

<ProfileContainer
  {active}
  {compact}
  emphasize={active}
  {example}
  {hideNullPercentage}
  {hideRight}
  isFetching={fetchingSummaries}
  {mode}
  on:select={toggleColumnProfile}
  onShiftClick={() => copyToClipboard(columnName)}
  {type}
>
  <ColumnProfileIcon isFetching={fetchingSummaries} slot="icon" {type} />
  <div slot="left">{columnName}</div>

  <!-- wrap in div to get size of grid item -->
  <div class={TIMESTAMP_TOKENS.textClass} slot="summary">
    <WithParentClientRect let:rect>
      <TimestampSpark
        bottom={4}
        color={"currentColor"}
        data={$timeSeries?.spark}
        height={18}
        top={4}
        width={rect?.width || 400}
        xAccessor="ts"
        yAccessor="count"
      />
    </WithParentClientRect>
  </div>
  <NullPercentageSpark
    nullCount={$nullPercentage?.nullCount}
    slot="nullity"
    totalRows={$nullPercentage?.totalRows}
    {type}
  />

  <div slot="details">
    <div class="pl-8 py-4" style:height="{timestampDetailHeight + 64 + 28}px">
      <WithParentClientRect let:rect>
        {#if $timeSeries?.data?.length && $timeSeries?.estimatedRollupInterval?.interval && $timeSeries?.smallestTimegrain}
          <TimestampDetail
            width={rect?.width - 56 || 400}
            mouseover={true}
            height={timestampDetailHeight}
            data={$timeSeries?.data}
            spark={$timeSeries?.spark}
            rollupTimeGrain={$timeSeries?.estimatedRollupInterval?.interval}
            estimatedSmallestTimeGrain={$timeSeries?.smallestTimegrain}
            xAccessor="ts"
            yAccessor="count"
          />
        {/if}
      </WithParentClientRect>
    </div>
  </div>
</ProfileContainer>
