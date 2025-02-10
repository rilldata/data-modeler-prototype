import { page } from "$app/stores";
import type { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { derived, writable } from "svelte/store";
import {
  MetricsEventScreenName,
  ScreenToEntityMap,
} from "../metrics/service/MetricsTypes";

export interface ActiveEntity {
  type: EntityType;
  id?: string;
  name: string;
}

/**
 * App wide store to store metadata
 * Currently caches active entity from URL
 */
interface AppStore {
  activeEntity: ActiveEntity | undefined;
  previousActiveEntity: ActiveEntity | undefined;
}

export const appScreen = derived(page, ($page) => {
  let activeEntity: {
    name: string;
    type: MetricsEventScreenName;
  };
  switch ($page.route.id) {
    case "/(application)/(developer)":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Home,
      };
      break;
    case "/(application)/(developer)/files/sources/[name]":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Source,
      };
      break;
    case "/(application)/(developer)/files/models/[name]":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Model,
      };
      break;
    case "/(application)/(developer)/files/dashboards/[name]":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Dashboard,
      };
      break;
    case "/(application)/(developer)/files/metrics/[name]":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.MetricsDefinition,
      };
      break;
    case "/welcome":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Splash,
      };
      break;
    case "/[organization]/[project]/[dashboard]":
      activeEntity = {
        name: $page?.params?.dashboard,
        type: MetricsEventScreenName.Dashboard,
      };
      break;
    default:
      // Return home as default
      activeEntity = { name: "", type: MetricsEventScreenName.Home };
  }

  appStore.setActiveEntity(
    activeEntity.name,
    ScreenToEntityMap[activeEntity.type],
  );
  return activeEntity;
});

// App store is being utilized for making previous entity inactive in the HTTP request queue
const { update } = writable({
  activeEntity: undefined,
  previousActiveEntity: undefined,
} as AppStore);

const appStoreReducers = {
  setActiveEntity(name: string, type: EntityType) {
    update((state) => {
      state.previousActiveEntity = state.activeEntity;

      if (state.previousActiveEntity) {
        httpRequestQueue.inactiveByName(state.previousActiveEntity.name);
      }
      state.activeEntity = {
        name,
        type,
      };
      return state;
    });
  },
};

export const appStore: typeof appStoreReducers = {
  ...appStoreReducers,
};
