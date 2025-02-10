import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

export function compileClickhouseSourceConnectorFile(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
) {
  switch (connector.name) {
    case "azure":
      return `sql: SELECT * FROM azure('${formValues.path as string}');

output:
  materialize: true`;
    case "gcs":
      return `sql: SELECT * FROM gcs('${formValues.path as string}', '{{ .env.connector.gcs.hmac_key }}', '{{ .env.connector.gcs.hmac_secret }}');

output:
  materialize: true`;
    case "s3":
      return `sql: SELECT * FROM s3('${formValues.path as string}', '{{ .env.connector.s3.aws_access_key_id }}', '{{ .env.connector.s3.aws_secret_access_key }}');

output:
  materialize: true`;
    case "postgres":
      return `sql: SELECT * FROM postgresql('${formValues.host as string}:${formValues.port as string}', '${formValues.database as string}', '${formValues.table as string}', '${formValues.user as string}', '{{ .env.connector.postgres.password }}'

output:
  materialize: true`;
    case "mysql":
      return `sql: SELECT * FROM mysql('${formValues.host as string}:${formValues.port as string}', '${formValues.database as string}', '${formValues.user as string}', '{{ .env.connector.mysql.password }}');

output:
  materialize: true`;
    default:
      throw new Error(`Unsupported connector: ${connector.name}`);
  }
}
