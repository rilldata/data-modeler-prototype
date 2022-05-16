import type {
  DerivedEntityRecord,
  EntityState,
  EntityStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  EntityStateService,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { guidGenerator } from "$lib/util/guid";

export interface ActiveEntity {
  type: EntityType;
  id: string;
}

export enum ApplicationStatus {
  Idle,
  Running,
}
export interface ApplicationEntity extends DerivedEntityRecord {}
export interface ApplicationState extends EntityState<ApplicationEntity> {
  activeEntity?: ActiveEntity;
  status: ApplicationStatus;
  projectId?: string;
  duckDbPath?: string;
}
export type ApplicationStateActionArg = EntityStateActionArg<
  ApplicationEntity,
  ApplicationState
>;

export class ApplicationStateService extends EntityStateService<
  ApplicationEntity,
  ApplicationState
> {
  public readonly entityType = EntityType.Application;
  public readonly stateType = StateType.Derived;

  public getEmptyState(): ApplicationState {
    return {
      lastUpdated: 0,
      entities: [],
      projectId: guidGenerator(),
      status: ApplicationStatus.Idle,
    };
  }

  public init(initialState: ApplicationState): void {
    if (!initialState.projectId) {
      initialState.projectId = guidGenerator();
    }
    super.init(initialState);
  }
}
