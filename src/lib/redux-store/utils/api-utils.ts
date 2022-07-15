import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import type {
  EntityRecord,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { ActionCreatorWithPreparedPayload } from "@reduxjs/toolkit";
import type { EntityRecordMapType } from "$common/data-modeler-state-service/entity-state-service/EntityStateServicesMap";
import { fetchWrapper } from "$lib/util/fetchWrapper";
import type { ValidationConfig } from "$lib/redux-store/utils/validation-utils";
import { validateEntity } from "$lib/redux-store/utils/validation-utils";
import type { RillReduxState } from "$lib/redux-store/store-root";

function getQueryArgs(args: Record<string, any>) {
  if (!args) return "";
  return "/?" + Object.keys(args).map((argKey) => `${argKey}=${args[argKey]}`);
}

export function generateApis<
  Type extends EntityType,
  FetchManyParams extends Record<string, any> = Record<string, unknown>,
  CreateParams extends Record<string, any> = Record<string, unknown>,
  Entity extends EntityRecord = EntityRecordMapType[Type][StateType.Persistent]
>(
  [entityType, sliceName, endpoint]: [EntityType, keyof RillReduxState, string],
  [addManyAction, addOneAction, updateAction, removeAction]: [
    ActionCreatorWithPreparedPayload<[entities: Array<Entity>], Array<Entity>>,
    ActionCreatorWithPreparedPayload<[entity: Entity], Entity>,
    ActionCreatorWithPreparedPayload<
      [id: string, changes: Partial<Entity>],
      { id: string; changes: Partial<Entity> }
    >,
    ActionCreatorWithPreparedPayload<[id: string], string>
  ],
  validations: Array<ValidationConfig<Entity>>,
  apiHooks?: {
    createHook?: (createdEntity: Entity) => void | Promise<void>;
    deleteHook?: (id: string) => void | Promise<void>;
  }
) {
  return {
    fetchManyApi: createAsyncThunk(
      `${entityType}/fetchManyApi`,
      async (args: FetchManyParams, thunkAPI) => {
        thunkAPI.dispatch(
          addManyAction(
            await fetchWrapper(`${endpoint}${getQueryArgs(args)}`, "GET")
          )
        );
      }
    ),
    createApi: createAsyncThunk(
      `${entityType}/createApi`,
      async (args: CreateParams, thunkAPI) => {
        const createdEntity = await fetchWrapper(endpoint, "PUT", args);
        thunkAPI.dispatch(addOneAction(createdEntity));
        if (apiHooks?.createHook) apiHooks.createHook(createdEntity);
      }
    ),
    updateApi: createAsyncThunk(
      `${entityType}/updateApi`,
      async (
        {
          id,
          changes,
          callback,
        }: { id: string; changes: Partial<Entity>; callback?: () => void },
        thunkAPI
      ) => {
        const validationChanges = await validateEntity(
          thunkAPI.getState()[sliceName].entities[id] as Entity,
          changes,
          validations
        );
        thunkAPI.dispatch(updateAction(id, validationChanges));
        thunkAPI.dispatch(
          updateAction(
            id,
            await fetchWrapper(`${endpoint}/${id}`, "POST", changes)
          )
        );
        if (callback) callback();
      }
    ),
    deleteApi: createAsyncThunk(
      `${entityType}/deleteApi`,
      async (id: string, thunkAPI) => {
        await fetchWrapper(`${endpoint}/${id}`, "DELETE");
        thunkAPI.dispatch(removeAction(id));
        if (apiHooks?.deleteHook) apiHooks.deleteHook(id);
      }
    ),
  };
}
