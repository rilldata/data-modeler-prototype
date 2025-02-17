<script lang="ts">
  import { createAdminServiceTriggerReport } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { getRuntimeServiceGetResourceQueryKey } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { useReport } from "../selectors";

  export let organization: string;
  export let project: string;
  export let report: string;

  $: ({ instanceId } = $runtime);

  const queryClient = useQueryClient();
  const triggerReport = createAdminServiceTriggerReport();
  const reportQuery = useReport(instanceId, report);

  async function handleRunNow() {
    const lastExecution =
      $reportQuery.data?.resource.report.state.executionHistory[0];
    await $triggerReport.mutateAsync({
      organization,
      project,
      name: report,
      data: undefined,
    });

    eventBus.emit("notification", {
      message: "Triggered an ad-hoc run of this report.",
      type: "success",
    });

    // Refetch the resource query until the new report run shows up in the recent history table
    while (
      !$reportQuery.data ||
      $reportQuery.data.resource.report.state.executionHistory[0] ===
        lastExecution
    ) {
      await queryClient.invalidateQueries(
        getRuntimeServiceGetResourceQueryKey(instanceId, {
          "name.name": report,
          "name.kind": ResourceKind.Report,
        }),
      );
      await new Promise((resolve) => setTimeout(resolve, 1000));
    }
  }
</script>

<Tooltip distance={8}>
  <Button
    type="primary"
    on:click={handleRunNow}
    disabled={$triggerReport.isLoading}
  >
    Run now
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="300px">
    Run this report immediately. A new report will be generated and emailed to
    recipients.
  </TooltipContent>
</Tooltip>
