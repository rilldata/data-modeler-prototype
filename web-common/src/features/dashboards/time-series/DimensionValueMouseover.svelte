<script lang="ts">
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import MultiMetricMouseoverLabel from "@rilldata/web-common/components/data-graphic/marks/MultiMetricMouseoverLabel.svelte";
  import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
  import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
  import type { TimeSeriesDatum } from "./timeseries-data-store";
  import type { Point } from "@rilldata/web-common/components/data-graphic/marks/types";
  import type { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";

  export let point: TimeSeriesDatum;
  export let xAccessor: string;
  export let yAccessor: string;
  export let mouseoverFormat: ReturnType<typeof createMeasureValueFormatter>;
  export let dimensionData: DimensionDataItem[];
  export let dimensionValue: string | null | undefined;
  export let validPercTotal: number | null;
  export let hovered = false;
  export let colors: string[];

  $: x = point?.[xAccessor];

  function truncate(str: string) {
    if (!str?.length) return str;

    const truncateLength = 34;

    if (str.length > truncateLength) {
      // Check if last character is space
      if (str[truncateLength - 1] === " ") {
        return str.slice(0, truncateLength - 1) + "...";
      }
      return str.slice(0, truncateLength) + "...";
    }
    return str;
  }

  let pointsData = dimensionData;
  $: if (dimensionValue !== undefined) {
    const higlighted = dimensionData.filter((d) => d.value === dimensionValue);

    if (higlighted.length) {
      pointsData = higlighted;
    }
  }

  $: yValues = pointsData.map((dimension) => {
    if (!x) return { y: null, name: "" };

    const { entry: bisected } = bisectData(
      new Date(x),
      "center",
      xAccessor,
      dimension?.data,
    );
    if (bisected === undefined) return { y: null, name: "" };
    const y = bisected[yAccessor];
    return {
      y,
      name: dimension?.value,
    };
  });

  $: points = yValues
    .map((dimension, i) => {
      const currentPointIsNull = dimension.y === null;
      const y = Number(dimension.y);
      let value = mouseoverFormat(
        dimension.y === null || typeof dimension.y === "string"
          ? dimension.y
          : y,
      );

      if (validPercTotal) {
        const percOfTotal = y / validPercTotal;
        value =
          mouseoverFormat(y) + ",  " + (percOfTotal * 100).toFixed(2) + "%";
      }
      const point: Point = {
        x: Number(x),
        y,
        value: value ?? undefined,
        yOverride: currentPointIsNull,
        yOverrideLabel: "no current data",
        yOverrideStyleClass: `fill-gray-600 italic`,
        key: dimension.name === null ? "null" : dimension.name,
        label: hovered ? truncate(dimension.name || "null") : "",
        pointColorClass: "fill-" + colors[i],
        valueStyleClass: "font-bold",
        valueColorClass: "fill-gray-600",
        labelColorClass: "fill-gray-600",
        labelStyleClass: "font-semibold",
      };

      return point;
    })
    .filter((d) => !d.yOverride);

  /** get the final point set*/
  $: pointSet = points;
</script>

{#if pointSet.length}
  <WithGraphicContexts>
    <MultiMetricMouseoverLabel
      isDimension={true}
      attachPointToLabel
      direction="left"
      flipAtEdge="body"
      formatValue={mouseoverFormat}
      point={pointSet}
    />
  </WithGraphicContexts>
{/if}
