import { goto } from "$app/navigation";
import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
import type { QueryClient } from "@tanstack/query-core";
import { get } from "svelte/store";
import { behaviourEvent } from "../../../metrics/initMetrics";
import {
  BehaviourEventAction,
  BehaviourEventMedium,
} from "../../../metrics/service/BehaviourEventTypes";
import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
import {
  type V1ConnectorDriver,
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import {
  compileConnectorYAML,
  updateDotEnvWithSecrets,
  updateRillYAMLWithOlapConnector,
} from "../../connectors/code-utils";
import { getFileAPIPathFromNameAndType } from "../../entity-management/entity-mappers";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import { getName } from "@rilldata/utils";
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
): Promise<void> {
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
      title: EMPTY_PROJECT_TITLE,
    });
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
   * Sources
   */

  if (formType === "source") {
    const [rewrittenConnector, rewrittenFormValues] = maybeRewriteToDuckDb(
      connector,
      formValues,
    );

    // Make a new <source>.yaml file
    const newSourceFilePath = getFileAPIPathFromNameAndType(
      values.name as string,
      EntityType.Table,
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

    await checkSourceImported(queryClient, newSourceFilePath);

    await goto(`/files/${newSourceFilePath}`);

    return;
  }

  /**
   * Connectors
   */

  const newConnectorName = getName(
    connector.name as string,
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

  // Go to the new connector file
  await goto(`/files/${newConnectorFilePath}`);
}
