import { describe, expect } from "@jest/globals";
import type { PersistentModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { PersistentTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import {
  getDuplicateNameChecker,
  getIncrementedNameGetter,
} from "@rilldata/web-local/common/utils/duplicateNameUtils";

function testDuplicateNameChecker(
  name: string,
  modelNames: Array<string>,
  tableNames: Array<string>,
  isDuplicate: boolean
) {
  expect(
    getDuplicateNameChecker(
      name,
      modelNames.map((modelName) => ({
        name: modelName,
      })) as Array<PersistentModelEntity>,
      tableNames.map((tableName) => ({
        name: tableName,
      })) as Array<PersistentTableEntity>
    )
  ).toEqual(isDuplicate);
}

function testIncrementedNameGetter(
  name: string,
  modelNames: Array<string>,
  tableNames: Array<string>,
  expectedName: string
) {
  expect(
    getIncrementedNameGetter(
      name,
      modelNames.map((modelName) => ({
        name: modelName,
      })) as Array<PersistentModelEntity>,
      tableNames.map((tableName) => ({
        name: tableName,
      })) as Array<PersistentTableEntity>
    )
  ).toEqual(expectedName);
}

describe("duplicateNameUtils", () => {
  describe("getDuplicateNameChecker", () => {
    it("happy path", () => {
      testDuplicateNameChecker("none", ["foo"], ["bar"], false);
      testDuplicateNameChecker("foo", ["foo"], ["bar"], true);
      testDuplicateNameChecker("bar", ["foo"], ["bar"], true);
    });

    it("case insensitive", () => {
      testDuplicateNameChecker("None", ["foo"], ["bar"], false);
      testDuplicateNameChecker("Foo", ["foo"], ["bar"], true);
      testDuplicateNameChecker("Bar", ["foo"], ["bar"], true);
    });
  });

  describe("getIncrementedNameGetter", () => {
    it("happy path", () => {
      testIncrementedNameGetter("none", ["foo"], ["bar"], "none");
      testIncrementedNameGetter("foo", ["foo"], ["bar"], "foo_1");
      testIncrementedNameGetter("bar", ["foo"], ["bar"], "bar_1");
    });

    it("case insensitive", () => {
      testIncrementedNameGetter("None", ["foo"], ["bar"], "None");
      testIncrementedNameGetter("Foo", ["foo"], ["bar"], "Foo_1");
      testIncrementedNameGetter("Bar", ["foo"], ["bar"], "Bar_1");
    });

    it("mixed", () => {
      testIncrementedNameGetter(
        "FOO",
        ["foo", "BAR_1"],
        ["bar", "Foo_1"],
        "FOO_2"
      );
      testIncrementedNameGetter(
        "BAR",
        ["foo", "BAR_1"],
        ["bar", "Foo_1"],
        "BAR_2"
      );
    });

    it("gaps", () => {
      testIncrementedNameGetter("Foo", ["foo"], ["foo_2"], "Foo_1");
    });
  });
});
