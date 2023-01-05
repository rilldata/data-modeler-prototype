/* Poll the filesystem under 3 scenarios:
 * - Scenario 1. The user navigates to a new page
 * - Scenario 2. Every X seconds
 * - Scenario 3. The user returns focus to the browser tab
 *
 * It's slightly complicated because we sync a different file depending on the page we're on.
 */

import { afterNavigate, beforeNavigate } from "$app/navigation";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceGetFile,
  RuntimeServiceGetFileQueryResult,
  runtimeServiceListFiles,
  RuntimeServiceListFilesQueryResult,
  runtimeServiceReconcile,
} from "@rilldata/web-common/runtime-client";
import type { Page } from "@sveltejs/kit";
import type { QueryClient } from "@sveltestack/svelte-query";
import { get, Readable, Writable } from "svelte/store";
import type { RuntimeState } from "../../application-state-stores/application-store";
import { invalidateAfterReconcile } from "../../svelte-query/invalidation";
import { getFilePathFromPagePath } from "../../util/entity-mappers";

const SYNC_FILE_SYSTEM_INTERVAL_MILLISECONDS = 5000;

export function syncFileSystemPeriodically(
  queryClient: QueryClient,
  runtimeStore: Writable<RuntimeState>,
  page: Readable<Page<Record<string, string>, string>>
) {
  let syncFileSystemInterval: NodeJS.Timer;
  let syncFileSystemOnVisibleDocument: () => void;
  let afterNavigateRanOnce: boolean;

  afterNavigate(async () => {
    // on first page load, afterNavigate races against onMount, which sets the runtimeInstanceId
    const runtimeInstanceId = await waitForRuntimeInstanceId(runtimeStore);

    // on first page load, afterNavigate runs twice, but we only want to run the below code once
    if (afterNavigateRanOnce) return;

    // Scenario 1: sync when the user navigates to a new page
    syncFileSystem(queryClient, runtimeInstanceId, page, 1);

    // setup Scenario 2: sync every X seconds
    syncFileSystemInterval = setInterval(
      async () => await syncFileSystem(queryClient, runtimeInstanceId, page, 2),
      SYNC_FILE_SYSTEM_INTERVAL_MILLISECONDS
    );

    // setup Scenario 3: sync when the user returns focus to the browser tab
    syncFileSystemOnVisibleDocument = async () => {
      if (document.visibilityState === "visible") {
        await syncFileSystem(queryClient, runtimeInstanceId, page, 3);
      }
    };
    window.addEventListener("focus", syncFileSystemOnVisibleDocument);

    afterNavigateRanOnce = true;
  });

  beforeNavigate(() => {
    // teardown Scenario 2
    clearInterval(syncFileSystemInterval);

    // teardown Scenario 3
    window.removeEventListener("focus", syncFileSystemOnVisibleDocument);

    afterNavigateRanOnce = false;
  });
}

async function syncFileSystem(
  queryClient: QueryClient,
  instanceId: string,
  page: Readable<Page<Record<string, string>, string>>,
  id: number
) {
  let changedPaths = [];

  const pagePath = get(page).url.pathname;
  console.log("syncFileSystem", instanceId, pagePath, id);
  if (isPathToAsset(pagePath)) {
    const filePath = getFilePathFromPagePath(pagePath);
    const isChanged = await refetchFileAndDetectChange(
      queryClient,
      instanceId,
      filePath
    );
    if (isChanged) {
      changedPaths.push(filePath);
    }
  }

  const newFiles = await refetchFileListAndDetectChanges(
    queryClient,
    instanceId
  );
  changedPaths.push(...newFiles);
  changedPaths = [...new Set(changedPaths)]; // removes duplicates, in case a new file is the same as the file on page

  // Option 1: reconcile the entire filesystem
  // const resp = await runtimeServiceReconcile(instanceId, {});
  // invalidateAfterReconcile(queryClient, instanceId, resp);

  // Option 2: reconcile only the changed paths
  if (changedPaths.length) {
    console.log("calling reconcile with changed paths:", changedPaths);
    const resp = await runtimeServiceReconcile(instanceId, { changedPaths });
    invalidateAfterReconcile(queryClient, instanceId, resp);
  }
}

async function refetchFileAndDetectChange(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string
): Promise<boolean> {
  const queryKey = getRuntimeServiceGetFileQueryKey(instanceId, filePath);

  const cachedFile =
    queryClient.getQueryData<RuntimeServiceGetFileQueryResult>(queryKey);
  await queryClient.invalidateQueries(queryKey);
  const freshFile = await queryClient.fetchQuery(queryKey, () =>
    runtimeServiceGetFile(instanceId, filePath)
  );

  // return true if the file has changed
  return freshFile?.blob !== cachedFile?.blob ? true : false;
}

async function refetchFileListAndDetectChanges(
  queryClient: QueryClient,
  instanceId: string
): Promise<string[]> {
  const queryKey = getRuntimeServiceListFilesQueryKey(instanceId);

  const cachedFileList =
    queryClient.getQueryData<RuntimeServiceListFilesQueryResult>(queryKey);
  await queryClient.invalidateQueries(queryKey);
  const freshFileList = await queryClient.fetchQuery(queryKey, () =>
    runtimeServiceListFiles(instanceId, {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    })
  );

  const newFiles = freshFileList?.paths.filter(
    (file) => !cachedFileList?.paths.includes(file)
  );
  return newFiles;
}

function isPathToAsset(path: string) {
  return (
    path.startsWith("/source") ||
    path.startsWith("/model") ||
    path.startsWith("/dashboard")
  );
}

async function waitForRuntimeInstanceId(runtimeStore: Writable<RuntimeState>) {
  return new Promise<string>((resolve) => {
    const unsubscribe = runtimeStore.subscribe((runtimeState) => {
      if (runtimeState.instanceId) {
        unsubscribe();
        resolve(runtimeState.instanceId);
      }
    });
  });
}
