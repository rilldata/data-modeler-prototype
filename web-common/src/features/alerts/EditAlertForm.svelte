<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceEditAlert } from "@rilldata/web-admin/client";
  import { getExploreName } from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
  import {
    extractAlertFormValues,
    extractAlertNotification,
  } from "@rilldata/web-common/features/alerts/extract-alert-form-values";
  import {
    useMetricsViewTimeRange,
    useMetricsViewValidSpec,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { eventBus } from "@rilldata/events";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import {
    type V1AlertSpec,
    type V1MetricsViewAggregationRequest,
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import BaseAlertForm from "./BaseAlertForm.svelte";
  import { getSnoozeValueFromAlertSpec } from "./delivery-tab/snooze";
  import {
    alertFormValidationSchema,
    type AlertFormValues,
    getAlertQueryArgsFromFormValues,
  } from "./form-utils";

  export let alertSpec: V1AlertSpec;
  export let metricsViewName: string;

  const editAlert = createAdminServiceEditAlert();
  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alertName = $page.params.alert;
  const queryArgsJson = JSON.parse(
    (alertSpec.resolverProperties?.query_args_json ??
      alertSpec.queryArgsJson) as string,
  ) as V1MetricsViewAggregationRequest;

  $: metricsViewSpec = useMetricsViewValidSpec(
    $runtime?.instanceId,
    metricsViewName,
  );
  $: timeRange = useMetricsViewTimeRange(
    $runtime?.instanceId,
    metricsViewName,
    { query: { queryClient } },
  );

  $: exploreName = getExploreName(alertSpec.annotations?.web_open_path ?? "");

  const formState = createForm<AlertFormValues>({
    initialValues: {
      name: alertSpec.title as string,
      exploreName: exploreName ?? metricsViewName,
      snooze: getSnoozeValueFromAlertSpec(alertSpec),
      evaluationInterval: alertSpec.intervalsIsoDuration ?? "",
      ...extractAlertNotification(alertSpec),
      ...extractAlertFormValues(
        queryArgsJson,
        $metricsViewSpec?.data ?? {},
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
              webOpenPath: exploreName ? `/explore/${exploreName}` : undefined,
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
      } catch {
        // showing error below
      }
    },
  });
  const { form } = formState;
  $: if ($metricsViewSpec?.data && $timeRange?.data) {
    const formValues = extractAlertFormValues(
      queryArgsJson,
      $metricsViewSpec.data,
      $timeRange.data,
    );
    for (const fk in formValues) {
      $form[fk] = formValues[fk];
    }
  }
</script>

<BaseAlertForm {formState} isEditForm={true} on:cancel on:close />
