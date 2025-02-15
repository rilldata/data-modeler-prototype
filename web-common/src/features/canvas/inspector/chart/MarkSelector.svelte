<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/SingleFieldInput.svelte";

  export let key: string;
  export let label: string;
  export let metricsView: string;
  export let value: FieldConfig | string;
  export let onChange: (updatedConfig: FieldConfig | string) => void;

  $: selected = !value || typeof value === "string" ? 0 : 1;

  // TODO: Replace with theme primary color
  $: color = typeof value === "string" ? value : "rgb(117, 126, 255)";

  function updateFieldConfig(field: string) {
    const updatedConfig: FieldConfig = {
      field,
      type: "nominal",
    };
    onChange(updatedConfig);
  }
</script>

<div class="space-y-2">
  <InputLabel small {label} id={key} />

  <FieldSwitcher
    small
    fields={["One color", "Split by"]}
    {selected}
    onClick={(_, field) => {
      if (field === "One color") {
        selected = 0;
        onChange(color);
      } else if (field === "Split by") {
        selected = 1;
      }
    }}
  />

  {#if selected === 0}
    <ColorInput
      small
      stringColor={color}
      label=""
      onChange={(color) => {
        onChange(color);
      }}
    />
  {:else if selected === 1}
    <SingleFieldInput
      metricName={metricsView}
      id={`${key}-field`}
      type="dimension"
      selectedItem={typeof value === "string" ? undefined : value?.field}
      onSelect={async (field) => {
        updateFieldConfig(field);
      }}
    />
  {/if}
</div>
