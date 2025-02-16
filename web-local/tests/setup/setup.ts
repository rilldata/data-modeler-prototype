import { prepProject, TEST_PROJECTS } from "web-local/tests/utils/projectPaths";
import { test as setup } from "../utils/test";
import { readdirSync } from "node:fs";

setup("should prep projects", async () => {
  await Promise.all(readdirSync(TEST_PROJECTS).map(prepProject));
});
