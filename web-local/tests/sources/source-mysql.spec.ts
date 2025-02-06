import { test, expect } from "@playwright/test";
import { execSync } from "child_process";
import path from "path";
import { fileURLToPath } from "url";
import mysql from "mysql2/promise";
import { test as RillTest } from "../utils/test";
import { mySQLDataset, waitForTable } from "../utils/sourceHelpers";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
export const DataPath = path.join(__dirname, "../data");

/// Due to bug where the source is loaded but its an endless spinning wheel, waitforTable function hangs. Need to fix this.

/// MySQL source ingestion test
/// Using Docker, spin up a mysql db and add data into the DB
/// Read data into Rill

const mysqlConfig = {
  host: "localhost",
  port: 3306,
  user: "root",
  password: "rootpass",
  database: "default",
};

test.describe("MySQL Test with Docker, then read into Rill", () => {
  let connection;

  test.beforeAll(async () => {
    try {
      console.log("Removing potential existing Docker container...");
      execSync("docker rm playwright-mysql --force", { stdio: "inherit" });
      console.log("Docker container removed successfully.");
    } catch (error) {
      console.log(error);
      console.log("Not Applicable: Continue.");
    }

    // Start MySQL Docker container
    execSync(`
      docker run --name playwright-mysql -e MYSQL_ROOT_PASSWORD=rootpass \
      -e MYSQL_DATABASE=default -e MYSQL_USER=default \
      -e MYSQL_PASSWORD=testpass -p 3306:3306 -d mysql:8.0 
    `);
    ``;
    // Wait for MySQL to become available
    console.log("Waiting for MySQL to initialize...");
    await new Promise((resolve) => setTimeout(resolve, 10000));
    // Dynamically resolve the file path

    execSync(
      `docker cp ${DataPath}/customer_data_mysql.csv playwright-mysql:/var/lib/mysql-files/`,
    );
    execSync(
      `docker cp ${DataPath}/sales_data_mysql.csv playwright-mysql:/var/lib/mysql-files/`,
    );
    // Connect to MySQL
    connection = await mysql.createConnection(mysqlConfig);

    // Create tables
    await connection.execute(`
        CREATE TABLE IF NOT EXISTS sales (
            sale_date DATE NOT NULL,                           -- Date of the sale event
            sale_id INT PRIMARY KEY,                          -- Unique identifier for each sale
            customer_id INT NOT NULL,                         -- Identifier for the customer
            products JSON,                                    -- List of products purchased (stored as JSON)
            sales_amount_usd DOUBLE NOT NULL,                 -- Total sales amount in USD
            discounts JSON,                                   -- Discounts applied (stored as JSON)
            duration_ms INT,                                  -- Time spent on the transaction (in milliseconds)
            is_online BOOLEAN,                                -- Whether the sale was made online
            region VARCHAR(255)                               -- Region where the sale occurred
        );
    `);

    await connection.execute(`
      CREATE TABLE IF NOT EXISTS customer_data (
    customer_id INT PRIMARY KEY,                     -- Unique identifier for each customer
    name VARCHAR(255) NOT NULL,                     -- Customer name
    email VARCHAR(255)  NOT NULL,                   -- Customer email address
    signup_date DATE NOT NULL,                      -- Date the customer signed up
    preferences JSON,                               -- Customer preferences (stored as JSON)
    total_spent_usd DOUBLE NOT NULL,                -- Total money spent by the customer in USD
    loyalty_tier VARCHAR(50),                       -- Loyalty tier (e.g., Gold, Silver, Bronze)
    is_active BOOLEAN                               -- Whether the customer account is active
            );
    `);
  });

  test.afterAll(async () => {
    // Clean up database
    await connection.execute("DROP TABLE IF EXISTS customer_data");
    await connection.execute("DROP TABLE IF EXISTS sales");
    await connection.end();

    // Stop and remove Docker container
    execSync("docker stop playwright-mysql");
    execSync("docker rm playwright-mysql --force");
  });

  test("Load and validate data in MySQL", async () => {
    // Insert data into MySQL
    await connection.query(`
        LOAD DATA INFILE '/var/lib/mysql-files/sales_data_mysql.csv'
        INTO TABLE sales
        FIELDS TERMINATED BY ',' 
        ENCLOSED BY '"'
        LINES TERMINATED BY '\n'
        IGNORE 1 ROWS
        (sale_date, sale_id, customer_id, products, sales_amount_usd, discounts, duration_ms, is_online, region);

        `);

    const [sales] = await connection.execute("SELECT * FROM sales");
    expect(sales).toHaveLength(100000);

    await connection.query(`
        LOAD DATA INFILE '/var/lib/mysql-files/customer_data_mysql.csv'
        INTO TABLE customer_data
        FIELDS TERMINATED BY ',' 
        ENCLOSED BY '"'
        LINES TERMINATED BY '\n'
        IGNORE 1 ROWS
        (customer_id, name, email, signup_date, preferences, total_spent_usd, loyalty_tier, is_active);

        `);

    const [customer_data] = await connection.execute(
      "SELECT * FROM customer_data",
    );
    expect(customer_data).toHaveLength(10000);
  });
  console.log(
    "Correctly Initialized mySQL Database with sales and customer_data tables.",
  );
  console.log("Starting Rill Developer...");

  RillTest("Reading MYSQL source into Rill", async ({ page }) => {
    // Issue: private issue #1023: Inconsistent loading. Sometimes just sits on loading. cant verify table
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
      mySQLDataset(page, "sales"), // Ensure the `sales` dataset is loaded
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
      mySQLDataset(page, "customer_data"), // Ensure the `customer_data` dataset is loaded
    ]);

    console.log("Customer data table validated.");
  });
});
