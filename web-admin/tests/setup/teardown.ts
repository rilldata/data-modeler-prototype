import fs from "fs";
import { execAsync } from "web-common/tests/utils/spawn";
import { test as teardown } from "./base";
import { RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE } from "./constants";

teardown("should stop all services", async () => {
  // Stop the admin and runtime services:
  // 1. Get the process ID from the file
  // 2. Get the process group ID
  // 3. Kill the whole process group
  // 4. Delete the process ID file
  const processID = fs.readFileSync(
    RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE,
    "utf8",
  );
  const { stdout: processGroupID } = await execAsync(
    `ps -o pgid= -p ${processID}`,
  );
  await execAsync(`kill -TERM -${processGroupID}`);
  await execAsync(`rm ${RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE}`);

  // Stop the cloud services
  await execAsync(
    "docker compose -f ../cli/cmd/devtool/data/cloud-deps.docker-compose.yml down --volumes",
  );

  // Remove the test repositories
  await execAsync("rm -rf tests/setup/git/repos");
});
