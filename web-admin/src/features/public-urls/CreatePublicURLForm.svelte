<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceIssueMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button, IconButton } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { copyToClipboard } from "@rilldata/actions";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Pencil } from "lucide-svelte";
  import { DateTime } from "luxon";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import {
    convertDateToMinutes,
    getExploreFields,
    getSanitizedDashboardStateParam,
    hasDashboardDimensionThresholdFilter,
    hasDashboardWhereFilter,
  } from "./form-utils";

  const queryClient = useQueryClient();
  const StateManagers = getStateManagers();

  const {
    dashboardStore,
    exploreName,
    selectors: {
      measures: { visibleMeasures },
      dimensions: { visibleDimensions },
    },
  } = StateManagers;

  $: ({ organization, project, dashboard } = $page.params);

  $: isTitleEmpty = $form.title.trim() === "";

  $: exploreFields = getExploreFields(
    $dashboardStore,
    $visibleDimensions,
    $visibleMeasures,
  );

  $: sanitizedState = getSanitizedDashboardStateParam(
    $dashboardStore,
    exploreFields,
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
              resourceType: ResourceKind.Explore as string,
              resourceName: dashboard,
              filter: hasWhereFilter ? $dashboardStore.whereFilter : undefined,
              fields: exploreFields,
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
  $: hasDimensionThresholdFilter =
    hasDashboardDimensionThresholdFilter($dashboardStore);

  $: if (setExpiration && $form.expiresAt === null) {
    // When `setExpiration` is toggled, initialize the expiration time to 60 days from today
    $form.expiresAt = DateTime.now().plus({ days: 60 }).toISO();
  } else if (!setExpiration) {
    $form.expiresAt = null;
  }

  $: ({ length: allErrorsLength } = $allErrors);

  $: includingTomorrowDate = DateTime.now().plus({ days: 1 }).startOf("day");

  let popoverOpen = false;
</script>

{#if !token}
  <form id={formId} on:submit|preventDefault={submit} use:enhance>
    <div class="flex flex-col gap-y-4">
      <h3 class="text-xs text-gray-800 font-normal">
        Create a shareable public URL for this view.
      </h3>

      <div class="flex flex-col gap-y-1">
        <Input
          id="name-input"
          bind:value={$form.title}
          placeholder="Label this URL"
        />
      </div>
    </div>

    <div
      class="mt-4"
      class:mb-4={!hasWhereFilter && !hasDimensionThresholdFilter}
    >
      <div class="flex items-center gap-x-2">
        <Switch small id="has-expiration" bind:checked={setExpiration} />
        <Label class="text-xs" for="has-expiration">Set expiration</Label>
      </div>
      {#if setExpiration}
        <div class="flex items-center gap-x-1 pl-[30px]">
          <label for="expires-at" class="text-slate-500 font-medium">
            Access expires {new Date($form.expiresAt).toLocaleDateString(
              "en-US",
              { year: "numeric", month: "short", day: "numeric" },
            )}
          </label>
          <Popover bind:open={popoverOpen}>
            <PopoverTrigger>
              <IconButton>
                <Pencil size="14px" class="text-primary-600" />
              </IconButton>
            </PopoverTrigger>
            <PopoverContent align="end" class="p-0">
              <Calendar
                selection={DateTime.fromISO($form.expiresAt)}
                singleDaySelection
                minDate={includingTomorrowDate}
                firstVisibleMonth={DateTime.fromISO($form.expiresAt)}
                onSelectDay={(date) => {
                  $form.expiresAt = date.toISO();
                  popoverOpen = false;
                }}
              />
            </PopoverContent>
          </Popover>
        </div>
      {/if}
    </div>

    <!-- TODO: revisit when time range lock is implemented -->
    <!-- <div class="mt-4" class:mb-4={!hasWhereFilter}>
      <div class="flex items-center gap-x-2">
        <Switch small id="lock-time-range" bind:checked={lockTimeRange} />

        <div class="flex flex-row items-center gap-x-1">
          <Label class="text-xs" for="lock-time-range">Lock time range</Label>
          <Tooltip location="right" alignment="middle" distance={8}>
            <div class="text-gray-500">
              <InfoCircle size="12px" />
            </div>
            <TooltipContent maxWidth="400px" slot="tooltip-content">
              Only data within this range will be visible
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
      {#if lockTimeRange}
        <div class="w-full pl-[30px]">
          <label for="lock-time-range" class="text-slate-500 font-medium">
            {#if interval.isValid}
              <RangeDisplay {interval} grain={activeTimeGrain} {abbreviation} />
            {/if}
          </label>
        </div>
      {/if}
    </div> -->

    {#if hasWhereFilter || hasDimensionThresholdFilter}
      <hr class="border-gray-200 mt-4 mb-4" />

      <div class="flex flex-col gap-y-1">
        <p class="text-xs text-gray-800 font-normal">
          The following filters will be locked and hidden:
        </p>
        <div class="flex flex-row gap-1 mt-2">
          <FilterChipsReadOnly
            exploreName={$exploreName}
            filters={$dashboardStore.whereFilter}
            dimensionThresholdFilters={$dashboardStore.dimensionThresholdFilters}
            timeRange={undefined}
            comparisonTimeRange={undefined}
          />
        </div>
      </div>

      <p class="text-xs text-gray-800 font-normal mt-4 mb-4">
        Measures and dimensions will be limited to current visible set.
      </p>
    {/if}

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
    @apply flex flex-col;
  }

  h3 {
    @apply font-semibold;
  }
</style>
