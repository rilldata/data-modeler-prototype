import { ClickHouseClient, createClient } from "@clickhouse/client";
import path from "path";
import { GenericContainer, type StartedTestContainer } from "testcontainers";
import { fileURLToPath } from "url";

export class ClickHouseTestContainer {
  private httpPort = 8123;
  private container: StartedTestContainer;
  private client: ClickHouseClient;

  async start(): Promise<void> {
    const __filename = fileURLToPath(import.meta.url);
    const __dirname = path.dirname(__filename);
    const testDataDir = path.join(__dirname, "../data");

    try {
      this.container = await new GenericContainer(
        "clickhouse/clickhouse-server",
      )
        .withExposedPorts(this.httpPort)
        .withBindMounts([
          {
            source: testDataDir,
            target: "/var/lib/clickhouse/user_files",
          },
        ])
        .start();
    } catch (error) {
      console.error("Failed to start ClickHouse server:", error);
      throw error;
    }
  }

  async stop(): Promise<void> {
    try {
      if (this.container) {
        await this.container.stop();
      }
    } catch (error) {
      console.error("Failed to stop ClickHouse server:", error);
      throw error;
    }
  }

  getClient(): ClickHouseClient {
    if (!this.client) {
      const host = this.container.getHost();
      const port = this.container.getMappedPort(this.httpPort);
      this.client = createClient({
        url: `http://${host}:${port}`,
      });
    }
    return this.client;
  }

  async seed(): Promise<void> {
    const client = this.getClient();

    try {
      await client.command({
        query: "DROP TABLE IF EXISTS ad_bids",
      });

      await client.command({
        query: `CREATE TABLE ad_bids (
        id UInt32,
        timestamp DateTime64(3),
        publisher String,
        domain String,
        bid_price Float64
      ) ENGINE = MergeTree()
      ORDER BY id`,
      });

      await client.command({
        query: `INSERT INTO ad_bids
  SELECT
    id,
    parseDateTime64BestEffort(replaceAll(replaceAll(timestamp, 'T', ' '), 'Z', '')) AS timestamp,
    publisher,
    domain,
    bid_price
  FROM file('/var/lib/clickhouse/user_files/AdBids.csv', 'CSVWithNames',
    'id UInt32, timestamp String, publisher String, domain String, bid_price Float64');`,
      });
    } catch (error) {
      console.error("Failed to seed ClickHouse server:", error);
      throw error;
    }
  }

  getHost(): string {
    return this.container.getHost();
  }

  getPort(): number {
    return this.container.getMappedPort(this.httpPort);
  }
}
