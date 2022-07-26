import type {
  ApplicationEntity,
  ApplicationStateActionArg,
  ApplicationStateService,
} from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type {
  DerivedModelEntity,
  DerivedModelEntityService,
  DerivedModelStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import type {
  DerivedTableEntity,
  DerivedTableEntityService,
  DerivedTableStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type {
  DimensionDefinitionEntity,
  DimensionDefinitionStateActionArg,
  DimensionDefinitionStateService,
} from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  MeasureDefinitionEntity,
  MeasureDefinitionStateActionArg,
  MeasureDefinitionStateService,
} from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type {
  MetricsDefinitionEntity,
  MetricsDefinitionStateActionArg,
  MetricsDefinitionStateService,
} from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type {
  PersistentModelEntity,
  PersistentModelEntityService,
  PersistentModelStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {
  PersistentTableEntity,
  PersistentTableEntityService,
  PersistentTableStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";

export type EntityStateServicesMapType = {
  [EntityType.Table]?: {
    [StateType.Persistent]?: PersistentTableEntityService;
    [StateType.Derived]?: DerivedTableEntityService;
  };
  [EntityType.Model]?: {
    [StateType.Persistent]?: PersistentModelEntityService;
    [StateType.Derived]?: DerivedModelEntityService;
  };
  [EntityType.Application]?: {
    [StateType.Persistent]?: never;
    [StateType.Derived]?: ApplicationStateService;
  };
  [EntityType.MetricsDefinition]?: {
    [StateType.Persistent]?: MetricsDefinitionStateService;
    [StateType.Derived]?: never;
  };
  [EntityType.MeasureDefinition]?: {
    [StateType.Persistent]?: MeasureDefinitionStateService;
    [StateType.Derived]?: never;
  };
  [EntityType.DimensionDefinition]?: {
    [StateType.Persistent]?: DimensionDefinitionStateService;
    [StateType.Derived]?: never;
  };
  [EntityType.MetricsExplorer]?: never;
};

export type EntityRecordMapType = {
  [EntityType.Table]: {
    [StateType.Persistent]: PersistentTableEntity;
    [StateType.Derived]: DerivedTableEntity;
  };
  [EntityType.Model]: {
    [StateType.Persistent]: PersistentModelEntity;
    [StateType.Derived]: DerivedModelEntity;
  };
  [EntityType.Application]: {
    [StateType.Persistent]: never;
    [StateType.Derived]: ApplicationEntity;
  };
  [EntityType.MetricsDefinition]: {
    [StateType.Persistent]: MetricsDefinitionEntity;
    [StateType.Derived]: never;
  };
  [EntityType.MeasureDefinition]: {
    [StateType.Persistent]: MeasureDefinitionEntity;
    [StateType.Derived]: never;
  };
  [EntityType.DimensionDefinition]: {
    [StateType.Persistent]: DimensionDefinitionEntity;
    [StateType.Derived]: never;
  };
  [EntityType.MetricsExplorer]: never;
};
export type EntityStateActionArgMapType = {
  [EntityType.Table]: {
    [StateType.Persistent]: PersistentTableStateActionArg;
    [StateType.Derived]: DerivedTableStateActionArg;
  };
  [EntityType.Model]: {
    [StateType.Persistent]: PersistentModelStateActionArg;
    [StateType.Derived]: DerivedModelStateActionArg;
  };
  [EntityType.Application]: {
    [StateType.Persistent]: never;
    [StateType.Derived]: ApplicationStateActionArg;
  };
  [EntityType.MetricsDefinition]: {
    [StateType.Persistent]: MetricsDefinitionStateActionArg;
    [StateType.Derived]: never;
  };
  [EntityType.MeasureDefinition]: {
    [StateType.Persistent]: MeasureDefinitionStateActionArg;
    [StateType.Derived]: never;
  };
  [EntityType.DimensionDefinition]: {
    [StateType.Persistent]: DimensionDefinitionStateActionArg;
    [StateType.Derived]: never;
  };
  [EntityType.MetricsExplorer]: never;
};
