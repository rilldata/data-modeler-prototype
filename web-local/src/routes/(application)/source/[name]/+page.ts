import { EntityType } from "@rilldata/web-common/lib/entity";
import {
  runtimeServiceGetCatalogEntry,
  runtimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
import { error } from "@sveltejs/kit";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const localConfig = await runtimeServiceGetConfig();

  // try to get the catalog entry.
  let catalogEntry;
  try {
    catalogEntry = await runtimeServiceGetCatalogEntry(
      localConfig.instance_id,
      params.name
    );
    // if this is a valid catalog entry, then we can return it.
    return {
      sourceName: params.name,
      path: catalogEntry?.entry?.source?.properties?.path,
      embedded: catalogEntry?.entry?.embedded,
    };
  } catch (err) {
    // no-op. we'll try to get the file below.
  }

  try {
    await runtimeServiceGetFile(
      localConfig.instance_id,
      getFilePathFromNameAndType(params.name, EntityType.Table)
    );

    return { sourceName: params.name };
  } catch (e) {
    if (e.response?.status && e.response?.data?.message) {
      throw error(e.response.status, e.response.data.message);
    } else {
      console.error(e);
      throw error(500, e.message);
    }
  }
}
