import { writable, Writable } from "svelte/store";

export enum DuplicateActions {
  None = "NONE",
  KeepBoth = "KEEP_BOTH",
  Overwrite = "OVERWRITE",
  Cancel = "CANCEL",
}

export const duplicateSourceAction: Writable<DuplicateActions> = writable(
  DuplicateActions.None,
);

export const duplicateSourceName: Writable<string | null> = writable(null);

export const sourceImportedName = writable<string | null>(null);
