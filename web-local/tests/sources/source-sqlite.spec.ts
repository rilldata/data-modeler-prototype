import { test, expect } from "@playwright/test";
import sqlite3 from "sqlite3";
import { open } from "sqlite";
import { test as RillTest } from "../utils/test";
import { sqlLiteDataset, waitForTable } from "../utils/sourceHelpers";
import fs from "fs";
import csv from "csv-parser";
import { fileURLToPath } from "url";
import path from "path";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
export const DataPath = path.join(__dirname, "../data");

let db;

async function readCSV<T = any>(filePath: string): Promise<T[]> {
  const rows: T[] = []; // Explicitly type the array
  return new Promise((resolve, reject) => {
    fs.createReadStream(filePath)
      .pipe(csv())
      .on("data", (row: T) => {
        rows.push(row);
      })
      .on("end", () => {
        resolve(rows);
      })
      .on("error", (err) => {
        reject(err);
      });
  });
}

test.describe("SQLite Test, then read into Rill", () => {
  test.beforeAll(async () => {
    test.setTimeout(60000); // Set timeout to 60 seconds

    // Create and initialize SQLite database
    db = await open({
      filename: "mydb.sqlite", // Use in-memory database
      driver: sqlite3.Database,
    });
    if (!db) {
      throw new Error("Failed to open SQLite database");
    }

    // Create tables
    await db.exec(`
      CREATE TABLE IF NOT EXISTS sales (
        sale_date DATE NOT NULL, -- Date of the sale event
        sale_id INTEGER PRIMARY KEY, -- Unique identifier for each sale
        customer_id INTEGER NOT NULL, -- Identifier for the customer
        products TEXT, -- List of products purchased (stored as JSON string)
        sales_amount_usd REAL NOT NULL, -- Total sales amount in USD
        discounts TEXT, -- Discounts applied (stored as JSON string)
        duration_ms INTEGER, -- Time spent on the transaction (in milliseconds)
        is_online BOOLEAN, -- Whether the sale was made online
        region TEXT -- Region where the sale occurred
      );
    `);

    await db.exec(`
      CREATE TABLE IF NOT EXISTS customer_data (
        customer_id INTEGER PRIMARY KEY, -- Unique identifier for each customer
        name TEXT NOT NULL, -- Customer name
        email TEXT NOT NULL, -- Customer email address
        signup_date DATE NOT NULL, -- Date the customer signed up
        preferences TEXT, -- Customer preferences (stored as JSON string)
        total_spent_usd REAL NOT NULL, -- Total money spent by the customer in USD
        loyalty_tier TEXT, -- Loyalty tier (e.g., Gold, Silver, Bronze)
        is_active BOOLEAN -- Whether the customer account is active
      );
    `);

    // Load data from CSVs
    const salesData = await readCSV(`${DataPath}/sales_data_sqlite.csv`);
    const customerData = await readCSV(`${DataPath}/customer_data_sqlite.csv`);

    for (const row of salesData) {
      await db.run(
        `
        INSERT INTO sales (sale_date, sale_id, customer_id, products, sales_amount_usd, discounts, duration_ms, is_online, region)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        [
          row.sale_date,
          row.sale_id,
          row.customer_id,
          row.products,
          row.sales_amount_usd,
          row.discounts,
          row.duration_ms,
          row.is_online,
          row.region,
        ],
      );
    }

    for (const row of customerData) {
      await db.run(
        `
        INSERT INTO customer_data (customer_id, name, email, signup_date, preferences, total_spent_usd, loyalty_tier, is_active)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        [
          row.customer_id,
          row.name,
          row.email,
          row.signup_date,
          row.preferences,
          row.total_spent_usd,
          row.loyalty_tier,
          row.is_active,
        ],
      );
    }
  });

  test.afterAll(async () => {
    // Close the SQLite database
    await db.close();
    fs.unlinkSync("mydb.sqlite"); // Deletes the database file
    console.log("Database file deleted");
  });

  test("Load and validate data in SQLite", async () => {
    // Validate data in the database
    const sales = await db.all("SELECT * FROM sales");
    expect(sales).toHaveLength(100000);

    const customerData = await db.all("SELECT * FROM customer_data");
    expect(customerData).toHaveLength(10000);
  });

  console.log(
    "Correctly Initialized SQLite Database with sales and customer_data tables.",
  );
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
      sqlLiteDataset(page, "sales"), // Ensure the `sales` dataset is loaded
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
      sqlLiteDataset(page, "customer_data"), // Ensure the `customer_data` dataset is loaded
    ]);

    console.log("Customer data table validated.");
  });
});
