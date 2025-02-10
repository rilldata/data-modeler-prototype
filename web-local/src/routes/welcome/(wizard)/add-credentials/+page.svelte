<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import AddDataForm from "@rilldata/web-common/features/sources/modal/AddDataForm.svelte";
  import LocalSourceUpload from "@rilldata/web-common/features/sources/modal/LocalSourceUpload.svelte";
  import { createRuntimeServiceListConnectorDrivers } from "@rilldata/web-common/runtime-client";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { onboardingState } = data;
  const { managementType, olapDriver, firstDataSource } = onboardingState;

  // Get connector driver
  const connectorsQuery = createRuntimeServiceListConnectorDrivers();
  $: firstConnectorName =
    $managementType === "rill-managed" ? $firstDataSource : $olapDriver;
  $: connectorDriver = $connectorsQuery.data?.connectors?.find(
    (c) => c.name === firstConnectorName,
  );

  async function onContinue(filePath: string) {
    if ($managementType === "rill-managed") {
      await onboardingState.complete();
      // Navigate to the new source (whether there's an error or not)
      await goto(`/files/${filePath}`);
    } else {
      // Continue in the onboarding wizard to create a dashboard
      await goto(`/welcome/make-your-first-dashboard`);
    }
  }

  function onBack() {
    try {
      onboardingState.cleanUp();
    } catch (e) {
      console.error(e);
    }
  }
</script>

{#if $firstDataSource === "local_file"}
  <div class="flex flex-col gap-y-4">
    <LocalSourceUpload onSuccess={onContinue}>
      <svelte:fragment slot="actions">
        <Button
          href="/welcome/select-connectors"
          on:click={onBack}
          type="link"
          large
        >
          Back
        </Button>
      </svelte:fragment>
    </LocalSourceUpload>
  </div>
{:else if connectorDriver}
  <div class="w-[496px]">
    <h2 class="text-lead pb-2">Connect to {connectorDriver.displayName}</h2>
    <AddDataForm
      formType={$managementType === "self-managed" ? "connector" : "source"}
      connector={connectorDriver}
      olapDriver={$olapDriver}
      onSuccess={onContinue}
    >
      <svelte:fragment slot="actions" let:submitting>
        <div class="flex flex-col gap-y-2">
          <Button
            type="primary"
            form="add-data-form"
            submitForm
            disabled={submitting}
            large
          >
            {submitting
              ? "Testing connection..."
              : $firstDataSource
                ? "Add data"
                : "Connect"}
          </Button>
          <Button
            type="link"
            large
            href="/welcome/select-connectors"
            on:click={onBack}
          >
            Back
          </Button>
        </div>
      </svelte:fragment>
    </AddDataForm>
  </div>
{/if}
