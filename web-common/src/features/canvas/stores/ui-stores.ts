import { derived, writable } from "svelte/store";

const MIN_DELTA = 12;

export const dropZone = (() => {
  const zoneId = writable<string | null>(null);
  const mouseDelta = writable(0);

  return {
    subscribe: derived([zoneId, mouseDelta], ([$dropZone, $mouseDelta]) => {
      if ($mouseDelta < MIN_DELTA) {
        return null;
      } else {
        return $dropZone;
      }
    }).subscribe,
    set: (id: string) => {
      zoneId.set(id);
    },
    setMouseDelta: (amount: number) => {
      mouseDelta.set(amount);
    },
    clear: () => {
      zoneId.set(null);
    },
  };
})();

export const hoveredDivider = (() => {
  const { subscribe, set } = writable<string | null>(null);
  let timeout: ReturnType<typeof setTimeout> | null = null;

  return {
    subscribe,
    set: (id: string) => {
      if (timeout) clearTimeout(timeout);
      set(id);
    },
    reset: () => {
      timeout = setTimeout(() => set(null), 50);
    },
  };
})();

export const activeDivider = (() => {
  const { subscribe, set } = writable<string | null>(null);

  return {
    subscribe,
    set,
    reset: () => set(null),
  };
})();
