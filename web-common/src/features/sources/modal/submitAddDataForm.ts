import { goto, invalidate } from "$app/navigation";
import { page } from "$app/stores";
import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
import type { QueryClient } from "@tanstack/query-core";
import { get } from "svelte/store";
import { waitUntil } from "../../../lib/waitUtils";
import { behaviourEvent } from "../../../metrics/initMetrics";
import {
  BehaviourEventAction,
  BehaviourEventMedium,
} from "../../../metrics/service/BehaviourEventTypes";
import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
import {
  type V1ConnectorDriver,
  connectorServiceOLAPListTables,
  getConnectorServiceOLAPListTablesQueryKey,
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import { compileClickhouseSourceConnectorFile } from "../../connectors/clickhouse/source-templates";
import {
  compileConnectorYAML,
  updateDotEnvWithSecrets,
  updateRillYAMLWithOlapConnector,
} from "../../connectors/code-utils";
import type { OlapDriver } from "../../connectors/olap/olap-config";
import { getFileAPIPathFromNameAndType } from "../../entity-management/entity-mappers";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import { getName } from "../../entity-management/name-utils";
import { ResourceKind } from "../../entity-management/resource-selectors";
import { EntityType } from "../../entity-management/types";
import { EMPTY_PROJECT_TITLE } from "../../welcome/constants";
import { isProjectInitialized } from "../../welcome/is-project-initialized";
import { compileSourceYAML, maybeRewriteToDuckDb } from "../sourceUtils";
import type { AddDataFormType } from "./types";
import { fromYupFriendlyKey } from "./yupSchemas";

interface AddDataFormValues {
  // name: string; // Commenting out until we add user-provided names for Connectors
  [key: string]: unknown;
}

export async function submitAddDataForm(
  queryClient: QueryClient,
  formType: AddDataFormType,
  connector: V1ConnectorDriver,
  values: AddDataFormValues,
  olapDriver?: OlapDriver, // only relevant for "source" formType
): Promise<string> {
  const instanceId = get(runtime).instanceId;

  // Emit telemetry
  await behaviourEvent?.fireSourceTriggerEvent(
    BehaviourEventAction.SourceAdd,
    BehaviourEventMedium.Button,
    getScreenNameFromPage(),
    MetricsEventSpace.Modal,
  );

  // If project is uninitialized, initialize an empty project
  const projectInitialized = await isProjectInitialized(instanceId);
  if (!projectInitialized) {
    await runtimeServiceUnpackEmpty(instanceId, {
      displayName: EMPTY_PROJECT_TITLE,
    });

    // Race condition: invalidate("init") must be called before we navigate to
    // `/files/${newFilePath}`. invalidate("init") is also called in the
    // `WatchFilesClient`, but there it's not guaranteed to get invoked before we need it.
    await invalidate("init");
  }

  // Convert the form values to Source YAML
  // TODO: Quite a few adhoc code is being added. We should revisit the way we generate the yaml.
  const formValues = Object.fromEntries(
    Object.entries(values).map(([key, value]) => {
      switch (key) {
        case "project_id":
        case "account":
        case "output_location":
        case "workgroup":
        case "database_url":
          return [key, value];
        default:
          return [fromYupFriendlyKey(key), value];
      }
    }),
  );

  /**
   * Sources / Models
   */

  if (formType === "source") {
    switch (olapDriver) {
      case "duckdb": {
        const [rewrittenConnector, rewrittenFormValues] = maybeRewriteToDuckDb(
          connector,
          formValues,
        );

        // Make a new <source>.yaml file
        const newSourceFilePath = getFileAPIPathFromNameAndType(
          values.name as string,
          EntityType.Source,
        );
        await runtimeServicePutFile(instanceId, {
          path: newSourceFilePath,
          blob: compileSourceYAML(rewrittenConnector, rewrittenFormValues),
          create: true,
          createOnly: false, // The modal might be opened from a YAML file with placeholder text, so the file might already exist
        });

        // Update the `.env` file
        await runtimeServicePutFile(instanceId, {
          path: ".env",
          blob: await updateDotEnvWithSecrets(
            queryClient,
            rewrittenConnector,
            rewrittenFormValues,
          ),
          create: true,
          createOnly: false,
        });

        await goto(`/files/${newSourceFilePath}`);

        // Return the path to the new source file
        return newSourceFilePath;
      }
      case "clickhouse": {
        const newModelFilePath = getFileAPIPathFromNameAndType(
          values.name as string,
          EntityType.Model,
        );

        // Make a new <model>.yaml file
        await runtimeServicePutFile(instanceId, {
          path: newModelFilePath,
          blob: compileClickhouseSourceConnectorFile(connector, formValues),
          create: true,
          createOnly: false,
        });

        // Update the `.env` file
        await runtimeServicePutFile(instanceId, {
          path: ".env",
          blob: await updateDotEnvWithSecrets(
            queryClient,
            connector,
            formValues,
          ),
          create: true,
          createOnly: false,
        });

        // Return the path to the new model file
        return newModelFilePath;
      }
      default:
        throw new Error(`Unsupported OLAP driver: ${olapDriver}`);
    }
  }

  /**
   * Connectors
   */

  // Determine the name of the new connector file
  const isOnboardingFlow = get(page).url.pathname.includes("/welcome");
  const connectorDriverName = connector.name as string;
  const newConnectorName = isOnboardingFlow
    ? connectorDriverName
    : getName(
        connectorDriverName,
        fileArtifacts.getNamesForKind(ResourceKind.Connector),
      );

  // Make a new `<connector>.yaml` file
  const newConnectorFilePath = getFileAPIPathFromNameAndType(
    newConnectorName,
    EntityType.Connector,
  );
  await runtimeServicePutFile(instanceId, {
    path: newConnectorFilePath,
    blob: compileConnectorYAML(connector, formValues),
    create: true,
    createOnly: false,
  });

  // Update the `.env` file
  await runtimeServicePutFile(instanceId, {
    path: ".env",
    blob: await updateDotEnvWithSecrets(queryClient, connector, formValues),
    create: true,
    createOnly: false,
  });

  // Update the `rill.yaml` file
  await runtimeServicePutFile(instanceId, {
    path: "rill.yaml",
    blob: await updateRillYAMLWithOlapConnector(queryClient, newConnectorName),
    create: true,
    createOnly: false,
  });

  // Check for a reconcile error
  const fileArtifact = fileArtifacts.getFileArtifact(newConnectorFilePath);
  if (fileArtifact.reconciling) {
    await waitUntil(() => !fileArtifact.reconciling, 500);
  }
  const hasErrorsStore = fileArtifact.getHasErrors(queryClient, instanceId);
  if (get(hasErrorsStore)) {
    throw fileArtifact.getAllErrors(queryClient, instanceId);
  }

  // Test the connection by calling `GetTables`
  const queryKey = getConnectorServiceOLAPListTablesQueryKey({
    instanceId,
    connector: newConnectorName,
  });
  const queryFn = () =>
    connectorServiceOLAPListTables({
      instanceId,
      connector: newConnectorName,
    });
  await queryClient.fetchQuery({ queryKey, queryFn });

  // Return the path to the new connector file
  return newConnectorFilePath;
}
