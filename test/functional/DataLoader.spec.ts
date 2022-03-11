import {TestBase} from "@adityahegde/typescript-test-utils";
import {FunctionalTestBase} from "./FunctionalTestBase";
import {
    ParquetFileTestData, CSVFileTestData,
    FileImportTestDataProvider,
    TestDataColumns,
} from "../data/DataLoader.data";
import {DATA_FOLDER} from "../data/generator/data-constants";
import {execSync} from "node:child_process";
import { extractFileExtension, extractTableName } from "$lib/util/extract-table-name";

const AdBidsFile = "data/AdBids.parquet";
const AdImpressionsFile = "data/AdImpressions.parquet";
const UserFile = "data/Users.csv";

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

    @TestBase.Test("fileImportTestData")
    public async shouldImportTableFromFile(inputFile: string, cardinality: number, columns: TestDataColumns): Promise<void> {
        const actualFilePath = `${DATA_FOLDER}/${inputFile}`;

        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile",
            [actualFilePath, `${extractTableName(inputFile)}_${extractFileExtension(inputFile)}`]);
        await this.waitForTables();

        const [table, derivedTable] = this.getTables("path", actualFilePath);

        expect(table.path).toBe(actualFilePath);
        expect(derivedTable.cardinality).toBe(cardinality);

        this.assertColumns(derivedTable.profile, columns);
    }

    @TestBase.Test()
    public async shouldUseTableNameFromArgs(): Promise<void> {
        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile",
          [UserFile, "UsersTable"]);
        await this.waitForTables();

        const [table] = this.getTables("name", "UsersTable");

        expect(table.path).toBe(UserFile);
        expect(table.name).toBe("UsersTable");
    }

    @TestBase.Test()
    public async shouldNotLoadInvalidTable(): Promise<void> {
        await this.clientDataModelerService.dispatch("addOrUpdateTableFromFile",
          ["data/AdBids", "AdBidsTableInvalid"]);
        await this.waitForTables();

        const [table] = this.getTables("name", "AdBidsTableInvalid");

        expect(table).toBeUndefined();
    }
}
