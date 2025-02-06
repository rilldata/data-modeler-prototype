import { test } from "@playwright/test";
import { test as RillTest } from "../utils/test";
import { cloud, waitForTable } from "../utils/sourceHelpers";

// GCS source ingestion test
// based on public bucket gs://playwright-gcs-qa/*
// Can add more files as required, currently parquet.gz files are erroring so removed.

test.describe("LOAD DATA FROM cloud", () => {
  RillTest("Reading Source into Rill from GCS", async ({ page }) => {
    console.log("Testing cloud sales data ingestion...");
    await Promise.all([
      waitForTable(page, "/sources/sales.yaml", [
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
      cloud(page, "sales.csv", "gcs"),
    ]);
    console.log("Sales table validated.");

    console.log("Testing cloud customer data ingestion...");
    await Promise.all([
      waitForTable(page, "/sources/customer_data.yaml", [
        "customer_id",
        "name",
        "email",
        "signup_date",
        "preferences",
        "total_spent_usd",
        "loyalty_tier",
        "is_active",
      ]),
      cloud(page, "customer_data.csv", "gcs"),
    ]);
    console.log("Customer data table validated.");

    console.log("TESTING VARIOUS TYPES OF FILES ON cloud");
    const AdBidsColumns = ["id", "timestamp", "publisher", "domain"];

    await Promise.all([
      waitForTable(page, "/sources/AdBids_csv.yaml", AdBidsColumns),
      cloud(page, "AdBids_csv.csv", "gcs"),
    ]);
    await Promise.all([
      waitForTable(page, "/sources/AdBids_csv_gz.yaml", AdBidsColumns),
      cloud(page, "AdBids_csv_gz.csv.gz", "gcs"),
    ]);
    await Promise.all([
      waitForTable(page, "/sources/AdBids_parquet.yaml", AdBidsColumns),
      cloud(page, "AdBids_parquet.parquet", "gcs"),
    ]);
    /*  broken parquet.gz
     await Promise.all([
       waitForTable(page, '/sources/AdBids_parquet_gz.yaml', AdBidsColumns),
       cloud(page, 'AdBids_parquet_gz.parquet.gz', 'gcs'),
     ]);
     */

    await Promise.all([
      waitForTable(page, "/sources/AdBids_txt.yaml", AdBidsColumns),
      cloud(page, "AdBids_txt.txt", "gcs"),
    ]);

    const UsersColumns = ["id", "name", "city", "country"];
    const UsersJsonColumns = [
      "id",
      "name",
      "isActive",
      "createdDate",
      "address",
      "tags",
      "projects",
      "scores",
      "flag",
    ];

    await Promise.all([
      waitForTable(page, "/sources/Users_csv.yaml", UsersColumns),
      cloud(page, "Users_csv.csv", "gcs"),
    ]);

    await Promise.all([
      waitForTable(page, "/sources/Users_json.yaml", UsersJsonColumns),
      cloud(page, "Users_json.json", "gcs"),
    ]);

    await Promise.all([
      waitForTable(page, "/sources/Users_parquet.yaml", UsersColumns),
      cloud(page, "Users_parquet.parquet", "gcs"),
    ]);

    const AdImpressionsColumns = ["id", "city", "country", "user_id"];

    await Promise.all([
      waitForTable(
        page,
        "/sources/AdImpressions_parquet.yaml",
        AdImpressionsColumns,
      ),
      cloud(page, "AdImpressions_parquet.parquet", "gcs"),
    ]);
    /*  broken parquet.gz
     await Promise.all([
       waitForTable(page, '/sources/AdImpressions_parquet_gz.yaml', AdImpressionsColumns),
       cloud(page, 'AdImpressions_parquet_gz.parquet.gz', 'gcs'),
     ]);
     */

    await Promise.all([
      waitForTable(
        page,
        "/sources/AdImpressions_tsv.yaml",
        AdImpressionsColumns,
      ),
      cloud(page, "AdImpressions_tsv.tsv", "gcs"),
    ]);
  });
});
