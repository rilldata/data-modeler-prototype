import { test as base } from "@playwright/test";
import { rmSync, writeFileSync, existsSync, mkdirSync } from "fs";
import { spawn } from "node:child_process";
import treeKill from "tree-kill";
import { getOpenPort } from "./getOpenPort";
import { asyncWaitUntil } from "@rilldata/utils";
import axios from "axios";

const BASE_PROJECT_DIRECTORY = "temp/test-project";

export const test = base.extend({
  page: async ({ page }, use) => {
    const TEST_PORT = await getOpenPort();
    const TEST_PORT_GRPC = await getOpenPort();
    const TEST_PROJECT_DIRECTORY = `${BASE_PROJECT_DIRECTORY}-${TEST_PORT}`;

    rmSync(TEST_PROJECT_DIRECTORY, {
      force: true,
      recursive: true,
    });

    if (!existsSync(TEST_PROJECT_DIRECTORY)) {
      mkdirSync(TEST_PROJECT_DIRECTORY, { recursive: true });
    }

    // Add `rill.yaml` file to the project repo
    writeFileSync(
      `${TEST_PROJECT_DIRECTORY}/rill.yaml`,
      'compiler: rill-beta\ntitle: "Test Project"',
    );

    const cmd = `start --no-open --port ${TEST_PORT} --port-grpc ${TEST_PORT_GRPC} --db ${TEST_PROJECT_DIRECTORY}/stage.db?rill_pool_size=4 ${TEST_PROJECT_DIRECTORY} --env connector.duckdb.external_table_storage=false`;

    const childProcess = spawn("../rill", cmd.split(" "), {
      stdio: "inherit",
      shell: true,
    });

    childProcess.on("error", console.log);

    // Ping runtime until it's ready
    await asyncWaitUntil(async () => {
      try {
        const response = await axios.get(
          `http://localhost:${TEST_PORT}/v1/ping`,
        );
        return response.status === 200;
      } catch {
        return false;
      }
    });

    await page.goto(`http://localhost:${TEST_PORT}`);

    await use(page);

    rmSync(TEST_PROJECT_DIRECTORY, {
      force: true,
      recursive: true,
    });

    const processExit = new Promise((resolve) => {
      childProcess.on("exit", resolve);
    });

    if (childProcess.pid) treeKill(childProcess.pid);

    await processExit;
  },
});
