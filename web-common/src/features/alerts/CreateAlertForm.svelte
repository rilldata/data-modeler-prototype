<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateAlert,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import { getHasSlackConnection } from "@rilldata/web-common/features/alerts/delivery-tab/notifiers-utils";
  import { SnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze";
  import {
    AlertFormValues,
    alertFormValidationSchema,
    getAlertQueryArgsFromFormValues,
  } from "@rilldata/web-common/features/alerts/form-utils";
  import { getEmptyMeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    mapComparisonTimeRange,
    mapTimeRange,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
  import {
    V1Operation,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { get } from "svelte/store";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { runtime } from "../../runtime-client/runtime-store";
  import BaseAlertForm from "./BaseAlertForm.svelte";
  import { useMetricsView } from "../dashboards/selectors";

  const user = createAdminServiceGetCurrentUser();
  const createAlert = createAdminServiceCreateAlert();

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  const {
    metricsViewName,
    dashboardStore,

    selectors: {
      timeRangeSelectors: { timeControlsState },
    },
  } = getStateManagers();
  const timeControls = get(timeControlsState);

  $: ({ organization, project } = $page.params);

  $: metricsViewQuery = useMetricsView($runtime.instanceId, $metricsViewName);

  $: metricsView = $metricsViewQuery.data;

  // Set defaults depending on UI state
  // if in TDD take active measure and comparison dimension
  // If expanded leaderboard, take first dimension and active dimensions
  let dimension = "";
  if ($dashboardStore.tdd.expandedMeasureName) {
    dimension = $dashboardStore.selectedComparisonDimension ?? "";
  } else {
    dimension = $dashboardStore.selectedDimensionName ?? "";
  }

  // TODO: get metrics view spec
  $: timeRange = mapTimeRange(timeControls, metricsView?.defaultTimeRange);
  $: comparisonTimeRange = mapComparisonTimeRange(
    $dashboardStore,
    timeControls,
    timeRange,
  );

  $: formState = createForm<AlertFormValues>({
    initialValues: {
      name: "",
      measure:
        $dashboardStore.tdd.expandedMeasureName ??
        $dashboardStore.leaderboardMeasureName ??
        "",
      splitByDimension: dimension,
      evaluationInterval: "",
      criteria: [
        {
          ...getEmptyMeasureFilterEntry(),
          measure: $dashboardStore.leaderboardMeasureName ?? "",
        },
      ],
      criteriaOperation: V1Operation.OPERATION_AND,
      snooze: SnoozeOptions[0].value, // Defaults to `Off`
      enableSlackNotification: false,
      slackChannels: [
        {
          channel: "",
        },
      ],
      slackUsers: [
        { email: $user.data?.user?.email ? $user.data.user.email : "" },
        { email: "" },
      ],
      enableEmailNotification: true,
      emailRecipients: [
        { email: $user.data?.user?.email ? $user.data.user.email : "" },
        { email: "" },
      ],
      // The remaining fields are not editable in the form, but it's helpful to have access to them throughout the alert dialog
      // Also, in the future, they might even be editable.
      metricsViewName: $metricsViewName,
      whereFilter: $dashboardStore.whereFilter,
      dimensionThresholdFilters: $dashboardStore.dimensionThresholdFilters,
      timeRange: timeRange
        ? {
            ...timeRange,
            end: timeControls.timeEnd,
          }
        : undefined,
      comparisonTimeRange: comparisonTimeRange
        ? {
            ...comparisonTimeRange,
            end: timeControls.timeEnd,
          }
        : undefined,
    } as AlertFormValues,
    validationSchema: alertFormValidationSchema,
    onSubmit: async (values) => {
      try {
        await $createAlert.mutateAsync({
          organization,
          project,
          data: {
            options: {
              title: values.name,
              queryName: "MetricsViewAggregation",
              queryArgsJson: JSON.stringify(
                getAlertQueryArgsFromFormValues(values),
              ),
              metricsViewName: values.metricsViewName,
              slackChannels: values.enableSlackNotification
                ? values.slackChannels.map((c) => c.channel).filter(Boolean)
                : undefined,
              slackUsers: values.enableSlackNotification
                ? values.slackUsers.map((c) => c.email).filter(Boolean)
                : undefined,
              emailRecipients: values.enableEmailNotification
                ? values.emailRecipients.map((r) => r.email).filter(Boolean)
                : undefined,
              renotify: !!values.snooze,
              renotifyAfterSeconds: values.snooze ? Number(values.snooze) : 0,
            },
          },
        });
        queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
        );
        dispatch("close");
        eventBus.emit("notification", {
          message: "Alert created",
          link: {
            href: `/${organization}/${project}/-/alerts`,
            text: "Go to alerts",
          },
        });
      } catch (e) {
        // showing error below
      }
    },
  });

  $: form = formState.form;
  $: hasSlackNotifier = getHasSlackConnection($runtime.instanceId);
  $: if ($hasSlackNotifier.data) {
    $form["enableSlackNotification"] = true;
  }

  $: if (timeControls.timeEnd) {
    $form["timeRange"].end = timeControls.timeEnd;
  }
  $: if (timeControls.comparisonTimeEnd && $form["comparisonTimeRange"]) {
    $form["comparisonTimeRange"].end = timeControls.comparisonTimeEnd;
  }
</script>

<BaseAlertForm {formState} isEditForm={false} on:cancel on:close />
