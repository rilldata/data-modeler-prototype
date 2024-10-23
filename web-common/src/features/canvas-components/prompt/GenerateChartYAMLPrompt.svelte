<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import ChartPromptHistoryDisplay from "@rilldata/web-common/features/canvas-components/prompt/ChartPromptHistoryDisplay.svelte";
  import { createFullChartGenerator } from "@rilldata/web-common/features/canvas-components/prompt/generateChart";
  import { useAllFileNames } from "@rilldata/web-common/features/entity-management/file-selectors";
  import { getName } from "@rilldata/utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open: boolean;
  export let table: string = "";
  export let connector: string = "";
  export let metricsView: string = "";

  const queryClient = useQueryClient();

  let prompt: string;

  $: chartFileNames = useAllFileNames(queryClient, $runtime.instanceId);
  $: generateVegaConfig = createFullChartGenerator($runtime.instanceId);

  async function createVegaConfig() {
    open = false;
    await generateVegaConfig(
      prompt,
      {
        table,
        connector,
        metricsView,
      },
      getName(`${table || metricsView}_chart`, $chartFileNames.data ?? []),
    );
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Title>
      Generate component yaml for "{table || metricsView}" using AI</Dialog.Title
    >

    <Input bind:value={prompt} label="Prompt" />
    <ChartPromptHistoryDisplay
      entityName={table || metricsView}
      on:reuse-prompt={({ detail }) => {
        prompt = detail;
      }}
    />
    <div class="ml-auto">
      <Button on:click={createVegaConfig} large type="primary">Generate</Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
