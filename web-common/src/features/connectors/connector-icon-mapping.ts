import AmazonAthena from "@rilldata/web-common/components/icons/connectors/AmazonAthena.svelte";
import AmazonRedshift from "@rilldata/web-common/components/icons/connectors/AmazonRedshift.svelte";
import AmazonS3 from "@rilldata/web-common/components/icons/connectors/AmazonS3.svelte";
import ApacheDruid from "@rilldata/web-common/components/icons/connectors/ApacheDruid.svelte";
import ApachePinot from "@rilldata/web-common/components/icons/connectors/ApachePinot.svelte";
import ClickHouse from "@rilldata/web-common/components/icons/connectors/ClickHouse.svelte";
import DuckDb from "@rilldata/web-common/components/icons/connectors/DuckDB.svelte";
import GoogleBigQuery from "@rilldata/web-common/components/icons/connectors/GoogleBigQuery.svelte";
import GoogleCloudStorage from "@rilldata/web-common/components/icons/connectors/GoogleCloudStorage.svelte";
import Https from "@rilldata/web-common/components/icons/connectors/HTTPS.svelte";
import LocalFile from "@rilldata/web-common/components/icons/connectors/LocalFile.svelte";
import MicrosoftAzureBlobStorage from "@rilldata/web-common/components/icons/connectors/MicrosoftAzureBlobStorage.svelte";
import MotherDuck from "@rilldata/web-common/components/icons/connectors/MotherDuck.svelte";
import MySql from "@rilldata/web-common/components/icons/connectors/MySQL.svelte";
import Postgres from "@rilldata/web-common/components/icons/connectors/Postgres.svelte";
import SqLite from "@rilldata/web-common/components/icons/connectors/SQLite.svelte";
import Salesforce from "@rilldata/web-common/components/icons/connectors/Salesforce.svelte";
import Snowflake from "@rilldata/web-common/components/icons/connectors/Snowflake.svelte";
import { SnowflakeIcon } from "lucide-svelte";
import AmazonS3Icon from "../../components/icons/connectors/AmazonS3Icon.svelte";
import ApacheDruidIcon from "../../components/icons/connectors/ApacheDruidIcon.svelte";
import ApachePinotIcon from "../../components/icons/connectors/ApachePinotIcon.svelte";
import ClickHouseIcon from "../../components/icons/connectors/ClickHouseIcon.svelte";
import DuckDbIcon from "../../components/icons/connectors/DuckDBIcon.svelte";
import GoogleBigQueryIcon from "../../components/icons/connectors/GoogleBigQueryIcon.svelte";

// These are symbols only (no text)
export const symbolIconMapping = {
  bigquery: GoogleBigQueryIcon,
  clickhouse: ClickHouseIcon,
  druid: ApacheDruidIcon,
  duckdb: DuckDbIcon,
  pinot: ApachePinotIcon,
  s3: AmazonS3Icon,
  snowflake: SnowflakeIcon,
};

// These are full logos (w/ text)
export const logoIconMapping = {
  athena: AmazonAthena,
  azure: MicrosoftAzureBlobStorage,
  bigquery: GoogleBigQuery,
  clickhouse: ClickHouse,
  druid: ApacheDruid,
  duckdb: DuckDb,
  gcs: GoogleCloudStorage,
  https: Https,
  local_file: LocalFile,
  motherduck: MotherDuck,
  mysql: MySql,
  pinot: ApachePinot,
  postgres: Postgres,
  redshift: AmazonRedshift,
  s3: AmazonS3,
  salesforce: Salesforce,
  snowflake: Snowflake,
  sqlite: SqLite,
};
