<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateReport,
    createAdminServiceGetCurrentUser,
    createAdminServiceEditReport,
  } from "@rilldata/web-admin/client";
  import { getDashboardNameFromReport } from "@rilldata/web-common/features/scheduled-reports/utils";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../components/button";
  import { eventBus } from "@rilldata/events";
  import { getLocalIANA } from "../../lib/time/timezone";
  import {
    V1ExportFormat,
    getRuntimeServiceListResourcesQueryKey,
    type V1ReportSpec,
    getRuntimeServiceGetResourceQueryKey,
    type V1ReportSpecAnnotations,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import BaseScheduledReportForm from "./BaseScheduledReportForm.svelte";
  import {
    convertFormValuesToCronExpression,
    getNextQuarterHour,
    getTimeIn24FormatFromDateTime,
    getTodaysDayOfWeek,
    getDayOfWeekFromCronExpression,
    getFrequencyFromCronExpression,
    getTimeOfDayFromCronExpression,
  } from "./time-utils";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { ResourceKind } from "../entity-management/resource-selectors";

  export let open: boolean;
  export let queryArgs: any | undefined = undefined;
  export let metricsViewProto: string | undefined = undefined;
  export let exploreName: string | undefined = undefined;
  export let reportSpec: V1ReportSpec | undefined = undefined;

  const user = createAdminServiceGetCurrentUser();

  $: if (!exploreName) {
    exploreName =
      getDashboardNameFromReport(reportSpec) ?? queryArgs.metricsViewName;
  }

  $: ({ organization, project, report: reportName } = $page.params);

  $: mutation = reportSpec
    ? createAdminServiceEditReport()
    : createAdminServiceCreateReport();

  const formState = createForm({
    initialValues: getInitialValues(reportSpec),
    validationSchema: yup.object({
      title: yup.string().required("Required"),
      recipients: yup.array().of(
        yup.object().shape({
          email: yup.string().email("Invalid email"),
        }),
      ),
    }),
    onSubmit: async (values) => {
      const refreshCron = convertFormValuesToCronExpression(
        values.frequency,
        values.dayOfWeek,
        values.timeOfDay,
      );

      try {
        await $mutation.mutateAsync({
          organization,
          project,
          name: reportName,
          data: {
            options: {
              title: values.title,
              refreshCron: refreshCron, // for testing: "* * * * *"
              refreshTimeZone: values.timeZone,
              queryName: reportSpec?.queryName ?? "MetricsViewAggregation",
              queryArgsJson: JSON.stringify(
                reportSpec?.queryArgsJson
                  ? JSON.parse(reportSpec.queryArgsJson)
                  : queryArgs,
              ),
              exportLimit: values.exportLimit || undefined,
              exportFormat: values.exportFormat,
              emailRecipients: values.recipients
                .map((r) => r.email)
                .filter(Boolean),
              webOpenState: reportSpec
                ? (reportSpec.annotations as V1ReportSpecAnnotations)[
                    "web_open_state"
                  ]
                : metricsViewProto,
              webOpenPath: exploreName ? `/explore/${exploreName}` : undefined,
            },
          },
        });

        if (reportSpec) {
          await queryClient.invalidateQueries(
            getRuntimeServiceGetResourceQueryKey($runtime.instanceId, {
              "name.name": reportName,
              "name.kind": ResourceKind.Report,
            }),
          );
        }

        await queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
        );

        open = false;

        eventBus.emit("notification", {
          message: `Report ${reportSpec ? "edited" : "created"}`,
          link: reportSpec
            ? undefined
            : {
                href: `/${organization}/${project}/-/reports`,
                text: "Go to scheduled reports",
              },
          type: "success",
        });
      } catch {
        // showing error below
      }
    },
  });

  const { isSubmitting, form } = formState;

  function getInitialValues(reportSpec: V1ReportSpec | undefined) {
    return {
      title: reportSpec?.title ?? "",
      frequency: reportSpec
        ? getFrequencyFromCronExpression(
            reportSpec.refreshSchedule?.cron as string,
          )
        : "Weekly",
      dayOfWeek: reportSpec
        ? getDayOfWeekFromCronExpression(
            reportSpec.refreshSchedule?.cron as string,
          )
        : getTodaysDayOfWeek(),
      timeOfDay: reportSpec
        ? getTimeOfDayFromCronExpression(
            reportSpec.refreshSchedule?.cron as string,
          )
        : getTimeIn24FormatFromDateTime(getNextQuarterHour()),
      timeZone: reportSpec?.refreshSchedule?.timeZone ?? getLocalIANA(),
      exportFormat: reportSpec
        ? (reportSpec?.exportFormat ?? V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED)
        : V1ExportFormat.EXPORT_FORMAT_CSV,
      exportLimit: reportSpec
        ? reportSpec.exportLimit === "0"
          ? ""
          : reportSpec.exportLimit
        : "",
      recipients: reportSpec?.notifiers
        ?.find((n) => n.connector === "email")
        ?.properties?.recipients?.map((email) => ({
          email: email,
        })) ?? [
        { email: $user.data?.user?.email ? $user.data.user.email : "" },
        { email: "" },
      ],
    };
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Title>Schedule report</Dialog.Title>

    <BaseScheduledReportForm
      formId="scheduled-report-form"
      {formState}
      exploreName={exploreName ?? ""}
    />

    <div class="flex items-center gap-x-2 mt-5">
      {#if $mutation.isError}
        <div class="text-red-500">{$mutation.error.message}</div>
      {/if}
      <div class="grow" />
      <Button on:click={() => (open = false)} type="secondary">Cancel</Button>
      <Button
        disabled={$isSubmitting ||
          $form["recipients"].filter((r) => r.email).length === 0}
        form="scheduled-report-form"
        submitForm
        type="primary"
      >
        {reportSpec ? "Save" : "Create"}
      </Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
