import type { V1Connector } from "web-common/src/runtime-client";
import * as yup from "yup";

export function getYupSchema(connector: V1Connector) {
  switch (connector.name) {
    case "s3":
      return yup.object().shape({
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
        path: yup
          .string()
          .matches(/^s3:\/\//, "Must be an S3 URI (e.g. s3://bucket/path)")
          .required("S3 URI is required"),
        aws_region: yup.string().required("Region is required"),
        aws_access_key: yup.string(),
        aws_access_secret: yup.string(),
      });
    case "gcs":
      return yup.object().shape({
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
        path: yup
          .string()
          .matches(/^gs:\/\//, "Must be a GS URI (e.g. gs://bucket/path)")
          .required("GS URI is required"),
      });
    default:
      throw new Error(`Unknown connector: ${connector.name}`);
  }
}

export function toYupFriendlyKey(key: string) {
  return key.replace(/\./g, "_");
}

export function fromYupFriendlyKey(key: string) {
  return key.replace(/_/g, ".");
}
