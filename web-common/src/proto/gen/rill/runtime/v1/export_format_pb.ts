// @generated by protoc-gen-es v1.4.0 with parameter "target=ts"
// @generated from file rill/runtime/v1/export_format.proto (package rill.runtime.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { proto3 } from "@bufbuild/protobuf";

/**
 * @generated from enum rill.runtime.v1.ExportFormat
 */
export enum ExportFormat {
  /**
   * @generated from enum value: EXPORT_FORMAT_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: EXPORT_FORMAT_CSV = 1;
   */
  CSV = 1,

  /**
   * @generated from enum value: EXPORT_FORMAT_XLSX = 2;
   */
  XLSX = 2,

  /**
   * @generated from enum value: EXPORT_FORMAT_PARQUET = 3;
   */
  PARQUET = 3,
}
// Retrieve enum metadata with: proto3.getEnumType(ExportFormat)
proto3.util.setEnumType(ExportFormat, "rill.runtime.v1.ExportFormat", [
  { no: 0, name: "EXPORT_FORMAT_UNSPECIFIED" },
  { no: 1, name: "EXPORT_FORMAT_CSV" },
  { no: 2, name: "EXPORT_FORMAT_XLSX" },
  { no: 3, name: "EXPORT_FORMAT_PARQUET" },
]);

