import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { derived, get, writable, type Writable } from "svelte/store";

type ConnectorExplorerItem = {
  connector: string;
  database?: string;
  databaseSchema?: string;
  table?: string;
};

type ConnectorExplorerState = {
  showConnectors: boolean;
  expandedItems: Record<string, boolean>;
  selectedItem: ConnectorExplorerItem | null;
};

export class ConnectorExplorerStore {
  allowNavigateToTable: boolean;
  allowContextMenu: boolean;
  allowSelectTable: boolean;
  allowShowSchema: boolean;
  store: Writable<ConnectorExplorerState>;
  onToggleItem: undefined | ((item: ConnectorExplorerItem) => void) = undefined;

  constructor(
    {
      allowNavigateToTable = true,
      allowContextMenu = true,
      allowShowSchema = true,
      allowSelectTable = false,

      showConnectors = true,
      expandedItems = {},
      selectedItem = null,

      localStorage = true,
    } = {},
    onToggleItem?: (item: ConnectorExplorerItem) => void,
  ) {
    this.allowNavigateToTable = allowNavigateToTable;
    this.allowContextMenu = allowContextMenu;
    this.allowShowSchema = allowShowSchema;
    this.allowSelectTable = allowSelectTable;

    if (onToggleItem) this.onToggleItem = onToggleItem;

    this.store = localStorage
      ? localStorageStore<ConnectorExplorerState>("connector-explorer-state", {
          showConnectors,
          expandedItems,
          selectedItem,
        })
      : writable({ showConnectors, expandedItems, selectedItem });
  }

  createItemIfNotExists(item: ConnectorExplorerItem) {
    this.store.update((state) => {
      const key = getItemKey(item);

      if (key in state.expandedItems) return state; // Item already exists

      return {
        ...state,
        expandedItems: {
          ...state.expandedItems,
          [key]: getDefaultState(item),
        },
      };
    });
  }

  duplicateStore(
    onToggleItem?: (item: ConnectorExplorerItem) => void | Promise<void>,
  ) {
    const state = get(this.store);
    return new ConnectorExplorerStore(
      {
        allowNavigateToTable: false,
        allowContextMenu: false,
        allowShowSchema: false,
        allowSelectTable: true,
        localStorage: false,
        showConnectors: state.showConnectors,
        expandedItems: {},
        selectedItem: null,
      },
      onToggleItem ?? this.onToggleItem,
    );
  }

  toggleExplorer = () =>
    this.store.update((state) => ({
      ...state,
      showConnectors: !state.showConnectors,
    }));

  getItem = (item: ConnectorExplorerItem) => {
    this.createItemIfNotExists(item);

    const key = getItemKey(item);

    return derived(this.store, ($state) => {
      return $state.expandedItems[key];
    });
  };

  toggleItem = (item: ConnectorExplorerItem) => {
    if (this.onToggleItem) this.onToggleItem(item);

    if (item.table) {
      if (this.allowSelectTable) {
        const isSelected = get(this.isItemSelected(item));

        if (isSelected) {
          this.clearSelection();
        } else {
          this.selectItem(item);
        }
      }
      if (!this.allowShowSchema) return;
    }

    this.store.update((state) => {
      const key = getItemKey(item);
      const currentState = state.expandedItems[key] ?? getDefaultState(item);
      return {
        ...state,
        expandedItems: {
          ...state.expandedItems,
          [key]: !currentState,
        },
      };
    });
  };

  deleteItem = (item: ConnectorExplorerItem) =>
    this.store.update((state) => {
      const keyPrefix = getItemKey(item);
      const updatedExpandedItems = Object.fromEntries(
        Object.entries(state.expandedItems).filter(
          ([key]) => !key.startsWith(keyPrefix),
        ),
      );
      return {
        ...state,
        expandedItems: updatedExpandedItems,
      };
    });

  selectItem(item: ConnectorExplorerItem) {
    this.store.update((state) => ({
      ...state,
      selectedItem: item,
    }));
  }

  clearSelection() {
    this.store.update((state) => ({
      ...state,
      selectedItem: null,
    }));
  }

  isItemSelected(item: ConnectorExplorerItem) {
    return derived(this.store, ($state) => {
      if (!this.allowSelectTable) return false;
      const selected = $state.selectedItem;
      if (!selected || !item) return false;

      // Compare by properties rather than reference equality
      return getItemKey(selected) === getItemKey(item);
    });
  }
}

export const connectorExplorerStore = new ConnectorExplorerStore();

// Helpers
function getItemKey(item: ConnectorExplorerItem): string {
  return [item.connector, item.database, item.databaseSchema, item.table]
    .filter(Boolean)
    .join("|");
}

function getDefaultState(item: ConnectorExplorerItem): boolean {
  if (item.databaseSchema || item.table) return false; // Database Schema or Table
  if (item.database) return true; // Database
  return true; // Connector
}
