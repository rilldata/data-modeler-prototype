<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceIssueMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { useQueryClient } from "@tanstack/svelte-query";
  import {
    convertDateToMinutes,
    getMetricsViewFields,
    getSanitizedDashboardStateParam,
    hasDashboardWhereFilter,
  } from "./form-utils";
  import Filters from "@rilldata/web-common/features/dashboards/filters/Filters.svelte";

  const queryClient = useQueryClient();
  const {
    dashboardStore,
    metricsViewName,
    selectors: {
      measures: { visibleMeasures },
      dimensions: { visibleDimensions },
    },
  } = getStateManagers();

  $: ({ organization, project } = $page.params);

  $: isTitleEmpty = $form.title.trim() === "";

  $: metricsViewFields = getMetricsViewFields(
    $dashboardStore,
    $visibleDimensions,
    $visibleMeasures,
  );

  $: sanitizedState = getSanitizedDashboardStateParam(
    $dashboardStore,
    metricsViewFields,
  );

  let token: string;
  let setExpiration = false;
  let apiError: string;

  const formId = "create-public-url-form";

  const initialValues = {
    expiresAt: null,
    title: "",
  };

  const validationSchema = object({
    expiresAt: string().nullable(),
    title: string().required("Title is required"),
  });

  const issueMagicAuthToken = createAdminServiceIssueMagicAuthToken();

  const { form, enhance, submit, allErrors, submitting } = superForm(
    defaults(initialValues, yup(validationSchema)),
    {
      SPA: true,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        try {
          const { token: _token } = await $issueMagicAuthToken.mutateAsync({
            organization,
            project,
            data: {
              metricsView: $metricsViewName,
              metricsViewFilter: hasWhereFilter
                ? $dashboardStore.whereFilter
                : undefined,
              metricsViewFields,
              ttlMinutes: setExpiration
                ? convertDateToMinutes(values.expiresAt).toString()
                : undefined,
              state: sanitizedState ? sanitizedState : undefined,
              title: values.title,
            },
          });
          token = _token;

          copyToClipboard(
            `${window.location.origin}/${organization}/${project}/-/share/${token}`,
            "URL copied to clipboard",
          );

          void queryClient.invalidateQueries(
            getAdminServiceListMagicAuthTokensQueryKey(organization, project),
          );
        } catch (error) {
          const typedError = error as HTTPError;
          apiError = typedError.response?.data?.message ?? typedError.message;
        }
      },
    },
  );

  $: hasWhereFilter = hasDashboardWhereFilter($dashboardStore);

  $: if (setExpiration && $form.expiresAt === null) {
    // When `setExpiration` is toggled, initialize the expiration time to 60 days from today
    $form.expiresAt = new Date(Date.now() + 60 * 24 * 60 * 60 * 1000)
      .toISOString()
      .slice(0, 10); // ISO string formatted for input[type="date"]
  } else if (!setExpiration) {
    $form.expiresAt = null;
  }

  $: ({ length: allErrorsLength } = $allErrors);
</script>

{#if !token}
  <form id={formId} on:submit|preventDefault={submit} use:enhance>
    <div class="information-container">
      <h3>Create a shareable public URL for this view</h3>
      <ul>
        <li>Measures and dimensions will be limited to current visible set.</li>
        <li>Filters will be locked and hidden.</li>
      </ul>

      <div class="name-input-container">
        <Label for="name-input" class="text-xs">URL label</Label>
        <input
          id="name-input"
          type="text"
          bind:value={$form.title}
          placeholder="Name this URL"
          class="w-full px-3 py-2 border border-gray-300 rounded-md"
        />
      </div>

      <!-- Filters -->
      {#if hasWhereFilter}
        <Filters readOnly timePills={false} />
      {/if}
    </div>

    <!-- Expiration -->
    <div>
      <div class="has-expiration-container">
        <Switch small id="has-expiration" bind:checked={setExpiration} />
        <Label class="text-xs" for="has-expiration">Set expiration</Label>
      </div>
      {#if setExpiration}
        <div class="expires-at-container">
          <label for="expires-at" class="expires-at-label w-1/3">
            Access expires
          </label>
          <!-- TODO: use a Rill date picker, once we have one that can select a single day -->
          <!-- Minimum date is tomorrow -->
          <input
            id="expires-at"
            type="date"
            bind:value={$form.expiresAt}
            min={new Date(Date.now() + 24 * 60 * 60 * 1000)
              .toISOString()
              .slice(0, 10)}
            class="w-2/3"
          />
        </div>
      {/if}
    </div>

    <Button
      type="primary"
      disabled={$submitting || isTitleEmpty}
      form={formId}
      submitForm
    >
      Create and copy URL
    </Button>

    {#if allErrorsLength > 0}
      {#each $allErrors as error (error.path)}
        <div class="text-red-500">{error.messages}</div>
      {/each}
    {:else if apiError}
      <div class="text-red-500">{apiError}</div>
    {/if}
  </form>
{:else}
  <!-- A successful form submission will automatically copy the link to the clipboard -->
  <div class="flex flex-col gap-y-2">
    <h3>Success! URL copied to clipboard.</h3>
  </div>
{/if}

<style lang="postcss">
  form {
    @apply flex flex-col gap-y-4;
  }

  h3 {
    @apply font-semibold;
  }

  .information-container {
    @apply flex flex-col gap-y-2;
  }

  ul {
    @apply list-disc list-inside;
  }

  .name-input-container {
    @apply flex flex-col gap-y-1;
  }

  .has-expiration-container {
    @apply flex items-center gap-x-2;
  }

  .expires-at-container {
    @apply mt-2;
    @apply flex items-center gap-x-2;
  }

  .expires-at-label {
    @apply text-slate-500 font-medium;
  }

  input {
    @apply size-full outline-none border-0;
  }

  #name-input {
    @apply flex justify-center items-center overflow-hidden;
    @apply h-8 pl-2 w-full;
    @apply border border-gray-300 rounded-sm;
    @apply text-xs;
  }

  #name-input:focus-within {
    @apply border-primary-500;
  }

  #name-input::placeholder {
    @apply text-xs;
  }
</style>
