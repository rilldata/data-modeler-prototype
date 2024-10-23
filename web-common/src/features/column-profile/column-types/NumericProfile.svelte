<script lang="ts">
  import { copyToClipboard } from "@rilldata/actions";
  import {
    FLOATS,
    INTERVALS,
    isFloat,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    QueryServiceColumnNumericHistogramHistogramMethod,
    createQueryServiceColumnDescriptiveStatistics,
    createQueryServiceColumnRugHistogram,
  } from "@rilldata/web-common/runtime-client";
  import { getPriorityForColumn } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";
  import { derived } from "svelte/store";
  import { httpRequestQueue } from "../../../runtime-client/http-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ColumnProfileIcon from "../ColumnProfileIcon.svelte";
  import ProfileContainer from "../ProfileContainer.svelte";
  import {
    getNullPercentage,
    getNumericHistogram,
    getTopK,
    isFetching,
  } from "../queries";
  import { chooseBetweenDiagnosticAndStatistical } from "../utils";
  import NumericPlot from "./details/NumericPlot.svelte";
  import NullPercentageSpark from "./sparks/NullPercentageSpark.svelte";
  import NumericSpark from "./sparks/NumericSpark.svelte";

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

  $: diagnosticHistogram = getNumericHistogram(
    $runtime?.instanceId,
    connector,
    database,
    databaseSchema,
    objectName,
    columnName,
    QueryServiceColumnNumericHistogramHistogramMethod.HISTOGRAM_METHOD_DIAGNOSTIC,
    enableProfiling,
    active,
  );
  let fdHistogram;
  $: if (isFloat(type)) {
    fdHistogram = getNumericHistogram(
      $runtime?.instanceId,
      connector,
      database,
      databaseSchema,
      objectName,
      columnName,
      QueryServiceColumnNumericHistogramHistogramMethod.HISTOGRAM_METHOD_FD,
      enableProfiling,
      active,
    );
  }

  /**
   * We have two choices of histogram method: diagnostic and freedman-diaconis.
   * For integers, we go with diagnostic. For floating points, let's choose between
   * the most viable of diagnostic and freedman-diaconis. We'll remove
   * this once we've refactored floating-point columns toward a KDE plot.
   */
  $: histogramData = isFloat(type)
    ? chooseBetweenDiagnosticAndStatistical(
        $diagnosticHistogram?.data ?? [],
        $fdHistogram?.data,
      )
    : $diagnosticHistogram?.data;

  $: rug = createQueryServiceColumnRugHistogram(
    $runtime?.instanceId,
    objectName,
    {
      connector,
      database,
      databaseSchema,
      columnName,
      priority: getPriorityForColumn("rug-histogram", active),
    },
    {
      query: {
        select($query) {
          return $query?.numericSummary?.numericOutliers?.outliers;
        },
        enabled: enableProfiling,
      },
    },
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

  $: summary = derived(
    createQueryServiceColumnDescriptiveStatistics(
      $runtime?.instanceId,
      objectName,
      {
        connector,
        database,
        databaseSchema,
        columnName: columnName,
        priority: getPriorityForColumn("descriptive-statistics", active),
      },
      {
        query: {
          enabled: enableProfiling,
        },
      },
    ),
    ($query) => {
      return $query?.data?.numericSummary?.numericStatistics;
    },
  );

  function toggleColumnProfile() {
    active = !active;
    httpRequestQueue.prioritiseColumn(objectName, columnName, active);
  }

  $: fetchingSummaries = FLOATS.has(type)
    ? isFetching($nulls, $diagnosticHistogram, $fdHistogram)
    : isFetching($nulls, $diagnosticHistogram);

  /** if we have a singleton where all summary information is the same, let's construct a single bin. */
  $: if (
    $summary !== undefined &&
    $summary?.min !== undefined &&
    $summary?.min === $summary?.max &&
    $nulls?.totalRows !== undefined
  ) {
    const min = $summary.min;
    const boundaries = 10;
    histogramData = [
      // add 4 more empty bins
      ...Array.from({ length: boundaries }).map((_, i) => {
        return {
          bucket: -boundaries + i,
          count: 0,
          high: min - (boundaries - i - 1),
          low: min - (boundaries - i),
        };
      }),
      {
        bucket: boundaries,
        count: $nulls?.totalRows,
        low: min,
        high: min + 1,
      },
      // add more empty bins
      ...Array.from({ length: boundaries }).map((_, i) => {
        return {
          bucket: boundaries + i + 1,
          count: 0,
          low: min + i,
          high: min + i + 1,
        };
      }),
    ];
  }
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
  <NumericSpark {compact} data={histogramData ?? []} slot="summary" {type} />
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
    <NumericPlot
      data={histogramData ?? []}
      rug={$rug?.data ?? []}
      summary={$summary}
      topK={$topK ?? []}
      totalRows={$nulls?.totalRows}
      {type}
    />
  </div>
</ProfileContainer>
