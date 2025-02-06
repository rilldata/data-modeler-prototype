import { test, expect } from "@playwright/test";
import { execSync, spawn } from "child_process";
import { test as RillTest } from "../utils/test";

import path from "node:path";
import { DuckDB, waitForTable } from "../utils/sourceHelpers";

//DuckDB requires 1 worker as it doesnt allow concurrency, will fail otherwise.

test.describe("Read DuckDB Table, then read into Rill", () => {
  async function ensureDuckDBInstalled() {
    try {
      // Check if DuckDB is installed
      execSync("duckdb --version", { stdio: "ignore" });
      console.log("DuckDB is already installed.");
    } catch (err) {
      console.log(err);
      console.log("DuckDB not found. Installing...");
      // Install DuckDB (example for macOS/Linux using wget)
      // https://github.com/duckdb/duckdb/releases/latest/download/duckdb_cli-linux-amd64.zip
      try {
        execSync(
          `
          wget https://github.com/duckdb/duckdb/releases/download/v1.1.3/duckdb_cli-osx-universal.zip &&
          unzip duckdb_cli-osx-universal.zip &&
          chmod +x duckdb &&
          sudo mv duckdb /usr/local/bin/
        `,
          { stdio: "inherit" },
        );

        console.log("DuckDB installed successfully.");
      } catch (error) {
        console.error("DuckDB installation failed:", error);
      }
      console.log("DuckDB installed successfully.");
    }
  }

  test.beforeAll(async () => {
    await ensureDuckDBInstalled();
    const currentDir = process.cwd();
    const dbPath = path.resolve(currentDir, "test/data/playwright.db");
    const commands = [
      `.open ${dbPath}`,
      "select count(*) from sales;",
      "select count(*) from customer_data;",
      ".exit",
    ];
    console.log(`Running DuckDB commands against: ${dbPath}`);

    await new Promise<void>((resolve, reject) => {
      const cli = spawn("duckdb", [], { shell: true });

      let output = "";
      cli.stdout.on("data", (data) => {
        output += data.toString();
      });

      cli.stderr.on("data", (data) => {
        console.error(`Error: ${data}`);
      });

      cli.on("close", (code) => {
        if (code === 0) {
          console.log("DuckDB CLI execution completed successfully.");
          const salesCountMatch = output.match(/100000/); // Expected count for sales
          const customerCountMatch = output.match(/10000/); // Expected count for customer_data

          expect(salesCountMatch).not.toBeNull();
          expect(customerCountMatch).not.toBeNull();
          resolve();
        } else {
          reject(new Error(`DuckDB CLI exited with code ${code}`));
        }
      });

      // Write commands to DuckDB CLI
      commands.forEach((cmd) => cli.stdin.write(`${cmd}\n`));
      cli.stdin.end();
    });
  });

  test("Validate data in DuckDB", async () => {
    console.log("Validation complete in beforeAll.");
  });

  console.log("Checked DuckDB with sales and customer_data tables.");
  console.log("Starting Rill Developer...");

  RillTest("Reading Source into Rill", async ({ page }) => {
    // Test loading the 'sales' table
    await Promise.all([
      waitForTable(page, "sources/sales.yaml", [
        "sale_date",
        "sale_id",
        "duration_ms",
        "customer_id",
        "sales_amount_usd",
        "products",
        "discounts",
        "region",
        "is_online",
      ]),
      DuckDB(page, "sales"), // Ensure the `sales` dataset is loaded
    ]);

    console.log("Sales table validated.");

    // Test loading the 'customer_data' table
    await Promise.all([
      waitForTable(page, "sources/customer_data.yaml", [
        "customer_id",
        "name",
        "email",
        "signup_date",
        "preferences",
        "total_spent_usd",
        "loyalty_tier",
        "is_active",
      ]),
      DuckDB(page, "customer_data"), // Ensure the `customer_data` dataset is loaded
    ]);

    console.log("Customer data table validated.");
  });
});
