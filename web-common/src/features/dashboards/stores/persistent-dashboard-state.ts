import type {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { get, type Readable, type Updater } from "svelte/store";

/**
 * Partial state of the dashboard that is stored in local storage.
 */
export type PersistentDashboardState = {
  visibleMeasures?: string[];
  visibleDimensions?: string[];
  leaderboardMeasureName?: string;

  dashboardSortType?: SortType;
  sortDirection?: SortDirection;
};

function persistentDashboardStoreActions(
  update: (this: void, updater: Updater<PersistentDashboardState>) => void,
) {
  function updateKey<K extends keyof PersistentDashboardState>(key: K) {
    return (val: PersistentDashboardState[K]) => {
      update((lup) => {
        lup[key] = val;
        return lup;
      });
    };
  }

  return {
    updateVisibleMeasures: updateKey("visibleMeasures"),
    updateVisibleDimensions: updateKey("visibleDimensions"),
    updateLeaderboardMeasureName: updateKey("leaderboardMeasureName"),
    updateDashboardSortType: updateKey("dashboardSortType"),
    updateSortDirection: updateKey("sortDirection"),
    reset() {
      // cleanup dashboard settings. note that `timeZone` is not reset.
      // it is intentional because it is an old feature
      update((pd) => {
        delete pd.visibleMeasures;
        delete pd.visibleDimensions;
        delete pd.leaderboardMeasureName;
        delete pd.dashboardSortType;
        delete pd.sortDirection;
        return pd;
      });
    },
  };
}

export type PersistentDashboardStore = Readable<PersistentDashboardState> &
  ReturnType<typeof persistentDashboardStoreActions>;
export function createPersistentDashboardStore(storeKey: string) {
  const { subscribe, update } = localStorageStore<PersistentDashboardState>(
    `${storeKey}-persistentDashboardStore`,
    {},
  );
  return {
    subscribe,
    ...persistentDashboardStoreActions(update),
  };
}

// TODO: once we move everything to state-managers we wont need this
let persistentDashboardStore: PersistentDashboardStore;
export function initPersistentDashboardStore(storeKey: string) {
  persistentDashboardStore = createPersistentDashboardStore(storeKey);
}

export function getPersistentDashboardStore() {
  return persistentDashboardStore;
}

export function getPersistentDashboardState(): PersistentDashboardState {
  if (!persistentDashboardStore) return {};
  return get(persistentDashboardStore);
}

export function hasPersistentDashboardData() {
  if (!persistentDashboardStore) return false;
  return Object.keys(get(persistentDashboardStore)).length > 0;
}
