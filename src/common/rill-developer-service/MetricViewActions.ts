import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import type { TimeSeriesRollup } from "$common/database-service/DatabaseTimeSeriesActions";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { TimeSeriesValue } from "$lib/redux-store/timeseries/timeseries-slice";
import type { BigNumberResponse } from "$common/database-service/DatabaseMetricsExplorerActions";
import { getMapFromArray } from "$common/utils/getMapFromArray";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { getFallbackMeasureName } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

export interface MetricViewMetaResponse {
  name: string;
  timeDimension: {
    name: string;
    timeRange: TimeSeriesTimeRange;
  };
  dimensions: Array<DimensionDefinitionEntity>;
  measures: Array<MeasureDefinitionEntity>;
}

export interface MetricViewRequestTimeRange {
  start: string;
  end: string;
  granularity: string;
}
export interface MetricViewDimensionValue {
  name: string;
  values: Array<unknown>;
}
export type MetricViewDimensionValues = Array<MetricViewDimensionValue>;
export interface RuntimeRequestFilter {
  include: MetricViewDimensionValues;
  exclude: MetricViewDimensionValues;
}

export interface MetricViewTimeSeriesRequest {
  measures: Array<string>;
  time: MetricViewRequestTimeRange;
  filter?: RuntimeRequestFilter;
}
export interface MetricViewTimeSeriesResponse {
  meta: Array<{ name: string; type: string }>;
  // data: Array<{ time: string } & Record<string, number>>;
  data: Array<TimeSeriesValue>;
}

export interface MetricViewTopListRequest {
  measures: Array<string>;
  time: Pick<MetricViewRequestTimeRange, "start" | "end">;
  limit: number;
  offset: number;
  sort: Array<{ name: string; direction: "desc" | "asc" }>;
  filter?: RuntimeRequestFilter;
}
export interface MetricViewTopListResponse {
  meta: Array<{ name: string; type: string }>;
  // data: Array<Record<string, number | string>>;
  data: Array<{ label: string; value: number }>;
}

export interface MetricViewBigNumberRequest {
  measures: Array<string>;
  time: Pick<MetricViewRequestTimeRange, "start" | "end">;
  filter?: RuntimeRequestFilter;
}
export interface MetricViewBigNumberResponse {
  meta: Array<{ name: string; type: string }>;
  data: Record<string, number>;
}

function convertToActiveValues(filters: RuntimeRequestFilter): ActiveValues {
  if (!filters) return {};
  const activeValues: ActiveValues = {};
  filters.include.forEach((value) => {
    activeValues[value.name] = value.values.map((val) => [val, true]);
  });
  filters.exclude.forEach((value) => {
    activeValues[value.name] ??= [];
    activeValues[value.name].push(
      ...(value.values.map((val) => [val, false]) as Array<[unknown, boolean]>)
    );
  });
  return activeValues;
}

/**
 * Actions that get info for metrics explore.
 * Based on rill runtime specs.
 */
export class MetricViewActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async getMetricViewMeta(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    // TODO: validation
    const timeRange: TimeSeriesTimeRange = (
      await this.rillDeveloperService.dispatch(
        rillRequestContext,
        "getTimeRange",
        [metricsDefId]
      )
    ).data;
    const meta: MetricViewMetaResponse = {
      name: rillRequestContext.record.metricDefLabel,
      timeDimension: {
        name: rillRequestContext.record.timeDimension,
        timeRange,
      },
      measures: await this.getValidMeasures(rillRequestContext.record),
      dimensions: this.getValidDimensions(rillRequestContext.record),
    };
    return ActionResponseFactory.getRawResponse(meta);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getMetricViewTimeSeries(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    request: MetricViewTimeSeriesRequest
  ) {
    // TODO: validation
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const timeSeries: TimeSeriesRollup = await this.databaseActionQueue.enqueue(
      {
        id: metricsDefId,
        priority: DatabaseActionQueuePriority.ActiveModel,
      },
      "generateTimeSeries",
      [
        {
          tableName: model.tableName,
          timestampColumn: rillRequestContext.record.timeDimension,
          measures: request.measures.map((measureId) => ({
            ...this.dataModelerStateService
              .getMeasureDefinitionService()
              .getById(measureId),
          })),
          filters: convertToActiveValues(request.filter),
          timeRange: {
            ...request.time,
            interval: request.time.granularity,
          },
        },
      ]
    );
    const response: MetricViewTimeSeriesResponse = {
      meta: [], // TODO
      data: timeSeries.rollup.results,
    };
    return ActionResponseFactory.getRawResponse(response);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getMetricViewTopList(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    dimensionId: string,
    request: MetricViewTopListRequest
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const measure = this.dataModelerStateService
      .getMeasureDefinitionService()
      .getById(request.measures[0]);
    const dimension = this.dataModelerStateService
      .getDimensionDefinitionService()
      .getById(dimensionId);
    const data = await this.databaseActionQueue.enqueue(
      {
        id: rillRequestContext.id,
        priority: DatabaseActionQueuePriority.ActiveModel,
      },
      "getLeaderboardValues",
      [
        model.tableName,
        dimension.dimensionColumn,
        measure.expression,
        convertToActiveValues(request.filter),
        rillRequestContext.record.timeDimension,
        request.time,
      ]
    );
    const response: MetricViewTopListResponse = {
      meta: [], // TODO
      data,
    };
    return ActionResponseFactory.getRawResponse(response);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getRuntimeBigNumber(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    request: MetricViewBigNumberRequest
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const bigNumberResponse: BigNumberResponse =
      await this.databaseActionQueue.enqueue(
        {
          id: rillRequestContext.id,
          priority: DatabaseActionQueuePriority.ActiveModel,
        },
        "getBigNumber",
        [
          model.tableName,
          request.measures.map((measureId) =>
            this.dataModelerStateService
              .getMeasureDefinitionService()
              .getById(measureId)
          ),
          convertToActiveValues(request.filter),
          rillRequestContext.record.timeDimension,
          request.time,
        ]
      );
    const response: MetricViewBigNumberResponse = {
      meta: [], // TODO
      data: bigNumberResponse.bigNumbers,
    };
    return ActionResponseFactory.getRawResponse(response);
  }

  private getValidDimensions(metricsDef: MetricsDefinitionEntity) {
    const derivedModel = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Derived)
      .getById(metricsDef.sourceModelId);
    if (!derivedModel) {
      return [];
    }

    const columnMap = getMapFromArray(
      derivedModel.profile,
      (column) => column.name
    );

    return this.dataModelerStateService
      .getDimensionDefinitionService()
      .getManyByField("metricsDefId", metricsDef.id)
      .filter((dimension) => columnMap.has(dimension.dimensionColumn));
  }

  private async getValidMeasures(metricsDef: MetricsDefinitionEntity) {
    const measures = this.dataModelerStateService
      .getMeasureDefinitionService()
      .getManyByField("metricsDefId", metricsDef.id);
    return (
      await Promise.all(
        measures.map(async (measure, index) => {
          const measureValidation = await this.rillDeveloperService.dispatch(
            RillRequestContext.getNewContext(),
            "validateMeasureExpression",
            [metricsDef.id, measure.expression]
          );
          return {
            ...measure,
            ...(measureValidation.data as MeasureDefinitionEntity),
            sqlName: getFallbackMeasureName(index, measure.sqlName),
          };
        })
      )
    ).filter((measure) => measure.expressionIsValid === ValidationState.OK);
  }
}
