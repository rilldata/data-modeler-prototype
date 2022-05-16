import { TestBase } from "@adityahegde/typescript-test-utils";
import { FunctionalTestBase } from "./FunctionalTestBase";
import {
  ParquetFileTestData,
  CSVFileTestData,
  FileImportTestDataProvider,
  TestDataColumns,
} from "../data/DataLoader.data";
import { DATA_FOLDER } from "../data/generator/data-constants";
import {
  extractFileExtension,
  extractTableName,
} from "$lib/util/extract-table-name";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";
import { SingleTableQuery, TwoTableJoinQuery } from "../data/ModelQuery.data";
import { asyncWait } from "$common/utils/waitUtils";

const UserFile = "test/data/Users.csv";

@TestBase.Suite
export class DataLoaderSpec extends FunctionalTestBase {
  @FunctionalTestBase.BeforeEachTest()
  public async setupTests() {
    await this.clientDataModelerService.dispatch("clearAllTables", []);
  }

  public fileImportTestData(): FileImportTestDataProvider {
    return {
      subData: [ParquetFileTestData, CSVFileTestData],
    };
  }

  // @TestBase.Test("fileImportTestData")
  // public async shouldImportTableFromFile(inputFile: string, cardinality: number, columns: TestDataColumns): Promise<void> {
  //     const actualFilePath = `${DATA_FOLDER}/${inputFile}`;

  //     await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile",
  //         [actualFilePath, `${extractTableName(inputFile)}_${extractFileExtension(inputFile)}`]);
  //     await this.waitForTables();
  //     await asyncWait(250);

  //     const [table, derivedTable] = this.getTables("path", actualFilePath);

  //     expect(table.path).toBe(actualFilePath);
  //     expect(derivedTable.cardinality).toBe(cardinality);

  //     this.assertColumns(derivedTable.profile, columns);
  // }

  @TestBase.Test()
  public async shouldUseTableNameFromArgs(): Promise<void> {
    await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile", [
      UserFile,
      "UsersTable",
    ]);
    await this.waitForTables();

    const [table] = this.getTables("name", "UsersTable");

    expect(table.path).toBe(UserFile);
    expect(table.name).toBe("UsersTable");
  }

  @TestBase.Test()
  public async shouldNotLoadInvalidTable(): Promise<void> {
    const response = await this.clientDataModelerService.dispatch(
      "addOrUpdateTableFromFile",
      ["test/data/AdBids", "AdBidsTableInvalid"]
    );
    await this.waitForTables();

    const [table] = this.getTables("name", "AdBidsTableInvalid");

    expect(table).toBeUndefined();
    expect(response.status).toBe(ActionStatus.Failure);
    expect(response.messages[0].errorType).toBe(ActionErrorType.ImportTable);
  }

  @TestBase.Test()
  public async shouldDropTable(): Promise<void> {
    await Promise.all([
      this.clientDataModelerService.dispatch("addOrUpdateTableFromFile", [
        "test/data/AdBids.csv",
      ]),
      this.clientDataModelerService.dispatch("addOrUpdateTableFromFile", [
        "test/data/AdImpressions.csv",
      ]),
    ]);
    await this.clientDataModelerService.dispatch("addModel", [
      { name: "query_0", query: SingleTableQuery },
    ]);
    await this.waitForTables();
    await this.waitForModels();

    await this.clientDataModelerService.dispatch("dropTable", [
      "AdImpressions",
    ]);
    await asyncWait(100);

    const [table, derivedTable] = this.getTables("tableName", "AdImpressions");
    expect(table).toBeUndefined();
    expect(derivedTable).toBeUndefined();

    const [model] = this.getModels("tableName", "query_0");
    const response = await this.clientDataModelerService.dispatch(
      "updateModelQuery",
      [model.id, TwoTableJoinQuery]
    );
    expect(response.status).toBe(ActionStatus.Failure);
  }
}
