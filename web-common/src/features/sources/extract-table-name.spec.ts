import { describe, expect, it } from "vitest";
import {
  sanitizeEntityName,
  extractFileExtension,
  extractFileName,
} from "@rilldata/utils";

function generateTestCases(
  fileName: string,
  expectedFileName: string,
  expectedExtension: string,
) {
  return [
    `path/to/file/${fileName}`,
    `/path/to/file/${fileName}`,
    `./path/to/file/${fileName}`,
    fileName,
    `/${fileName}`,
    `./${fileName}`,
  ].map((path) => ({
    title: `fileName=${fileName} path=${path}`,
    expectedFileName,
    expectedExtension,
    path,
  }));
}
const TestCases = [
  ...generateTestCases("22-02-10.parquet", "_22_02_10", ".parquet"),
  ...generateTestCases("-22-02-11.parquet", "_22_02_11", ".parquet"),
  ...generateTestCases("_22-02-12.parquet", "_22_02_12", ".parquet"),
  ...generateTestCases("table.parquet", "table", ".parquet"),
  ...generateTestCases("table.v1.parquet", "table_v1", ".v1.parquet"),
  ...generateTestCases("table", "table", ""),
  ...generateTestCases(
    "table.v1.parquet.gz",
    "table_v1_parquet",
    ".v1.parquet.gz",
  ),
];

describe("extract-table-name", () => {
  describe("should extract and sanitise table name", () => {
    for (const variation of TestCases) {
      it(variation.title, () => {
        expect(sanitizeEntityName(extractFileName(variation.path))).toBe(
          variation.expectedFileName,
        );
      });
    }
  });

  describe("should extract extension", () => {
    for (const variation of TestCases) {
      it(variation.title, () => {
        expect(extractFileExtension(variation.path)).toBe(
          variation.expectedExtension,
        );
      });
    }
  });
});
