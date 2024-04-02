import { get } from "svelte/store";
import { writable } from "svelte/store";
import { browser } from "$app/environment";
import * as devalue from "devalue";

type DashboardName = string;

class HiddenDimensionsStore {
  private store = writable<Set<string>>(new Set([]));
  private dashboardName: DashboardName;

  constructor(dashboardName: DashboardName) {
    this.dashboardName = dashboardName;
    if (browser) {
      const local =
        localStorage.getItem(`hidden-dimensions-${dashboardName}`) || "[]";
      try {
        this.store.set(devalue.parse(local) as Set<string>);
      } catch {
        this.store.set(new Set([]));
      }
    }
  }

  localify = () => {
    if (browser) {
      localStorage.setItem(
        `hidden-dimensions-${this.dashboardName}`,
        devalue.stringify(get(this.store)),
      );
    }
  };

  add = (dimension: string) => {
    const set = get(this.store);
    set.add(dimension);
    this.store.set(set);

    this.localify();
  };

  remove = (dimension: string) => {
    const set = get(this.store);
    set.delete(dimension);
    this.store.set(set);
    this.localify();
  };

  toggle = (dimension: string) => {
    const set = get(this.store);
    if (set.has(dimension)) {
      set.delete(dimension);
    } else {
      set.add(dimension);
    }
    this.store.set(set);
    this.localify();
  };

  subscribe = this.store.subscribe;
}

class AllSelectedDimensions {
  private map = new Map<DashboardName, HiddenDimensionsStore>();

  get(dashboardName: DashboardName) {
    let set = this.map.get(dashboardName);
    if (!set) {
      set = new HiddenDimensionsStore(dashboardName);
      this.map.set(dashboardName, set);
    }

    return set;
  }
}

export const allSelectedDimensions = new AllSelectedDimensions();
