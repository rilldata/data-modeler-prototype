<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { getTypeOptions } from "@rilldata/web-common/features/alerts/criteria-tab/getTypeOptions";
  import { CriteriaOperationOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import { parseCriteriaError } from "@rilldata/web-common/features/alerts/criteria-tab/parseCriteriaError";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { useMetricsViewValidSpec } from "@rilldata/web-common/features/dashboards/selectors";
  import { debounce } from "@rilldata/utils";
  import { createForm } from "svelte-forms-lib";
  import { slide } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let formState: ReturnType<typeof createForm<AlertFormValues>>;
  export let index: number;

  const { form, errors, validateField } = formState;

  $: metricsView = useMetricsViewValidSpec(
    $runtime.instanceId,
    $form["metricsViewName"],
  );

  $: measure = $metricsView.data?.measures?.find(
    (m) => m.name === $form["measure"],
  );
  $: measureOptions = [
    {
      value: $form["measure"],
      label: measure?.label?.length
        ? measure.label
        : (measure?.expression ?? $form["measure"]),
    },
  ];
  $: selectedMeasure = $metricsView.data?.measures?.find(
    (m) => m.name === $form["criteria"][index].measure,
  );

  $: typeOptions = getTypeOptions($form, selectedMeasure);

  // Debounce the update of value. This avoids constant refetches
  let value: string = $form["criteria"][index].value1;
  const valueUpdater = debounce(() => {
    if ($form["criteria"][index]) $form["criteria"][index].value1 = value;
    void validateField("criteria");
  }, 500);

  // memoize `type` to avoid unnecessary calls to `validateField("criteria")`
  $: type = $form["criteria"][index].type;
  // changing type should re-trigger `criteria` validation,
  // especially when changed to/from a percent type
  $: if (type) void validateField("criteria");

  $: groupErr = parseCriteriaError($errors["criteria"], index);
</script>

<div class="flex flex-row gap-2">
  <Select
    bind:value={$form["criteria"][index].measure}
    id="field"
    label=""
    options={measureOptions}
    placeholder="Measure"
    width={160}
  />
  <Select
    bind:value={$form["criteria"][index].type}
    id="type"
    label=""
    options={typeOptions}
    placeholder="type"
    width={256}
  />
  <Select
    bind:value={$form["criteria"][index].operation}
    id="operation"
    label=""
    options={CriteriaOperationOptions}
    placeholder="Operator"
    width={70}
  />
  <!-- Error is not returned as an object for criteria[index]. We instead have parsed groupErr -->
  <Input
    alwaysShowError
    bind:value
    id="value"
    onInput={valueUpdater}
    placeholder={"0"}
    width="auto"
  />
</div>
{#if groupErr}
  <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
    {groupErr}
  </div>
{/if}
