import {
  getQueryServiceMetricsViewAggregationQueryKey,
  getQueryServiceMetricsViewComparisonQueryKey,
  getQueryServiceMetricsViewTimeSeriesQueryKey,
  type V1Expression,
} from "@rilldata/web-common/runtime-client/index.js";

import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import type { QueryFunction } from "@tanstack/svelte-query";
import { error } from "@sveltejs/kit";
import {
  createInExpression,
  negateExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  V1TimeGrain,
  queryServiceMetricsViewTimeSeries,
} from "@rilldata/web-common/runtime-client";
import {
  queryServiceMetricsViewAggregation,
  queryServiceMetricsViewComparison,
} from "@rilldata/web-common/runtime-client";

import type {
  QueryServiceMetricsViewTimeSeriesBody,
  QueryServiceMetricsViewAggregationBody,
  QueryServiceMetricsViewComparisonBody,
  V1MetricsViewComparisonResponse,
} from "@rilldata/web-common/runtime-client";
import { prepareSortedQueryBody } from "@rilldata/web-common/features/dashboards/dashboard-utils";
import { SortType } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { DateTime } from "luxon";
import { getRange } from "@rilldata/web-common/lib/time/ranges/get-range";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types.js";
import { ISODurationToTimePreset } from "@rilldata/web-common/lib/time/ranges/index.js";

export const ssr = false;

