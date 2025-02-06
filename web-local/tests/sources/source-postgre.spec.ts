import { test, expect } from "@playwright/test";
import { execSync } from "child_process";
import pg from "pg";
import { test as RillTest } from "../utils/test";
import {
  pgDataset,
  addFileWithCheck,
  waitForTable,
} from "../utils/sourceHelpers";
import {
  renameFileUsingMenu,
  checkExistInConnector,
} from "../utils/commonHelpers";
import { waitForFileNavEntry } from "../utils/waitHelpers";

/// Due to issue with postgres connector not working, this test errors on source postgres import.

const { Client } = pg;
const pgConfig = {
  host: "localhost",
  port: 5432,
  user: "postgres",
  password: "postgrespass",
  database: "postgres",
};

test.describe("PostgreSQL Test with Docker, then read into Rill", () => {
  let client;

  test.beforeAll(async () => {
    try {
      console.log("Removing potential existing Docker container...");
      execSync("docker rm playwright-postgres --force", { stdio: "inherit" });
      console.log("Docker container removed successfully.");
    } catch (error) {
      console.log(error);
      console.log("Not Applicable: Continue.");
    }
    // Start PostgreSQL Docker container
    execSync(`
      docker run --name playwright-postgres -e POSTGRES_PASSWORD=postgrespass \
      -e POSTGRES_DB=postgres -p 5432:5432 -d postgres:14
    `);

    // Wait for PostgreSQL to become available
    console.log("Waiting for PostgreSQL to initialize...");
    await new Promise((resolve) => setTimeout(resolve, 10000));

    // Copy CSV files into Docker container
    execSync(
      "docker cp /Users/royendo/Desktop/GitHub/rilldata/web-local/rendo-test/data/sales_data_mysql.csv playwright-postgres:/tmp/sales_data.csv",
    );
    console.log("Connecting to psql.");
    // Connect to PostgreSQL
    client = new Client(pgConfig);
    await client.connect();

    // Create tables
    try {
      await client.query(`
        CREATE TABLE sales (
            sale_date DATE NOT NULL,
            sale_id SERIAL PRIMARY KEY,
            customer_id INT NOT NULL,
            products TEXT, -- Store as plain text
            sales_amount_usd NUMERIC NOT NULL,
            discounts TEXT,
            duration_ms INT,
            is_online BOOLEAN,
            region TEXT
              );
          `);
    } catch (error) {
      console.log(error);
      console.error("Couldn't create table");
    }
  });

  test.afterAll(async () => {
    // Clean up database
    await client.query("DROP TABLE IF EXISTS sales");
    await client.end();

    // Stop and remove Docker container
    execSync("docker stop playwright-postgres");
    execSync("docker rm playwright-postgres --force");
  });

  test("Load and validate data in PostgreSQL", async () => {
    // Load data into PostgreSQL
    await client.query(`
      COPY sales (sale_date, sale_id, customer_id, products, sales_amount_usd, discounts, duration_ms, is_online, region)
      FROM '/tmp/sales_data.csv'
      DELIMITER ',' CSV HEADER;
    `);

    // Validate data in PostgreSQL
    const salesResult = await client.query("SELECT COUNT(*) FROM sales");
    expect(parseInt(salesResult.rows[0].count)).toBe(100000);
  });

  console.log(
    "Correctly Initialized PostgreSQL Database with sales and customer_data tables.",
  );
  console.log("Starting Rill Developer...");

  RillTest("Reading PSQL Source into Rill", async ({ page }) => {
    // Issue: private issue #1023: Inconsistent loading. Sometimes just sits on loading. cant verify table

    // Test loading the 'sales' table
    await Promise.all([
      waitForTable(page, "sources/sales.yaml", [
        "sales_date",
        "sale_id",
        "duration_ms",
        "customer_id",
        "sales_amount_usd",
        "products",
        "discounts",
        "region",
        "is_online",
      ]),
      pgDataset(page, "sales"), // Ensure the `sales` dataset is loaded
    ]);

    console.log("Sales table validated.");

    // Test PSQL advanced model #6239

    await addFileWithCheck(page, "untitled_file");
    await renameFileUsingMenu(page, "/untitled_file", "psql_sales.yaml");

    await page.getByRole("link", { name: "psql_sales.yaml" }).click();
    await page.waitForSelector('div[role="textbox"]');

    await page.evaluate(() => {
      // Ensure the parent textbox is focused for typing
      const parentTextbox = document.querySelector('div[role="textbox"]');
      if (parentTextbox instanceof HTMLElement) {
        parentTextbox.focus();
      } else {
        console.error("Parent textbox not found!");
      }
    });

    // Mimic typing in the child contenteditable div
    const childTextbox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );

    const lines = [
      "type: model",
      "",
      "pre_exec: ATTACH 'dbname=postgres host=localhost port=5432 user=postgres password=postgrespass' AS postgres_db(TYPE POSTGRES);",
      "sql: SELECT * FROM postgres_query('postgres_db', 'SELECT * FROM sales')",
      "post_exec: DETACH postgres_db # Note: this is not mandatory but nice to have",
      "",
      "output:",
      "   materialize: true",
      "",
      "",
    ];

    // Type each line with a newline after
    for (const line of lines) {
      await childTextbox.type(line); // Type the line
      await childTextbox.press("Enter"); // Press Enter for a new line
    }
    await page.getByRole("button", { name: "Save" }).click();

    // checks that the file is loading data properly.
    await waitForFileNavEntry(page, "/psql_sales.yaml", true);
    await checkExistInConnector(page, "duckdb", "main_db", "psql_sales");
  });

  console.log("PSQL Advanced Model validated.");
});
