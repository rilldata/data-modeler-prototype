import { test } from "@playwright/test";
import { rmSync, writeFileSync, existsSync, mkdirSync } from "fs";
import { spawn } from "node:child_process";
import type { ChildProcess } from "node:child_process";
import treeKill from "tree-kill";
import { isPortOpen } from "@rilldata/web-local/lib/util/isPortOpen";
import { asyncWaitUntil, waitUntil } from "@rilldata/utils";
import axios from "axios";

const TEST_PROJECT_DIRECTORY = "temp/test-project";
const TEST_PORT = 8083;
const TEST_PORT_GRPC = 9083;

export function startRuntimeForEachTest() {
  let childProcess: ChildProcess;
  let rillShutdown = false;

  test.beforeEach(async () => {
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

    childProcess = spawn("../rill", cmd.split(" "), {
      stdio: "pipe",
      shell: true,
    });
    childProcess.on("error", console.log);
    // Runtime sometimes ends the process but still hasnt released closed the duckdb connection.
    // So wait for the stdio to close. We also need to set `stdio: pipe` and forward the io
    childProcess.on("close", () => {
      rillShutdown = true;
    });
    childProcess.stdout?.on("data", (chunk: Uint8Array) => {
      process.stdout?.write(chunk);
    });
    childProcess.stderr?.on("data", (chunk: Uint8Array) => {
      process.stdout?.write(chunk);
    });

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
  });

  test.afterEach(async () => {
    const processExit = new Promise<void>((resolve) => {
      if (childProcess.pid)
        treeKill(childProcess.pid, () => {
          resolve();
        });
      else {
        resolve();
      }
    });
    await asyncWaitUntil(async () => !(await isPortOpen(TEST_PORT)));
    await processExit;

    await waitUntil(() => rillShutdown, 5000);

    rmSync(TEST_PROJECT_DIRECTORY, {
      force: true,
      recursive: true,
    });
  });
}
