<script lang="ts">
  import { copyToClipboard } from "@rilldata/actions";
  import {
    DATA_TYPE_COLORS,
    INTERVALS,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import {
    getCountDistinct,
    getNullPercentage,
    getTopK,
    isFetching,
  } from "../queries";
  import TopK from "./details/TopK.svelte";
  import ColumnCardinalitySpark from "./sparks/ColumnCardinalitySpark.svelte";
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

  let topKLimit = 15;

  let active = false;

  $: nulls = getNullPercentage(
    $runtime?.instanceId,
    connector,
    database,
    databaseSchema,
    objectName,
    columnName,
    enableProfiling,
  );

  $: columnCardinality = getCountDistinct(
    $runtime?.instanceId,
    connector,
    database,
    databaseSchema,
    objectName,
    columnName,
    enableProfiling,
  );

  $: topK = getTopK(
    $runtime?.instanceId,
    connector,
    database,
    databaseSchema,
    objectName,
    columnName,
    enableProfiling,
  );

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  $: fetchingSummaries = isFetching($nulls);
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

  <svelte:fragment slot="left">{columnName}</svelte:fragment>

  <ColumnCardinalitySpark
    cardinality={$columnCardinality?.cardinality}
    {compact}
    slot="summary"
    totalRows={$columnCardinality?.totalRows}
    {type}
  />
  <NullPercentageSpark
    nullCount={$nulls?.nullCount}
    slot="nullity"
    totalRows={$nulls?.totalRows}
    {type}
  />
  <div
    class="pl-10 pr-4 py-4"
    class:hidden={INTERVALS.has(type)}
    slot="details"
  >
    <TopK
      colorClass={DATA_TYPE_COLORS["STRUCT"].bgClass}
      k={topKLimit}
      topK={$topK}
      totalRows={$columnCardinality?.totalRows}
      {type}
    />
  </div>
</ProfileContainer>
