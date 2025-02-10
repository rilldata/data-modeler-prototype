<script lang="ts">
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    ConnectorDriverPropertyType,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import type { OlapDriver } from "../../connectors/olap/olap-config";
  import { inferSourceName } from "../sourceUtils";
  import { humanReadableErrorMessage } from "./errors";
  import { submitAddDataForm } from "./submitAddDataForm";
  import type { AddDataFormType } from "./types";
  import { getYupSchema, toYupFriendlyKey } from "./yupSchemas";

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let olapDriver: OlapDriver;
  export let onSuccess: (newFilePath: string) => Promise<void>;

  const formId = `add-data-form`;

  $: isSourceForm = formType === "source";
  $: isConnectorForm = formType === "connector";
  $: properties = isConnectorForm
    ? (connector.configProperties ?? [])
    : (connector.sourceProperties ?? []);

  let error: string | null = null;

  const schema = yup(getYupSchema[connector.name as keyof typeof getYupSchema]);

  const { form, errors, enhance, tainted, submit, submitting } = superForm(
    defaults(schema),
    {
      SPA: true,
      validators: schema,
      resetForm: false,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;
        try {
          const newFilePath = await submitAddDataForm(
            queryClient,
            formType,
            connector,
            values,
            olapDriver,
          );
          await onSuccess(newFilePath);
        } catch (e) {
          // Check that e?.response?.data conforms to the `RpcStatus` interface
          if (e?.response?.data?.code && e?.response?.data?.message) {
            error = humanReadableErrorMessage(
              connector.name,
              e?.response?.data?.code,
              e?.response?.data?.message,
            );
          } else {
            error = e;
          }
        }
      },
    },
  );

  function onStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;

    if (name === "path") {
      if ($tainted?.name) return;
      const name = inferSourceName(connector, value);
      if (name)
        form.update(
          ($form) => {
            $form.name = name;
            return $form;
          },
          { taint: false },
        );
    }
  }
</script>

<div class="h-full w-full flex flex-col gap-y-4">
  <form
    class="overflow-y-auto"
    id={formId}
    use:enhance
    on:submit|preventDefault={submit}
  >
    {#if isSourceForm}
      <div class="pb-2 text-slate-500">
        Need help? Refer to our
        <a
          href="https://docs.rilldata.com/build/connect"
          rel="noreferrer noopener"
          target="_blank">docs</a
        > for more information.
      </div>
    {/if}
    {#if error}
      <div class="pb-0.5">
        <SubmissionError message={error} />
      </div>
    {/if}

    {#each properties as property (property.key)}
      {#if property.key !== undefined && !property.noPrompt}
        {@const label =
          property.displayName + (property.required ? "" : " (optional)")}
        <div class="py-1.5">
          {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
            <Input
              id={toYupFriendlyKey(property.key)}
              label={property.displayName}
              placeholder={property.placeholder}
              optional={!property.required}
              secret={property.secret}
              hint={property.hint}
              errors={$errors[toYupFriendlyKey(property.key)]}
              bind:value={$form[toYupFriendlyKey(property.key)]}
              onInput={(_, e) => onStringInputChange(e)}
              alwaysShowError
            />
          {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
            <label for={property.key} class="flex items-center">
              <input
                id={property.key}
                type="checkbox"
                bind:checked={$form[property.key]}
                class="h-5 w-5"
              />
              <span class="ml-2 text-sm">{label}</span>
            </label>
          {:else if property.type === ConnectorDriverPropertyType.TYPE_INFORMATIONAL}
            <InformationalField
              description={property.description}
              hint={property.hint}
              href={property.docsUrl}
            />
          {/if}
        </div>
      {/if}
    {/each}
  </form>
  <slot name="actions" submitting={$submitting} />
</div>
