import { writable } from "svelte/store";
import type { Writable } from "svelte/store";

class LastVisitedURLS {
  private map = new Map<string, Writable<string>>();

  update(key: string, value: string) {
    this.map.get(key)?.set(value);
  }

  get(key: string) {
    let existing = this.map.get(key);
    if (!existing) {
      existing = writable(key);
      this.map.set(key, existing);
    }

    return existing;
  }
}

export const lastVisitedURLs = new LastVisitedURLS();
