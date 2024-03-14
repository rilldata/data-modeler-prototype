---
title: Configure connector credentials
sidebar_label: Configure credentials
sidebar_position: 10
---

Rill requires credentials to connect to remote data sources such as private buckets (S3, GCS, Azure) or other DuckDB sources (MotherDuck).

When running Rill locally, Rill attempts to find existing credentials configured on your computer. When deploying projects to Rill Cloud, you must explicitly provide service account credentials with correct access permissions.

For instructions on how to create a service account and set credentials in Rill Cloud, see our reference docs for:

- [Amazon S3](./s3.md)
- [Google Cloud Storage (GCS)](./gcs.md)
- [Azure Blob Storage (Azure)](./azure.md)
- [Postgres](./postgres.md)
- [MotherDuck](./motherduck.md)
- [Amazon Athena](./athena.md)
- [Amazon Redshift](./redshift.md)
- [BigQuery](./bigquery.md)
- [Snowflake](./snowflake.md)
- [Salesforce](./salesforce.md)
