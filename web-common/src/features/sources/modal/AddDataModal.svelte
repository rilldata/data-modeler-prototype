<script lang="ts">
  import { goto } from "$app/navigation";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import { type V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { logoIconMapping } from "../../connectors/connector-icon-mapping";
  import {
    useCurrentOlapConnector,
    useSourceConnectorsForCurrentOlapConnector,
  } from "../../connectors/olap/selectors";
  import { duplicateSourceName } from "../sources-store";
  import AddDataForm from "./AddDataForm.svelte";
  import DuplicateSource from "./DuplicateSource.svelte";
  import LocalSourceUpload from "./LocalSourceUpload.svelte";
  import RequestConnectorForm from "./RequestConnectorForm.svelte";

  let step = 0;
  let selectedConnector: null | V1ConnectorDriver = null;
  let requestConnector = false;

  $: olapConnector = useCurrentOlapConnector($runtime.instanceId);
  $: olapConnectorType = $olapConnector.data?.type;
  $: sourceConnectors = useSourceConnectorsForCurrentOlapConnector(
    $runtime.instanceId,
  );

  onMount(() => {
    function listen(e: PopStateEvent) {
      step = e.state?.step ?? 0;
      requestConnector = e.state?.requestConnector ?? false;
      selectedConnector = e.state?.selectedConnector ?? null;
    }
    window.addEventListener("popstate", listen);

    return () => {
      window.removeEventListener("popstate", listen);
    };
  });

  function goToConnectorForm(connector: V1ConnectorDriver) {
    const state = {
      step: 2,
      selectedConnector: connector,
      requestConnector: false,
    };
    window.history.pushState(state, "", "");
    dispatchEvent(new PopStateEvent("popstate", { state }));
  }

  function goToRequestConnector() {
    const state = { step: 2, requestConnector: true };
    window.history.pushState(state, "", "");
    dispatchEvent(new PopStateEvent("popstate", { state }));
  }

  function back() {
    window.history.back();
  }

  function onSubmit(newFilePath: string) {
    void goto(`/files/${newFilePath}`);
    resetModal();
  }

  function resetModal() {
    const state = { step: 0, selectedConnector: null, requestConnector: false };
    window.history.pushState(state, "", "");
    dispatchEvent(new PopStateEvent("popstate", { state: state }));
  }

  async function onCancelDialog() {
    await behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceCancel,
      BehaviourEventMedium.Button,
      getScreenNameFromPage(),
      MetricsEventSpace.Modal,
    );

    resetModal();
  }
</script>

{#if step >= 1 || $duplicateSourceName}
  <Dialog.Root
    open
    onOpenChange={async (open) => {
      if (!open) {
        await onCancelDialog();
      }
    }}
  >
    <Dialog.Content noClose>
      <section class="mb-1">
        <Dialog.Title>
          {#if $duplicateSourceName !== null}
            Duplicate source
          {:else if selectedConnector}
            {selectedConnector?.displayName}
          {:else if requestConnector}
            Request a connector
          {:else if step === 1}
            Add a source
          {/if}
        </Dialog.Title>

        {#if $duplicateSourceName}
          <DuplicateSource onCancel={resetModal} onComplete={resetModal} />
        {:else if requestConnector}
          <RequestConnectorForm on:close={resetModal} on:back={back} />
        {:else if step === 1 && olapConnectorType}
          {#if $sourceConnectors.length === 0}
            <div class="text-slate-500">
              No connectors available for your current OLAP engine.
            </div>
          {:else}
            <div class="connector-grid">
              {#each $sourceConnectors as connector (connector.name)}
                {#if connector.name}
                  <button
                    id={connector.name}
                    on:click={() => goToConnectorForm(connector)}
                    class="connector-tile-button"
                  >
                    <div class="connector-wrapper">
                      <svelte:component
                        this={logoIconMapping[connector.name]}
                      />
                    </div>
                  </button>
                {/if}
              {/each}
            </div>
          {/if}
        {:else if step === 2 && olapConnectorType && selectedConnector}
          {#if selectedConnector.name === "local_file"}
            <LocalSourceUpload on:close={resetModal} on:back={back} />
          {:else if selectedConnector && selectedConnector.name}
            <AddDataForm
              connector={selectedConnector}
              formType="source"
              olapDriver={olapConnectorType}
              {onSubmit}
            >
              <svelte:fragment slot="actions" let:submitting>
                <div class="flex items-center gap-x-2 ml-auto">
                  <Button on:click={back} type="secondary">Back</Button>
                  <Button
                    disabled={submitting}
                    form="add-data-form"
                    submitForm
                    type="primary"
                  >
                    Add data
                  </Button>
                </div>
              </svelte:fragment>
            </AddDataForm>
          {/if}
        {/if}
      </section>

      {#if step === 1}
        <!-- <section>
          <Dialog.Title>Connect an OLAP engine</Dialog.Title>

          <div class="connector-grid">
            {#each connectors?.filter((c) => c.name && OLAP_CONNECTORS.includes(c.name)) as connector (connector.name)}
              {#if connector.name}
                <button
                  id={connector.name}
                  class="connector-tile-button"
                  on:click={() => goToConnectorForm(connector)}
                >
                  <div class="connector-wrapper">
                    <svelte:component this={ICONS[connector.name]} />
                  </div>
                </button>
              {/if}
            {/each}
          </div>
        </section> -->

        <div class="text-slate-500">
          Don't see what you're looking for?
          <button
            class="text-primary-500 hover:text-primary-600 font-medium"
            on:click={goToRequestConnector}
          >
            Request a new connector
          </button>
        </div>
      {/if}
    </Dialog.Content>
  </Dialog.Root>
{/if}

<style lang="postcss">
  section {
    @apply flex flex-col gap-y-3;
  }

  .connector-grid {
    @apply grid grid-cols-3 gap-4;
  }

  .connector-tile-button {
    aspect-ratio: 2/1;
    @apply basis-40;
    @apply border border-gray-300 rounded;
    @apply cursor-pointer overflow-hidden;
  }

  .connector-wrapper {
    @apply py-3 px-7 size-full;
    @apply flex items-center justify-center;
  }

  .connector-tile-button:hover {
    @apply bg-gray-100;
  }
</style>
