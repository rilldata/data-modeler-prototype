<script lang="ts">
  import { page } from "$app/stores";
  import { getPrettySelectedTimeRange } from "@rilldata/web-admin/features/bookmarks/selectors";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { InfoIcon } from "lucide-svelte";
  import type { SuperForm } from "sveltekit-superforms";

  export let metricsViewName: string;
  export let exploreName: string;
  export let formState: SuperForm<
    {
      displayName?: string;
      description?: string;
      shared?: string;
      filtersOnly?: boolean;
      absoluteTimeRange?: boolean;
    },
    any
  >;

  $: ({ instanceId } = $runtime);

  $: exploreState = useExploreState(exploreName);

  let timeRange: V1TimeRange;
  $: timeRange = {
    isoDuration: $exploreState.selectedTimeRange?.name,
    start: $exploreState.selectedTimeRange?.start?.toISOString() ?? "",
    end: $exploreState.selectedTimeRange?.end?.toISOString() ?? "",
  };

  $: selectedTimeRange = getPrettySelectedTimeRange(
    queryClient,
    instanceId,
    metricsViewName,
    exploreName,
  );

  const { form, errors, enhance, submit } = formState;

  // Adding it here to get a newline in
  const CategoryTooltip = `Your bookmarks can only be viewed by you.
Managed bookmarks will be available to all viewers of this dashboard.`;
</script>

<form
  use:enhance
  class="flex flex-col gap-4 z-50"
  id="create-bookmark-dialog"
  on:submit|preventDefault={submit}
>
  <Input
    bind:value={$form["displayName"]}
    errors={$errors["displayName"]}
    id="displayName"
    label="Label"
  />
  <Input
    bind:value={$form["description"]}
    errors={$errors["description"]}
    id="description"
    label="Description"
    optional
  />
  <div class="flex flex-col gap-y-2">
    <Label class="flex flex-col gap-y-1 text-sm">
      <div class="text-gray-800 font-medium">Filters</div>
      <div class="text-gray-500">Inherited from underlying dashboard view.</div>
    </Label>
    <FilterChipsReadOnly
      dimensionThresholdFilters={$exploreState.dimensionThresholdFilters}
      filters={$exploreState.whereFilter}
      {exploreName}
      {timeRange}
    />
  </div>
  <ProjectAccessControls
    organization={$page.params.organization}
    project={$page.params.project}
  >
    <Select
      bind:value={$form["shared"]}
      id="shared"
      label="Category"
      options={[
        { value: "false", label: "Your bookmarks" },
        { value: "true", label: "Managed bookmarks" },
      ]}
      slot="manage-project"
      tooltip={CategoryTooltip}
    />
  </ProjectAccessControls>
  <div class="flex items-center space-x-2">
    <Switch bind:checked={$form["filtersOnly"]} id="filtersOnly" />
    <Label class="font-normal flex gap-x-1 items-center" for="filtersOnly">
      <span>Save filters only</span>
      <Tooltip distance={8}>
        <InfoIcon class="text-gray-500" size="14px" strokeWidth={2} />
        <TooltipContent
          class="whitespace-pre-line"
          maxWidth="600px"
          slot="tooltip-content"
        >
          Toggling this on will only save the filter set above, not the full
          dashboard layout and state.
        </TooltipContent>
      </Tooltip>
    </Label>
  </div>
  <div class="flex items-center space-x-2">
    <Switch bind:checked={$form["absoluteTimeRange"]} id="absoluteTimeRange" />
    <Label class="flex flex-col font-normal" for="absoluteTimeRange">
      <div class="text-left text-sm flex gap-x-1 items-center">
        <span>Absolute time range</span>
        <Tooltip distance={8}>
          <InfoIcon class="text-gray-500" size="14px" strokeWidth={2} />
          <TooltipContent
            class="whitespace-pre-line"
            maxWidth="600px"
            slot="tooltip-content"
          >
            The bookmark will use the dashboard's relative time if this toggle
            is off.
          </TooltipContent>
        </Tooltip>
      </div>
      <div class="text-gray-500 text-sm">{$selectedTimeRange}</div>
    </Label>
  </div>
</form>
