import { goto } from "$app/navigation";
import { get } from "svelte/store";
import { notifications } from "../../../components/notifications";
import { appScreen } from "../../../layout/app-store";
import { overlay } from "../../../layout/overlay-store";
import { behaviourEvent } from "../../../metrics/initMetrics";
import type { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  type MetricsEventSpace,
} from "../../../metrics/service/MetricsTypes";
import {
  createRuntimeServiceGenerateMetricsViewFile,
  runtimeServiceGetFile,
} from "../../../runtime-client";
import { useDashboardFileNames } from "../../dashboards/selectors";
import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
import { getName } from "../../entity-management/name-utils";
import { EntityType } from "../../entity-management/types";
import CancelGeneration from "./CancelGeneration.svelte";

/**
 * Wrapper function that takes care of UI side effects on top of creating a dashboard from a table.
 */
export function useCreateDashboardFromTableUIAction(
  instanceId: string,
  tableName: string,
  behaviourEventMedium: BehaviourEventMedium,
  metricsEventSpace: MetricsEventSpace,
  toggleContextMenu: () => void = () => {},
) {
  const dashboardNames = useDashboardFileNames(instanceId);

  const generateMetricsViewFileMutation =
    createRuntimeServiceGenerateMetricsViewFile();

  let isAICancelled = false;

  // Return a function that can be called to create a dashboard from a table
  return async () => {
    overlay.set({
      title: "Hang tight! AI is personalizing your dashboard",
      component: CancelGeneration,
      componentProps: {
        onCancel: () => {
          isAICancelled = true;
        },
      },
    });

    toggleContextMenu(); // TODO: see if we can bring this out of this function

    const newDashboardName = getName(
      `${tableName}_dashboard`,
      get(dashboardNames).data ?? [],
    );

    try {
      console.log("Using AI to generate dashboard for " + tableName);
      const newFilePath = getFilePathFromNameAndType(
        newDashboardName,
        EntityType.MetricsDefinition,
      );

      void get(generateMetricsViewFileMutation).mutateAsync({
        instanceId,
        data: {
          table: tableName,
          path: newFilePath,
          useAi: true,
        },
      });

      console.log("Waiting for AI...");
      // Poll until the AI generation is complete or canceled
      while (!isAICancelled) {
        // Wait 1 second
        await new Promise((resolve) => setTimeout(resolve, 1000));

        try {
          await runtimeServiceGetFile(instanceId, newFilePath);
          // AI is done
          break;
        } catch (err) {
          // AI is not done
        }
      }

      // If canceled, then submit another with AI=false
      if (isAICancelled) {
        console.log("AI was canceled");
        await get(generateMetricsViewFileMutation).mutateAsync({
          instanceId,
          data: {
            table: tableName,
            path: getFilePathFromNameAndType(
              newDashboardName,
              EntityType.MetricsDefinition,
            ),
            useAi: false,
          },
        });
      }

      await goto(`/dashboard/${newDashboardName}`);
      void behaviourEvent.fireNavigationEvent(
        newDashboardName,
        behaviourEventMedium,
        metricsEventSpace,
        get(appScreen)?.type,
        MetricsEventScreenName.Dashboard,
      );
    } catch (err) {
      notifications.send({
        message: "Failed to create a dashboard for " + tableName,
        detail: err.response?.data?.message ?? err.message,
      });
    }
    overlay.set(null);
  };
}
