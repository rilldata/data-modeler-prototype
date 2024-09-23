<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceEditAlert } from "@rilldata/web-admin/client";
  import {
    extractAlertFormValues,
    extractAlertNotification,
  } from "@rilldata/web-common/features/alerts/extract-alert-form-values";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    V1AlertSpec,
    V1MetricsViewAggregationRequest,
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import BaseAlertForm from "./BaseAlertForm.svelte";
  import { getSnoozeValueFromAlertSpec } from "./delivery-tab/snooze";
  import {
    alertFormValidationSchema,
    AlertFormValues,
    getAlertQueryArgsFromFormValues,
  } from "./form-utils";

  export let alertSpec: V1AlertSpec;
  export let metricsViewName: string;
  export let defaultTimeRange: string | undefined;

  const editAlert = createAdminServiceEditAlert();
  const dispatch = createEventDispatcher();

  $: ({ instanceId } = $runtime);

  $: ({ project, alert: alertName, organization } = $page.params);
  const queryArgsJson = JSON.parse(
    (alertSpec.resolverProperties?.query_args_json ??
      alertSpec.queryArgsJson) as string,
  ) as V1MetricsViewAggregationRequest;

  $: timeRange = useMetricsViewTimeRange(instanceId, metricsViewName, {
    query: { queryClient },
  });

  const formState = createForm<AlertFormValues>({
    initialValues: {
      name: alertSpec.title as string,
      snooze: getSnoozeValueFromAlertSpec(alertSpec),
      evaluationInterval: alertSpec.intervalsIsoDuration ?? "",
      ...extractAlertNotification(alertSpec),
      ...extractAlertFormValues(
        queryArgsJson,
        defaultTimeRange,
        $timeRange?.data ?? {},
      ),
    },
    validationSchema: alertFormValidationSchema,
    onSubmit: async (values) => {
      try {
        await $editAlert.mutateAsync({
          organization,
          project,
          name: alertName,
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
        void queryClient.invalidateQueries(
          getRuntimeServiceGetResourceQueryKey($runtime.instanceId, {
            "name.name": alertName,
            "name.kind": ResourceKind.Alert,
          }),
        );
        void queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
        );
        dispatch("close");
        eventBus.emit("notification", {
          message: "Alert edited",
          type: "success",
        });
      } catch (e) {
        // showing error below
      }
    },
  });
  const { form } = formState;
  $: if ($timeRange?.data) {
    const formValues = extractAlertFormValues(
      queryArgsJson,
      defaultTimeRange,
      $timeRange.data,
    );
    for (const fk in formValues) {
      $form[fk] = formValues[fk];
    }
  }
</script>

<BaseAlertForm {formState} isEditForm={true} on:cancel on:close />
