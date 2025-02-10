import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { derived, get, writable, type Writable } from "svelte/store";
import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceDeleteFile,
  runtimeServiceGetFile,
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
  type V1GetFileResponse,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import { updateRillYAMLWithOlapConnector } from "../../connectors/code-utils";
import type { OlapDriver } from "../../connectors/olap/olap-config";
import { EMPTY_PROJECT_TITLE } from "../constants";

const ONBOARDING_STATE_FILE_PATH = "/tmp/onboarding-state.json";

export class OnboardingState {
  managementType: Writable<"rill-managed" | "self-managed">;
  olapDriver: Writable<OlapDriver>;
  firstDataSource: Writable<string | undefined>;
  runtimeInstanceId: string;

  constructor() {
    this.runtimeInstanceId = get(runtime).instanceId;
  }

  async isOnboardingStateFilePresent() {
    try {
      await queryClient.fetchQuery({
        queryKey: getRuntimeServiceGetFileQueryKey(this.runtimeInstanceId, {
          path: ONBOARDING_STATE_FILE_PATH,
        }),
        queryFn: () =>
          runtimeServiceGetFile(this.runtimeInstanceId, {
            path: ONBOARDING_STATE_FILE_PATH,
          }),
      });
      return true;
    } catch {
      return false;
    }
  }

  async initializeOnboardingState() {
    // Initialize the state
    this.managementType = writable("rill-managed");
    this.olapDriver = writable("duckdb");
    this.firstDataSource = writable(undefined);

    // Create the onboarding state file
    await this.save();
  }

  async fetchAndParse() {
    let response: V1GetFileResponse;

    try {
      response = await queryClient.fetchQuery({
        queryKey: getRuntimeServiceGetFileQueryKey(this.runtimeInstanceId, {
          path: ONBOARDING_STATE_FILE_PATH,
        }),
        queryFn: () =>
          runtimeServiceGetFile(this.runtimeInstanceId, {
            path: ONBOARDING_STATE_FILE_PATH,
          }),
      });
    } catch (error) {
      if (error?.response?.data?.message?.includes("no such file")) {
        await this.initializeOnboardingState();
        return;
      } else {
        console.error("throwing error", error);
        throw error;
      }
    }

    if (!response.blob) {
      throw new Error("No file content found");
    }

    // parse the state
    const state = JSON.parse(response.blob);

    // set the state
    this.managementType = writable(state.managementType);
    this.olapDriver = writable(state.olapDriver);
    this.firstDataSource = writable(state.firstDataSource);
  }

  async save() {
    const state = {
      managementType: get(this.managementType),
      olapDriver: get(this.olapDriver),
      firstDataSource: get(this.firstDataSource),
    };

    const jsonState = JSON.stringify(state);

    console.log("Saving onboarding state", jsonState);
    await runtimeServicePutFile(this.runtimeInstanceId, {
      path: ONBOARDING_STATE_FILE_PATH,
      blob: jsonState,
    });
  }

  getNumberOfSteps = () => {
    return derived(this.managementType, (managementType) =>
      managementType === "rill-managed" ? 2 : 3,
    );
  };

  getCurrentStep = () => {
    // /welcome/select-connectors => 1
    // /welcome/add-credentials => 2
    // /welcome/make-your-first-dashboard => 3
    return derived(page, (page) => {
      if (page.url.pathname.includes("select-connectors")) {
        return 1;
      }
      if (page.url.pathname.includes("add-credentials")) {
        return 2;
      }
      if (page.url.pathname.includes("make-your-first-dashboard")) {
        return 3;
      }

      return 0;
    });
  };

  selectManagementType(type: "rill-managed" | "self-managed") {
    this.managementType.set(type);

    if (type === "rill-managed") {
      this.selectOLAP("duckdb");
    } else {
      this.selectOLAP("clickhouse");
    }

    // Reset the first data source
    this.firstDataSource.set(undefined);

    void this.save();
  }

  selectOLAP(olap: OlapDriver) {
    this.olapDriver.set(olap);

    // reset the first data source
    this.firstDataSource.set(undefined);

    void this.save();
  }

  toggleFirstDataSource(dataSource: string) {
    if (get(this.firstDataSource) === dataSource) {
      this.firstDataSource.set(undefined);
    } else {
      this.firstDataSource.set(dataSource);
    }

    void this.save();
  }

  async skipFirstSource() {
    // Unpack an empty project
    console.log("Unpacking empty project");
    await runtimeServiceUnpackEmpty(this.runtimeInstanceId, {
      displayName: EMPTY_PROJECT_TITLE,
      force: true,
    });

    // Create a managed OLAP connector file
    console.log("Creating managed OLAP connector file");
    await runtimeServicePutFile(this.runtimeInstanceId, {
      path: `connectors/${get(this.olapDriver)}.yaml`,
      blob: `type: connector

driver: ${get(this.olapDriver)}
managed: true`,
    });

    // Add the chosen OLAP connector to the rill.yaml file
    console.log("Updating rill.yaml file");
    await runtimeServicePutFile(this.runtimeInstanceId, {
      path: "rill.yaml",
      blob: await updateRillYAMLWithOlapConnector(
        queryClient,
        get(this.olapDriver),
      ),
    });

    // Delete the onboarding state file
    console.log("Deleting onboarding state file");
    try {
      await runtimeServiceDeleteFile(this.runtimeInstanceId, {
        path: ONBOARDING_STATE_FILE_PATH,
      });
    } catch (error) {
      console.error(error);
    }

    // Exit the onboarding wizard, and go to the home page
    await goto(`/`);
  }

  // Clean up all files created by the Add Data form
  async cleanUp() {
    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: "connectors",
      force: true,
    });

    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: "rill.yaml",
    });

    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: ".env",
    });

    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: ".gitignore",
    });
  }

  async complete() {
    const isRillManagedOLAP = get(this.managementType) === "rill-managed";

    if (isRillManagedOLAP) {
      // Create a managed connector file
      await runtimeServicePutFile(this.runtimeInstanceId, {
        path: `connectors/${get(this.olapDriver)}.yaml`,
        blob: `type: connector
driver: ${get(this.olapDriver)}
managed: true`,
        create: true,
        createOnly: false,
      });
    }

    // Update the `rill.yaml` file
    await runtimeServicePutFile(this.runtimeInstanceId, {
      path: "rill.yaml",
      blob: await updateRillYAMLWithOlapConnector(
        queryClient,
        get(this.olapDriver),
      ),
      create: true,
      createOnly: false,
    });

    // Delete the onboarding state file
    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: ONBOARDING_STATE_FILE_PATH,
    });
  }
}

export function getOnboardingState() {
  return new OnboardingState();
}
