import { DataModelerCliCommand } from "$cli/DataModelerCliCommand";
import { Command } from "commander";
import { existsSync, mkdirSync } from "fs";
import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class InitCommand extends DataModelerCliCommand {
    private alreadyInitialised: boolean;

    public getCommand(): Command {
        return this.applyCommonSettings(
            new Command("init"),
            "Initialize a new project either in the current folder or supplied folder.",
        )
            .option("--duckdb <duckDbPath>", "Optional path to existing duckdb database file.")
            .action((opts, command) => {
                const {project} = command.optsWithGlobals();

                const projectPath = project ?? process.cwd();
                InitCommand.makeDirectoryIfNotExists(projectPath);
                this.alreadyInitialised = existsSync(`${projectPath}/state`);

                if (!InitCommand.verifyDuckDbPath(opts.duckdb)) {
                    console.log(`Failed to initialize project under ${projectPath}`);
                    return;
                }

                return this.run({ projectPath, duckDbPath: opts.duckdb });
            });
    }

    protected async sendActions(): Promise<void> {
        if (!existsSync(`${this.projectPath}/models`)) {
            mkdirSync(`${this.projectPath}/models`, {});
        }
        if (!this.alreadyInitialised) {
            // add a single model.
            await this.dataModelerService.dispatch("addModel", [{}]);
            const addedModel = this.dataModelerStateService
                .getEntityStateService(EntityType.Model, StateType.Derived)
                .getCurrentState().entities[0];
            // make that model active.
            await this.dataModelerService.dispatch("setActiveAsset",
                [EntityType.Model, addedModel.id]);
            console.log("\nYou have successfully initialized a new project with Rill Developer.");
        } else {
            console.log("\nA project in this directory has already been initialized.");
        }
        console.log("\nThis application is extremely alpha and we want to hear from you if you have any questions or ideas to share! "+
            "You can reach us in our Rill Discord Channel at https://bit.ly/3NSMKdT.");
    }

    private static makeDirectoryIfNotExists(path: string) {
        if (!existsSync(path)) {
            console.log(`Directory ${path} doesn't exist. Creating the directory.`);
            // Use nodejs methods instead of running commands for making directory
            // This will ensure we can create the directory on all Operating Systems
            mkdirSync(path, { recursive: true });
        } else if (path !== process.cwd()) {
            console.log(`Directory ${path} already exist. Attempting to init the project.`);
        }
    }

    private static verifyDuckDbPath(duckDbPath: string): boolean {
        if (!duckDbPath) return true;

        if (!existsSync(duckDbPath)) {
            console.log(`Duckdb database path provided ${duckDbPath} doesnt exist.`);
            return false;
        }

        console.log(`Importing tables from Duckdb database : ${duckDbPath} .\n` +
            `Make sure to close any write connections to this database before running this.`);
        
        return true;
    }
}
