import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import {
  AlertFormValues,
  getAlertQueryArgsFromFormValues,
} from "@rilldata/web-common/features/alerts/form-utils";
import { getComparisonProperties } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
  ComparisonPercentOfTotal,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
import {
  createQueryServiceMetricsViewAggregation,
  queryServiceMetricsViewAggregation,
  StructTypeField,
  TypeCode,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/query-core";
import type {
  CreateQueryOptions,
  CreateQueryResult,
} from "@tanstack/svelte-query";
import { derived, get } from "svelte/store";

export type AlertPreviewResponse = {
  rows: V1MetricsViewAggregationResponseDataItem[];
  schema: VirtualizedTableColumns[];
};

export function getAlertPreviewData(
  queryClient: QueryClient,
  formValues: AlertFormValues,
): CreateQueryResult<AlertPreviewResponse> {
  return derived(
    [useMetricsView(get(runtime).instanceId, formValues.metricsViewName)],
    ([metricsViewResp], set) => {
      return createQueryServiceMetricsViewAggregation(
        get(runtime).instanceId,
        formValues.metricsViewName,
        getAlertPreviewQueryRequest(formValues),
        {
          query: getAlertPreviewQueryOptions(
            queryClient,
            formValues,
            metricsViewResp.data,
          ),
        },
      ).subscribe(set);
    },
  );
}

function getAlertPreviewQueryRequest(
  formValues: AlertFormValues,
): V1MetricsViewAggregationRequest {
  const req = getAlertQueryArgsFromFormValues(formValues);
  req.limit = "50"; // arbitrary limit to make sure we do not pull too much of data
  if (req.timeRange) {
    req.timeRange.end = formValues.timeRange.end;
  }
  if (req.comparisonTimeRange) {
    req.comparisonTimeRange.end = formValues.timeRange?.end;
  }
  return req;
}

function getAlertPreviewQueryOptions(
  queryClient: QueryClient,
  formValues: AlertFormValues,
  metricsViewSpec: V1MetricsViewSpec | undefined,
): CreateQueryOptions<
  Awaited<ReturnType<typeof queryServiceMetricsViewAggregation>>,
  unknown,
  AlertPreviewResponse
> {
  return {
    enabled:
      !!formValues.measure &&
      !!metricsViewSpec &&
      (!formValues.timeRange || !!formValues.timeRange.end) &&
      (!formValues.comparisonTimeRange || !!formValues.comparisonTimeRange.end),
    select: (resp) => {
      return {
        rows: resp.data as V1MetricsViewAggregationResponseDataItem[],
        schema: (resp.schema?.fields
          ?.map((field) =>
            getSchemaEntryForField(metricsViewSpec as V1MetricsViewSpec, field),
          )
          .filter(Boolean) ?? []) as VirtualizedTableColumns[],
      };
    },
    queryClient,
  };
}

function getSchemaEntryForField(
  metricsViewSpec: V1MetricsViewSpec,
  field: StructTypeField,
): VirtualizedTableColumns | undefined {
  if (metricsViewSpec.dimensions) {
    for (const dimension of metricsViewSpec.dimensions) {
      if (dimension.name === field.name) {
        return {
          name: field.name as string,
          type: field.type?.code ?? TypeCode.CODE_STRING,
          label: dimension.label ?? field.name,
          enableResize: false,
          enableSorting: false,
        };
      }
    }
  }

  if (metricsViewSpec.measures) {
    for (const measure of metricsViewSpec.measures) {
      if (measure.name + ComparisonDeltaPreviousSuffix === field.name)
        return undefined;

      let label: VirtualizedTableColumns["label"] = measure.label ?? field.name;
      let format = measure.formatPreset;
      let type: string = field.type?.code ?? TypeCode.CODE_STRING;
      if (
        measure.name + ComparisonDeltaAbsoluteSuffix === field.name ||
        measure.name + ComparisonDeltaRelativeSuffix === field.name ||
        measure.name + ComparisonPercentOfTotal === field.name
      ) {
        const comparisonProps = getComparisonProperties(field.name, measure);
        label = comparisonProps.component;
        format = comparisonProps.format;
        type = comparisonProps.type;
      } else if (measure.name !== field.name) {
        continue;
      }

      return {
        name: field.name as string,
        type,
        label,
        format,
        enableResize: false,
        enableSorting: false,
      };
    }
  }

  return {
    name: field.name as string,
    type: field.type?.code ?? TypeCode.CODE_STRING,
    label: field.name,
    enableResize: false,
    enableSorting: false,
  };
}