export async function load({ parent, params, url }) {
  const parentData = await parent();
  const dashboardName = params.name;

  const metricsView = parentData.resources?.find(
    (resource) => resource.meta?.name?.name === dashboardName,
  );

  if (!metricsView) {
    throw error(404, "dashboard not found");
  }

  const spec = metricsView?.metricsView?.spec;
  const state = metricsView?.metricsView?.state;

  const instanceId = parentData.instance?.instanceId ?? "default";

  if (!spec || !state || !dashboardName) {
    throw error(404, "Metrics view not found");
  }

  const { min, max } = parentData.timeRange;

  const availableTimeZones = ["UTC"].concat(spec.availableTimeZones ?? []);

  const searchParams = new URLSearchParams(url.searchParams);

  const queryMeasure = searchParams.get("measure");

  const { queryStart, queryEnd, queryGrain, queryZone } =
    getTimeRangeParams(searchParams);

  const measures = spec.measures ?? [];
  const dimensions = spec.dimensions ?? [];
  const timeGrain =
    queryGrain ?? spec.smallestTimeGrain ?? V1TimeGrain.TIME_GRAIN_DAY;
  const measureNames = measures
    .map((m) => m.name)
    .filter((name): name is string => !!name);
  const sortMeasure =
    queryMeasure ?? spec.defaultMeasures[0] ?? measureNames[0];
  const defaultTimeRange = ISODurationToTimePreset(spec.defaultTimeRange);
  const timeZone = queryZone ?? availableTimeZones[0];
  console.log("HUH???", spec.defaultMeasures, measureNames, sortMeasure);
  const range = getRange({
    min,
    max,
    period: defaultTimeRange,
    zone: timeZone,
    queryStart,
    queryEnd,
  });

  const timeStart = range.start.toISO() ?? min;
  const timeEnd = range.end.toISO() ?? max;

  console.log({ timeStart, timeEnd });

  const where: V1Expression = {};

  dimensions.forEach(({ name }) => {
    if (!name) return;
    const valueStrings = searchParams.getAll(name);

    if (!valueStrings.length || !valueStrings) return;

    valueStrings.forEach((valueString) => {
      if (valueString) {
        const values = valueString.split(",") ?? [];
        if (values.length) {
          let expression = createInExpression(name, values, false);

          const exclude = values[0] === "!";

          if (exclude) {
            values.shift();
            expression = negateExpression(expression);
          }

          if (where?.cond?.exprs) {
            where.cond.exprs.push(expression);
          } else {
            where.cond = {
              op: "OPERATION_AND",
              exprs: [expression],
            };
          }
        }
      }
    });
  });

  let adjustedEnd: string | null = null;

  if (timeStart && timeEnd) {
    const luxonEnd = DateTime.fromISO(timeEnd);
    const luxonStart = DateTime.fromISO(timeStart);

    const numberOfPeriods = luxonEnd
      .diff(luxonStart, grainMap[timeGrain])
      .toObject()[grainMap[timeGrain]];

    adjustedEnd = luxonStart
      .plus({ [grainMap[timeGrain]]: Math.ceil(numberOfPeriods) + 1 })
      .setZone(timeZone)
      .toISO();
  }

  const body: QueryServiceMetricsViewTimeSeriesBody = {
    measureNames,
    timeZone: "UTC",
    timeGranularity: timeGrain,
    where,
    timeStart,
    timeEnd: adjustedEnd ?? undefined,
  };

  const totalsBody: QueryServiceMetricsViewAggregationBody = {
    measures,
    where,
    timeStart,
    timeEnd,
  };

  const timeSeriesQuery: QueryFunction<
    Awaited<ReturnType<typeof queryServiceMetricsViewTimeSeries>>
  > = ({ signal }) =>
    queryServiceMetricsViewTimeSeries(instanceId, dashboardName, body, signal);

  const totalsQuery: QueryFunction<
    Awaited<ReturnType<typeof queryServiceMetricsViewAggregation>>
  > = ({ signal }) =>
    queryServiceMetricsViewAggregation(
      instanceId,
      dashboardName,
      totalsBody,
      signal,
    );

  function createLeadboardQuery(dimensionName: string, dashboardName: string) {
    const localWhere = structuredClone(where);

    if (localWhere.cond?.exprs?.length) {
      localWhere.cond.exprs = localWhere.cond.exprs.filter((expr) => {
        return expr.cond?.exprs?.[0].ident !== dimensionName;
      });
    }

    const leaderBoardBody: QueryServiceMetricsViewComparisonBody =
      prepareSortedQueryBody(
        dimensionName,
        measureNames,
        {
          ready: true,
          showComparison: false,
          isFetching: false,
          timeStart,
          timeEnd,
        },
        sortMeasure,
        SortType.PERCENT,
        false,
        localWhere,
        undefined,
        50,
      );

    const leaderboardQuery: QueryFunction<
      Awaited<ReturnType<typeof queryServiceMetricsViewComparison>>
    > = ({ signal }) =>
      queryServiceMetricsViewComparison(
        instanceId,
        dashboardName,
        leaderBoardBody,
        signal,
      );

    return queryClient.fetchQuery({
      queryFn: leaderboardQuery,
      queryKey: getQueryServiceMetricsViewComparisonQueryKey(
        instanceId,
        dashboardName,
        leaderBoardBody,
      ),
    });
  }

  const totals = queryClient
    .fetchQuery({
      queryFn: totalsQuery,
      queryKey: getQueryServiceMetricsViewAggregationQueryKey(
        instanceId,
        dashboardName,
        totalsBody,
      ),
    })
    .catch(console.error);

  const timeSeries = queryClient
    .fetchQuery({
      queryFn: timeSeriesQuery,
      queryKey: getQueryServiceMetricsViewTimeSeriesQueryKey(
        instanceId,
        dashboardName,
        body,
      ),
    })
    .catch(console.error);

  const leaderBoards: Record<
    string,
    Promise<void | V1MetricsViewComparisonResponse>
  > = {};

  dimensions.forEach((dimension) => {
    if (!dimension.name) return;

    leaderBoards[dimension.name] = createLeadboardQuery(
      dimension.name,
      dashboardName,
    );
  });

  console.log("yeah");
  return {
    metricsView: spec,
    dimensions: state.validSpec?.dimensions ?? [],
    measures: state.validSpec?.measures ?? [],
    timeZone,
    timeStart,
    timeSeries: await timeSeries,
    timeEnd,
    totals: await totals,
    leaderBoards,
    timeGrain,
    sortMeasure,
    smallestTimeGrain: spec.smallestTimeGrain,
    availableTimeRanges: spec.availableTimeRanges,
  };
}

const grainMap = {
  TIME_GRAIN_HOUR: "hours",
  TIME_GRAIN_DAY: "days",
  TIME_GRAIN_WEEK: "weeks",
  TIME_GRAIN_MONTH: "months",
  TIME_GRAIN_YEAR: "years",
};

function getTimeRangeParams(searchParams: URLSearchParams) {
  const timeZone = (() => {
    const zone = searchParams.get("timeZone");
    searchParams.delete("timeZone");
    return zone ?? null;
  })();

  const timeStart = (() => {
    const start = searchParams.get("start");
    searchParams.delete("start");
    return start ?? null;
  })();

  const timeEnd = (() => {
    const end = searchParams.get("end");
    searchParams.delete("end");
    return end ?? null;
  })();

  const timeGrain = (() => {
    const grain = searchParams.get("timeGrain");
    searchParams.delete("timeGrain");

    if (isTimeGrain(grain)) return grain;
    return null;
  })();

  return {
    queryStart: timeStart,
    queryEnd: timeEnd,
    queryGrain: timeGrain,
    queryZone: timeZone,
  };
}

function isTimeGrain(grain: string | null): grain is V1TimeGrain {
  if (!grain) return false;
  return grain in V1TimeGrain;
}
