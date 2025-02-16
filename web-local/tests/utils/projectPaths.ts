import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import type { V1ListResourcesResponse } from "@rilldata/web-common/runtime-client";
import { rmSync } from "fs";
import { join } from "node:path";
import { cpSync } from "node:fs";
import treeKill from "tree-kill";
import { spawnAndMatch } from "web-common/tests/utils/spawn";
import { getOpenPort } from "web-local/tests/utils/getOpenPort";

export const BASE_PROJECT_DIRECTORY = "temp/test-project";
export const TEST_PROJECTS = "tests/data/projects";

/**
 * Prep a project by pre-ingesting and building the tmp folder.
 */
export async function prepProject(name: string) {
  const TEST_PROJECT_SRC_DIRECTORY = join(TEST_PROJECTS, name);
  const TEST_PROJECT_DIRECTORY = join(BASE_PROJECT_DIRECTORY, name);
  const TEST_PORT = await getOpenPort();
  const TEST_PORT_GRPC = await getOpenPort();

  rmSync(TEST_PROJECT_DIRECTORY, { force: true, recursive: true });
  cpSync(TEST_PROJECT_SRC_DIRECTORY, TEST_PROJECT_DIRECTORY, {
    recursive: true,
    force: true,
  });

  const { process } = await spawnAndMatch(
    "../rill",
    [
      "start",
      "--no-open",
      "--port",
      "" + TEST_PORT,
      "--port-grpc",
      "" + TEST_PORT_GRPC,
      TEST_PROJECT_DIRECTORY,
    ],
    new RegExp("Serving Rill on: http://localhost"),
  );

  await asyncWaitUntil(
    async () => {
      const resp = await fetch(
        "http://localhost:" + TEST_PORT + "/v1/instances/default/resources",
      );
      if (!resp.ok) return false;
      const json = (await resp.json()) as V1ListResourcesResponse;
      const relevantResources = json.resources?.filter(
        (r) => r.meta?.name?.kind !== "rill.runtime.v1.ProjectParser",
      );
      return (
        relevantResources?.every(
          (r) => r.meta?.reconcileStatus === "RECONCILE_STATUS_IDLE",
        ) ?? false
      );
    },
    5 * 60 * 1000,
    1000,
  );

  const processExit = new Promise((resolve) => {
    process.on("exit", resolve);
  });
  if (process.pid) treeKill(process.pid);
  await processExit;
}
