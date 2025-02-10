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
  athena: { component: AmazonAthena, width: 92, height: 36 },
  azure: { component: MicrosoftAzureBlobStorage, width: 128, height: 30 },
  bigquery: { component: GoogleBigQuery, width: 100, height: 35 },
  clickhouse: { component: ClickHouse, width: 108, height: 18 },
  druid: { component: ApacheDruid, width: 85, height: 22 },
  duckdb: { component: DuckDb, width: 85, height: 24 },
  gcs: { component: GoogleCloudStorage, width: 110, height: 52 },
  https: { component: Https, width: 92, height: 36 },
  local_file: { component: LocalFile, width: 92, height: 36 },
  motherduck: { component: MotherDuck, width: 114, height: 20 },
  mysql: { component: MySql, width: 76, height: 52 },
  pinot: { component: ApachePinot, width: 80, height: 32 },
  postgres: { component: Postgres, width: 121, height: 17 },
  redshift: { component: AmazonRedshift, width: 91, height: 36 },
  s3: { component: AmazonS3, width: 91, height: 39 },
  salesforce: { component: Salesforce, width: 66, height: 46 },
  snowflake: { component: Snowflake, width: 117, height: 35 },
  sqlite: { component: SqLite, width: 75, height: 35 },
};
