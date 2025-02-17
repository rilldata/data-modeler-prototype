<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { useProjectDeployment } from "./selectors";

  const queryClient = useQueryClient();

  export let organization: string;
  export let project: string;

  $: ({ instanceId } = $runtime);

  $: projectDeployment = useProjectDeployment(organization, project);
  $: ({ data: deployment } = $projectDeployment);
  $: isDeploymentNotOk =
    deployment?.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_OK;

  $: hasResourceErrorsQuery = createRuntimeServiceListResources(
    instanceId,
    undefined,
    {
      query: {
        select: (data) => {
          return (
            data.resources.filter((resource) => !!resource.meta.reconcileError)
              .length > 0
          );
        },
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );
  $: ({
    data: hasResourceErrors,
    error: hasResourceErrorsError,
    isLoading: hasResourceErrorsLoading,
  } = $hasResourceErrorsQuery);

  $: projectParserQuery = useProjectParser(queryClient, instanceId, {
    refetchOnMount: true,
    refetchOnWindowFocus: true,
  });
  $: ({
    data: projectParserData,
    error: projectParserError,
    isLoading: projectParserLoading,
  } = $projectParserQuery);
  $: hasParseErrors =
    projectParserData?.projectParser.state.parseErrors.length > 0;
</script>

{#if hasResourceErrorsLoading || projectParserLoading}
  <LoadingSpinner />
{:else if hasResourceErrorsError || projectParserError}
  <CancelCircle className="text-red-600" />
{:else if isDeploymentNotOk || hasResourceErrors || hasParseErrors}
  <CancelCircle className="text-red-600" />
{:else}
  <CheckCircle className="text-green-400" />
{/if}
