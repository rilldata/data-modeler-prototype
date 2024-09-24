<script lang="ts">
  import { page } from "$app/stores";
  import type { BookmarkFormValues } from "@rilldata/web-admin/features/bookmarks/form-utils";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { createForm } from "svelte-forms-lib";
  import { InfoIcon } from "lucide-svelte";
  import Filters from "@rilldata/web-common/features/dashboards/filters/Filters.svelte";

  export let formState: ReturnType<typeof createForm<BookmarkFormValues>>;

  const { form, errors } = formState;

  // Adding it here to get a newline in
  const CategoryTooltip = `Your bookmarks can only be viewed by you.
Managed bookmarks will be available to all viewers of this dashboard.`;
</script>

<form
  class="flex flex-col gap-4 z-50"
  id="create-bookmark-dialog"
  on:submit|preventDefault={() => {
    /* Switch was triggering this causing clicking on them submitting the form */
  }}
>
  <div class="flex flex-col gap-y-2">
    <Label class="flex flex-col gap-y-1 text-sm">
      <div class="text-gray-800 font-medium">Filters</div>
    </Label>
    <Filters hideRanges hideIcons bookmarks />
  </div>
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
</form>
