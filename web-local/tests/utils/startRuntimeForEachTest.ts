import { test } from "@playwright/test";
import { asyncWaitUntil, waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { isPortOpen } from "@rilldata/web-local/lib/util/isPortOpen";
import axios from "axios";
import { existsSync, mkdirSync, rmSync, writeFileSync } from "fs";
import type { ChildProcess } from "node:child_process";
import { spawn } from "node:child_process";
import treeKill from "tree-kill";

const TEST_PROJECT_DIRECTORY = "temp/test-project";
const TEST_PORT = 8083;
const TEST_PORT_GRPC = 9083;

interface StartupOptions {
  includeRillYaml: boolean;
}

export function startRuntimeForEachTest(
  options: StartupOptions = { includeRillYaml: true },
) {
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
    if (options.includeRillYaml) {
      // Add `rill.yaml` file to the project repo
      writeFileSync(
        `${TEST_PROJECT_DIRECTORY}/rill.yaml`,
        'compiler: rill-beta\ntitle: "Test Project"',
      );
    }

    const cmd = `start --no-open --port ${TEST_PORT} --port-grpc ${TEST_PORT_GRPC} ${TEST_PROJECT_DIRECTORY}`;

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
