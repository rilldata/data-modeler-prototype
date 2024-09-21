<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import CriteriaForm from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaForm.svelte";
  import { CriteriaGroupOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { getEmptyMeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { Trash2Icon } from "lucide-svelte";
  import { createForm } from "svelte-forms-lib";

  export let formState: ReturnType<typeof createForm<AlertFormValues>>;

  const { form } = formState;

  function handleDeleteCriteria(index: number) {
    $form["criteria"] = $form["criteria"].filter((_, i: number) => i !== index);
  }

  function handleAddCriteria() {
    $form["criteria"] = $form["criteria"].concat({
      ...getEmptyMeasureFilterEntry(),
      measure: $form.measures[0],
    });
  }
</script>

{#if $form["criteria"]}
  <div class="flex flex-col gap-2">
    {#each $form["criteria"] as _, index}
      {#if index > 0}
        <div class="flex flex-row items-center justify-center">
          <div class="mr-2">
            <Select
              bind:value={$form["criteriaOperation"]}
              id="field"
              label=""
              options={CriteriaGroupOptions}
              placeholder="Measure"
            />
          </div>
        </div>
      {/if}
      <div class="flex flex-col gap-2">
        <div class="flex flex-row items-center">
          <div class="w-full text-lg">{index + 1}</div>
          <button class="mr-2" on:click={() => handleDeleteCriteria(index)}>
            <Trash2Icon size="16px" />
          </button>
        </div>
        <CriteriaForm {formState} {index} />
      </div>
    {/each}
    <Button type="dashed" on:click={handleAddCriteria}>+ Add Criteria</Button>
  </div>
{/if}
