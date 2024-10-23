<script lang="ts">
  import { goto } from "$app/navigation";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import type { V1Resource } from "../../../runtime-client";
  import type { HTTPError } from "../../../runtime-client/fetchWrapper";
  import { extractFileName } from "@rilldata/utils";
  import { featureFlags } from "../../feature-flags";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";

  const { ai } = featureFlags;

  export let sourcePath: string | null;

  let fileArtifact: FileArtifact;
  let sourceQuery: CreateQueryResult<V1Resource, HTTPError>;

  $: sourceName = extractFileName(sourcePath ?? "");

  $: runtimeInstanceId = $runtime.instanceId;

  $: if (sourcePath) {
    fileArtifact = fileArtifacts.getFileArtifact(sourcePath);
    sourceQuery = fileArtifact.getResource(queryClient, runtimeInstanceId);
  }
  $: sinkConnector = $sourceQuery?.data?.source?.spec?.sinkConnector;

  $: createExploreFromTable =
    sourcePath !== null
      ? useCreateMetricsViewFromTableUIAction(
          $runtime.instanceId,
          sinkConnector as string,
          "",
          "",
          sourceName,
          true,
          BehaviourEventMedium.Button,
          MetricsEventSpace.Modal,
        )
      : null;

  function close() {
    sourceImportedPath.set(null);
  }

  async function goToSource() {
    await goto(`/files${$sourceImportedPath ?? ""}`);
    close();
  }

  async function generateMetrics() {
    // This should never happen, because the button is
    // disabled when this is null, but adding this check
    // for type narrowing and just in case.
    if (createExploreFromTable === null) return;
    close();
    await createExploreFromTable();
  }
</script>

<AlertDialog.Root open={sourcePath !== null}>
  <AlertDialog.Content>
    <AlertDialog.Title>Source imported successfully</AlertDialog.Title>

    <AlertDialog.Description>
      <span class="font-mono text-slate-800 break-all">{sourceName}</span> has been
      ingested. What would you like to do next?
    </AlertDialog.Description>

    <AlertDialog.Footer>
      <AlertDialog.Action asChild let:builder>
        <AlertDialog.Cancel asChild let:builder>
          <Button builders={[builder]} on:click={goToSource} type="secondary">
            View this source
          </Button>
        </AlertDialog.Cancel>

        <Button
          builders={[builder]}
          disabled={createExploreFromTable === null}
          on:click={generateMetrics}
          type="primary"
        >
          Generate dashboard

          {#if $ai}
            with AI
            <WandIcon class="w-3 h-3" />
          {/if}
        </Button>
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
