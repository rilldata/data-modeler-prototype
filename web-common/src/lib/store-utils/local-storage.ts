import { browser } from "$app/environment";
import { writable } from "svelte/store";
import { debounce } from "@rilldata/utils";

/** Creates a store whose value is stored in localStorage as a string.
 * Only JSON-serializable values can be used.
 */
export function localStorageStore<T>(itemKey: string, defaultValue: T) {
  const store = writable<T>(defaultValue);

  if (browser) {
    const stored = localStorage.getItem(itemKey);
    if (stored !== null) {
      try {
        const parsed = JSON.parse(stored);
        if (parsed !== undefined) {
          store.set(parsed);
        }
      } catch {
        // ignore
      }
    }
  }
  const debouncer = debounce(
    (v: T) => localStorage.setItem(itemKey, JSON.stringify(v)),
    300,
  );
  store.subscribe(debouncer);

  return {
    ...store,
    reset() {
      store.set(defaultValue);
      localStorage.setItem(itemKey, JSON.stringify(defaultValue));
    },
  };
}
