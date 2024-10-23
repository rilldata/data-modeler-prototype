import { goto } from "$app/navigation";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { waitUntil } from "@rilldata/utils";
import type { V1Resource } from "../../runtime-client";
import { fileArtifacts } from "../entity-management/file-artifacts";
import { ResourceKind } from "../entity-management/resource-selectors";
import { createResourceFile } from "../file-explorer/new-files";

export async function createAndPreviewExplore(
  queryClient: QueryClient,
  instanceId: string,
  metricsViewResource: V1Resource,
) {
  // Create the Explore file
  const filePath = await createResourceFile(
    ResourceKind.Explore,
    metricsViewResource,
  );

  // Wait until the Explore resource is ready
  const fileArtifact = fileArtifacts.getFileArtifact(filePath);
  const resource = fileArtifact.getResource(queryClient, instanceId);
  await waitUntil(() => get(resource).data !== undefined, 10000);
  const name = get(resource).data?.meta?.name?.name;
  if (!name) throw new Error("Failed to create an Explore resource");

  // Check if the Explore has errors
  const hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);

  // Depending on the presence of errors, navigate to the Explore workspace or to the Explore Preview
  if (get(hasErrors)) {
    await goto(`/files${filePath}`);
  } else {
    await goto(`/explore/${name}`);
  }
}
