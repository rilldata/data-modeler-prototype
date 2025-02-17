/**
 * The `errorStore` holds the state of any runtime errors, which
 * the `ErrorBoundary` component catches and routes to the `ErrorPage`.
 */

import { derived, writable, type Writable } from "svelte/store";

export interface UserFacingError {
  statusCode: number | null;
  header: string;
  body: string;
  detail?: string;
  fatal?: boolean;
}

export interface ErrorStore extends Writable<UserFacingError> {
  reset: () => void;
}

const createErrorStore = (): ErrorStore => {
  const { subscribe, set, update } = writable({
    statusCode: null,
    header: "",
    body: "",
    fatal: false,
  });

  const reset = (): void => {
    set({ statusCode: null, header: "", body: "", fatal: false });
  };

  return { subscribe, set, update, reset };
};

export const errorStore = createErrorStore();

export const isErrorStoreEmpty = derived(errorStore, ($errorStore) => {
  const { statusCode, header, body } = $errorStore;
  return statusCode === null && header === "" && body === "";
});
