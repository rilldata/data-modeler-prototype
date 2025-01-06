import path from "path";
import { fileURLToPath } from "url";
import { execAsync, spawnAndMatch } from "../utils/spawn";

const timeout = 120_000;

export default async function globalSetup() {
  // Get the repository root directory, the only place from which `rill devtool` is allowed to be run
  const currentDir = path.dirname(fileURLToPath(import.meta.url));
  const repoRoot = path.resolve(currentDir, "../../../");

  // Start the cloud services (except for the UI, which is run by Playwright)
  // This will block until the services are ready
  await spawnAndMatch(
    "rill",
    ["devtool", "start", "e2e", "--reset", "--except", "ui"],
    /All services ready/,
    {
      cwd: repoRoot,
      timeoutMs: timeout,
    },
  );

  // Pull the repositories to be used for testing
  await execAsync(
    "git clone https://github.com/rilldata/rill-examples.git tests/setup/git/repos/rill-examples",
  );
}
