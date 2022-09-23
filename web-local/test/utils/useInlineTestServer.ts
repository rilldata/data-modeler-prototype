import { RootConfig } from "$web-local/common/config/RootConfig";
import { DatabaseConfig } from "$web-local/common/config/DatabaseConfig";
import { StateConfig } from "$web-local/common/config/StateConfig";
import { ServerConfig } from "$web-local/common/config/ServerConfig";
import { InlineTestServer } from "./InlineTestServer";
import type { TestServer } from "./TestServer";

/**
 * Creates a server with 'port' in the same process.
 * Make sure to use unique port for each suite.
 *
 * Automatically starts and stops the server.
 * Returns config and the server reference.
 * Check {@link TestServer} and {@link InlineTestServer} for various methods.
 *
 * TODO: auto assign port
 */
export function useInlineTestServer(port: number) {
  const config = new RootConfig({
    database: new DatabaseConfig({ databaseName: ":memory:" }),
    state: new StateConfig({ autoSync: false }),
    server: new ServerConfig({ serverPort: port }),
    projectFolder: "temp/test",
  });
  const inlineServer = new InlineTestServer(config);

  beforeAll(async () => {
    await inlineServer.init();
  });

  afterAll(async () => {
    await inlineServer.destroy();
  });

  return {
    config,
    inlineServer,
  };
}

/**
 * Call this at the top level of suite to load test tables.
 */
export function useTestTables(server: TestServer) {
  beforeAll(async () => {
    await server.loadTestTables();
  });
}

/**
 * Call this at the top level of suite to load a model with given query and name.
 * Make sure to call {@link useTestTables} before this.
 */
export function useTestModel(server: TestServer, query: string, name: string) {
  beforeAll(async () => {
    await server.dataModelerService.dispatch("addModel", [{ query, name }]);
    await server.waitForModels();
  });
}
