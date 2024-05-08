import { browser } from "$app/environment";
import { createQueryHighlightStore } from "@rilldata/web-common/features/models/query-highlight-store";
import { createApplicationBuildMetadataStore } from "@rilldata/web-common/layout/build-metadata";
import { setContext } from "svelte";

/** This function will initialize the existing node stores and will connect them
 * to the Node server. It is best used in various application layouts to ensure that all children of the layout
 * get the same stores available to them. Once we have moved away from node & the resulting state model, we should
 * deprecate this function.
 */
export function initializeNodeStoreContexts() {
  const queryHighlight = createQueryHighlightStore();

  /** set build-specific metadata as a context.  */
  const buildMetadataStore = createApplicationBuildMetadataStore();
  setContext("rill:app:metadata", buildMetadataStore);

  /** Set the existing node stores, which are consumed through getContext within routes. */
  if (browser) {
    setContext("rill:app:query-highlight", queryHighlight);
  }
}
