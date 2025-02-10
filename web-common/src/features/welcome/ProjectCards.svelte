<script lang="ts">
  import { goto } from "$app/navigation";
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import { get } from "svelte/store";
  import Card from "../../components/card/Card.svelte";
  import CardDescription from "../../components/card/CardDescription.svelte";
  import CardTitle from "../../components/card/CardTitle.svelte";
  import { queryClient } from "../../lib/svelte-query/globalQueryClient";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../metrics/service/MetricsTypes";
  import {
    createRuntimeServiceUnpackExample,
    getRuntimeServiceListFilesQueryKey,
    runtimeServiceListFiles,
    type V1ListFilesResponse,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { EXAMPLES } from "./constants";

  const unpackExampleProject = createRuntimeServiceUnpackExample();

  let selectedProjectName: string | null = null;

  $: ({ instanceId } = $runtime);

  $: ({ mutateAsync: unpackExample } = $unpackExampleProject);

  async function unpackProjectAndGoToDashboard(
    example: (typeof EXAMPLES)[number],
  ) {
    selectedProjectName = example.name;

    await behaviourEvent?.fireSplashEvent(
      BehaviourEventAction.ExampleAdd,
      BehaviourEventMedium.Card,
      MetricsEventSpace.Workspace,
      selectedProjectName,
    );

    try {
      // Unpack the example project
      await unpackExample({
        instanceId,
        data: {
          name: selectedProjectName,
          force: true,
        },
      });

      // Get the first dashboard file, and navigate to it
      const files = await queryClient.fetchQuery<V1ListFilesResponse>({
        queryKey: getRuntimeServiceListFilesQueryKey(
          get(runtime).instanceId,
          undefined,
        ),
        queryFn: ({ signal }) => {
          return runtimeServiceListFiles(
            get(runtime).instanceId,
            undefined,
            signal,
          );
        },
      });
      const firstDashboardFile = files.files?.find((file) =>
        file.path?.startsWith("/dashboards/"),
      );
      await goto(`/files${firstDashboardFile?.path}`);
    } catch {
      selectedProjectName = null;
    }
  }
</script>

<section class="flex flex-col items-center gap-y-5">
  <Subheading>Or jump right into a project.</Subheading>
  <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
    {#each EXAMPLES as example (example.name)}
      <Card
        imageUrl={example.image}
        disabled={!!selectedProjectName}
        isLoading={selectedProjectName === example.name}
        on:click={async () => {
          await unpackProjectAndGoToDashboard(example);
        }}
      >
        <CardTitle>{example.title}</CardTitle>
        <CardDescription>{example.description}</CardDescription>
      </Card>
    {/each}
  </div>
</section>
