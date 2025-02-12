import { getDashboardFromAggregationRequest } from "@rilldata/web-admin/features/dashboards/query-mappers/getDashboardFromAggregationRequest";
import { getDashboardFromComparisonRequest } from "@rilldata/web-admin/features/dashboards/query-mappers/getDashboardFromComparisonRequest";
import type {
  QueryMapperArgs,
  QueryRequests,
} from "@rilldata/web-admin/features/dashboards/query-mappers/types";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import { getFullInitExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createQueryServiceMetricsViewTimeRange,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewComparisonRequest,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, readable } from "svelte/store";

type DashboardStateForQuery = {
  exploreState?: MetricsExplorerEntity;
  exploreName?: string;
};

/**
 * Builds the dashboard url from query name and args.
 * Used to show the relevant dashboard for a report/alert.
 */
export function mapQueryToDashboard(
  exploreName: string,
  queryName: string | undefined,
  queryArgsJson: string | undefined,
  executionTime: string | undefined,
  annotations: Record<string, string>,
): CompoundQueryResult<DashboardStateForQuery> {
  if (!queryName || !queryArgsJson || !executionTime)
    return readable({
      isFetching: false,
      error: "Required parameters are missing.",
    });

  let metricsViewName: string = "";
  const req: QueryRequests = convertRequestKeysToCamelCase(
    JSON.parse(queryArgsJson),
  );
  let getDashboardState: (
    args: QueryMapperArgs<QueryRequests>,
  ) => Promise<MetricsExplorerEntity>;

  // get metrics view name and the query mapper function based on the query name.
  switch (queryName) {
    case "MetricsViewAggregation":
      metricsViewName =
        (req as V1MetricsViewAggregationRequest).metricsView ?? "";
      getDashboardState = getDashboardFromAggregationRequest;
      break;

    case "MetricsViewComparison":
      metricsViewName =
        (req as V1MetricsViewComparisonRequest).metricsViewName ?? "";
      getDashboardState = getDashboardFromComparisonRequest;
      break;

    // TODO
    // case "MetricsViewToplist":
    // case "MetricsViewRows":
    // case "MetricsViewTimeSeries":
  }

  if (!metricsViewName) {
    // error state
    return readable({
      isFetching: false,
      error:
        "Failed to find metrics view name. Please check the format of the report.",
    });
  }
  // backwards compatibility for older alerts created on metrics explore directly
  if (!exploreName) exploreName = metricsViewName;

  const instanceId = get(runtime).instanceId;

  return derived(
    [
      useExploreValidSpec(instanceId, exploreName),
      // TODO: handle non-timestamp dashboards
      createQueryServiceMetricsViewTimeRange(
        get(runtime).instanceId,
        metricsViewName,
        {},
      ),
    ],
    ([validSpecResp, timeRangeSummary], set) => {
      if (
        !validSpecResp.data?.metricsView ||
        !validSpecResp.data?.explore ||
        !timeRangeSummary.data
      ) {
        set({
          isFetching: true,
          error: "",
        });
        return;
      }

      if (validSpecResp.error || timeRangeSummary.error) {
        // error state
        set({
          isFetching: false,
          error:
            validSpecResp.error?.message ?? timeRangeSummary.error?.message,
        });
        return;
      }

      const { metricsView, explore } = validSpecResp.data;

      initLocalUserPreferenceStore(metricsViewName);
      const defaultExplorePreset = getDefaultExplorePreset(
        validSpecResp.data.explore,
        timeRangeSummary.data,
      );
      const { partialExploreState } = convertPresetToExploreState(
        validSpecResp.data.metricsView,
        validSpecResp.data.explore,
        defaultExplorePreset,
        [], // TODO
      );
      const defaultExploreState = getFullInitExploreState(
        metricsViewName,
        partialExploreState,
      );
      getDashboardState({
        queryClient,
        instanceId,
        dashboard: defaultExploreState,
        req,
        metricsView,
        explore,
        timeRangeSummary: timeRangeSummary.data.timeRangeSummary,
        executionTime,
        annotations,
      })
        .then((newExploreState) => {
          set({
            isFetching: false,
            error: "",
            data: {
              exploreState: newExploreState,
              exploreName,
            },
          });
        })
        .catch((err) => {
          set({
            isFetching: false,
            error: err.message,
          });
        });
    },
  );
}

/**
 * This method corrects the underscore naming to camel case.
 * This is the drawback of storing the request object as is.
 */
function convertRequestKeysToCamelCase(
  req: Record<string, any>,
): Record<string, any> {
  const newReq: Record<string, any> = {};

  for (const key in req) {
    const newKey = key.replace(/_(\w)/g, (_, c: string) => c.toUpperCase());
    const val = req[key];
    if (val && typeof val === "object" && !("length" in val)) {
      newReq[newKey] = convertRequestKeysToCamelCase(val);
    } else {
      newReq[newKey] = val;
    }
  }

  return newReq;
}
