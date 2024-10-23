import { debounce } from "@rilldata/utils";
import { derived, writable } from "svelte/store";
import {
  DEFAULT_INSPECTOR_WIDTH,
  DEFAULT_PREVIEW_TABLE_HEIGHT,
} from "../config";

type WorkspaceLayout = {
  inspector: {
    width: number;
    visible: boolean;
  };
  table: {
    height: number;
    visible: boolean;
  };
  view: "code" | "split" | "viz";
};

class WorkspaceLayoutStore {
  private inspectorVisible = writable<boolean>(true);
  private inspectorWidth = writable<number>(DEFAULT_INSPECTOR_WIDTH);
  private tableVisible = writable<boolean>(true);
  private tableHeight = writable<number>(DEFAULT_PREVIEW_TABLE_HEIGHT);
  public view = writable<"code" | "split" | "viz">("viz");

  constructor(key: string) {
    const history = localStorage.getItem(key);

    if (history) {
      const parsed = JSON.parse(history) as WorkspaceLayout;
      this.inspectorVisible.set(parsed?.inspector?.visible ?? true);
      this.inspectorWidth.set(
        parsed?.inspector?.width ?? DEFAULT_INSPECTOR_WIDTH,
      );
      this.tableHeight.set(
        parsed?.table?.height ?? DEFAULT_PREVIEW_TABLE_HEIGHT,
      );
      this.tableVisible.set(parsed?.table?.visible ?? true);
      if (parsed?.view) this.view.set(parsed.view);
    }

    const debouncer = debounce(
      (v: WorkspaceLayout) => localStorage.setItem(key, JSON.stringify(v)),
      500,
    );

    this.subscribe((v) => debouncer(v));
  }

  subscribe = derived(
    [
      this.inspectorVisible,
      this.inspectorWidth,
      this.tableHeight,
      this.tableVisible,
      this.view,
    ],
    ([
      $inspectorVisible,
      $inspectorWidth,
      $tableHeight,
      $tableVisible,
      $view,
    ]) => {
      const layout: WorkspaceLayout = {
        inspector: {
          visible: $inspectorVisible,
          width: $inspectorWidth,
        },
        table: {
          height: $tableHeight,
          visible: $tableVisible,
        },
        view: $view,
      };
      return layout;
    },
  ).subscribe;

  get inspector() {
    return {
      visible: this.inspectorVisible,
      width: this.inspectorWidth,
      open: () => this.inspectorVisible.set(true),
      close: () => this.inspectorVisible.set(false),
      toggle: () => this.inspectorVisible.update((v) => !v),
    };
  }

  get table() {
    return {
      height: this.tableHeight,
      visible: this.tableVisible,
      open: () => this.tableVisible.set(true),
      close: () => this.tableVisible.set(false),
      toggle: () => this.tableVisible.update((v) => !v),
    };
  }
}

class Workspaces {
  private workspaces = new Map<string, WorkspaceLayoutStore>();

  get = (context: string) => {
    let store = this.workspaces.get(context);

    if (!store) {
      store = new WorkspaceLayoutStore(context);
      this.workspaces.set(context, store);
    }

    return store;
  };
}

export const workspaces = new Workspaces();
