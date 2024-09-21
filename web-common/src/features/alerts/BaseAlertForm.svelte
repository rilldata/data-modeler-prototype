<script lang="ts">
  import { DialogTitle } from "@rilldata/web-common/components/dialog-v2";
  import * as DialogTabs from "@rilldata/web-common/components/dialog/tabs";
  import {
    generateAlertName,
    getTouched,
  } from "@rilldata/web-common/features/alerts/utils";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { X } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import type { createForm } from "svelte-forms-lib";
  import Button from "../../components/button/Button.svelte";
  import AlertDialogCriteriaTab from "./criteria-tab/AlertDialogCriteriaTab.svelte";
  import AlertDialogDataTab from "./data-tab/AlertDialogDataTab.svelte";
  import AlertDialogDeliveryTab from "./delivery-tab/AlertDialogDeliveryTab.svelte";
  import {
    type AlertFormValues,
    checkIsTabValid,
    FieldsByTab,
  } from "./form-utils";

  export let formState: ReturnType<typeof createForm<AlertFormValues>>;
  export let isEditForm: boolean;

  const dispatch = createEventDispatcher();

  const formId = isEditForm ? "edit-alert-form" : "create-alert-form";
  const dialogTitle = isEditForm ? "Edit Alert" : "Create Alert";

  const { form, errors, handleSubmit, validateField, isSubmitting, touched } =
    formState;

  const tabs = ["Data", "Criteria", "Delivery"];

  /**
   * Because this form's fields are spread over multiple tabs, we implement our own `isValid` logic for each tab.
   * A tab is valid (i.e. it's okay to proceed to the next tab) if:
   * 1) The tab's required fields are filled out
   * 2) The tab's fields don't have errors.
   */
  $: isTabValid = checkIsTabValid(currentTabIndex, $form, $errors);

  let currentTabIndex = 0;

  $: metricsViewName = $form["metricsViewName"]; // memoise to avoid rerenders
  $: metricsView = useMetricsView($runtime.instanceId, metricsViewName);

  function handleCancel() {
    if (getTouched($touched)) {
      dispatch("cancel");
    } else {
      dispatch("close");
    }
  }

  function handleBack() {
    currentTabIndex -= 1;
  }

  function handleNextTab() {
    if (!isTabValid) {
      FieldsByTab[currentTabIndex]?.forEach((field) => validateField(field));
      return;
    }
    currentTabIndex += 1;

    if (isEditForm || currentTabIndex !== 2 || $touched.name) {
      return;
    }
    // if the user came to the delivery tab and name was not changed then auto generate it
    const name = generateAlertName($form, $metricsView.data ?? {});
    if (!name) return;
    $form.name = name;
  }

  $: measures = $form.measures;
  function measureUpdated(measures: string[]) {
    $form.criteria.forEach((c) => {
      if (measures.includes(c.measure)) return;
      c.measure = measures[0];
    });
  }
  $: measureUpdated(measures);
</script>

<!-- 802px = 1px border on each side of the form + 3 tabs with a 200px fixed-width -->
<form
  class="transform bg-white rounded-md flex flex-col w-[802px]"
  id={formId}
  on:submit|preventDefault={handleSubmit}
>
  <DialogTitle
    class="px-6 py-4 text-gray-900 text-lg font-semibold leading-7 flex flex-row items-center justify-between"
  >
    <div>{dialogTitle}</div>
    <Button type="link" noStroke compact on:click={handleCancel}>
      <X strokeWidth={3} size={16} class="text-black" />
    </Button>
  </DialogTitle>
  <DialogTabs.Root value={tabs[currentTabIndex]}>
    <DialogTabs.List class="border-t border-gray-200">
      {#each tabs as tab, i}
        <!-- inner width is 800px. so, width = ceil(800/3) = 267 -->
        <DialogTabs.Trigger value={tab} tabIndex={i} class="w-[267px]">
          {tab}
        </DialogTabs.Trigger>
      {/each}
    </DialogTabs.List>
    <div class="p-3 bg-slate-100 h-[600px] overflow-auto">
      <DialogTabs.Content {currentTabIndex} tabIndex={0} value={tabs[0]}>
        <AlertDialogDataTab {formState} />
      </DialogTabs.Content>
      <DialogTabs.Content {currentTabIndex} tabIndex={1} value={tabs[1]}>
        <AlertDialogCriteriaTab {formState} />
      </DialogTabs.Content>
      <DialogTabs.Content {currentTabIndex} tabIndex={2} value={tabs[2]}>
        <AlertDialogDeliveryTab {formState} />
      </DialogTabs.Content>
    </div>
  </DialogTabs.Root>
  <div class="px-6 py-3 flex items-center gap-x-2">
    <div class="grow" />
    {#if currentTabIndex === 0}
      <Button on:click={handleCancel} type="secondary">Cancel</Button>
    {:else}
      <Button on:click={handleBack} type="secondary">Back</Button>
    {/if}
    {#if currentTabIndex !== 2}
      <Button type="primary" on:click={handleNextTab}>Next</Button>
    {:else}
      <Button type="primary" disabled={$isSubmitting} form={formId} submitForm>
        Create
      </Button>
    {/if}
  </div>
</form>
