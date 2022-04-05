import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";=
import { FunctionalTestBase } from "./FunctionalTestBase";
import type { DatabaseService } from "$common/database-service/DatabaseService";
import type { TimeGrain } from "$common/database-service/DatabaseColumnActions";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import { StateConfig } from "$common/config/StateConfig";
import { dataModelerServiceFactory } from "$common/serverFactory";

import { timeGrainSeriesData } from "../data/TimeGrain.data"
import type { GeneratedTimeseriesTestCase } from "../data/TimeGrain.data"

const SYNC_TEST_FOLDER = "temp/sync-test";

function ctas(table, select_statement, temp = true) {
    return `CREATE ${temp ? 'TEMPORARY' : ''} VIEW ${table} AS (${select_statement})`
}

function generateSeries(table:string, start:string, end:string, interval:string) {
    return ctas(table, `SELECT generate_series as ts from generate_series(TIMESTAMP '${start}', TIMESTAMP '${end}', interval ${interval})`)
}

@FunctionalTestBase.Suite
export class StateSyncServiceSpec extends FunctionalTestBase  {
    protected databaseService: DatabaseService;

    public async setup(): Promise<void> {
        const config = new RootConfig({
            database: new DatabaseConfig({ databaseName: ":memory:" }),
            state: new StateConfig({ autoSync: true, syncInterval: 50 }),
            projectFolder: SYNC_TEST_FOLDER, profileWithUpdate: false,
        });
        await super.setup(config);
        const secondServerInstances = dataModelerServiceFactory(config);
        this.databaseService = secondServerInstances.dataModelerService.getDatabaseService();
        await this.databaseService.init();
    }
    public seriesGeneratedTimegrainData(): DataProviderData<[GeneratedTimeseriesTestCase]> {
        return timeGrainSeriesData;
    }

    @TestBase.Test("seriesGeneratedTimegrainData")
    public async shouldIdentifyTimegrain(args:GeneratedTimeseriesTestCase) {
        // generate the temporary table.
        // @ts-ignore
        await this.databaseService.databaseClient.execute(generateSeries(args.table, args.start, args.end, args.interval));
        const timeGrain = await this.databaseService.dispatch("estimateTimeGrain", [args.table, "ts"]) as TimeGrain;
        expect(args.expectedTimeGrain).toBe(timeGrain);
    }
}
