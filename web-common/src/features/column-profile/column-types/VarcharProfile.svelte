<script lang="ts">
  import { copyToClipboard } from "@rilldata/actions";
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
  export let example: any;
  export let type: string;

  export let hideRight = false;
  export let compact = false;
  export let hideNullPercentage = false;
  export let mode: "example" | "summaries" = "summaries";

  export let enableProfiling = true;

  let active = false;

  let topKLimit = 15;

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
    active,
  );

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  $: fetchingSummaries = isFetching($nulls, $columnCardinality);
</script>

<ProfileContainer
  {active}
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
  />
  <NullPercentageSpark
    nullCount={$nulls?.nullCount}
    slot="nullity"
    totalRows={$nulls?.totalRows}
    {type}
  />

  <div
    class="pl-10 pr-4 py-4"
    slot="details"
    style:min-height="{Math.min(
      topKLimit,
      $columnCardinality?.cardinality ?? Infinity,
    ) *
      18 +
      42 || 250}px"
  >
    <div>
      <TopK
        k={topKLimit}
        topK={$topK}
        totalRows={$columnCardinality?.totalRows}
        {type}
      />
    </div>
  </div>
</ProfileContainer>
