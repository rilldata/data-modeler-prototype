<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/SingleFieldInput.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import FieldConfigDropdown from "./FieldConfigDropdown.svelte";

  export let key: string;
  export let config: { label?: string };
  export let metricsView: string;
  export let fieldConfig: FieldConfig;
  export let onChange: (updatedConfig: FieldConfig) => void;

  const {
    canvasEntity: {
      spec: { getTimeDimensionForMetricView },
    },
  } = getCanvasStateManagers();

  $: isDimension = key === "x";
  $: timeDimension = getTimeDimensionForMetricView(metricsView);

  function updateFieldConfig(fieldName: string) {
    const isTime = $timeDimension && fieldName === $timeDimension;

    let updatedConfig: FieldConfig;
    if (isTime && $timeDimension) {
      updatedConfig = {
        ...fieldConfig,
        field: $timeDimension,
        type: "temporal",
      };
    } else {
      updatedConfig = {
        ...fieldConfig,
        field: fieldName,
        type: isTime ? "temporal" : isDimension ? "nominal" : "quantitative",
      };
    }

    onChange(updatedConfig);
  }

  function updateFieldProperty(property: keyof FieldConfig, value: any) {
    const updatedConfig: FieldConfig = {
      ...fieldConfig,
      [property]: value,
    };

    onChange(updatedConfig);
  }
</script>

<div class="gap-y-1">
  <div class="flex justify-between items-center">
    <InputLabel small label={config.label ?? key} id={key} />
    <FieldConfigDropdown {key} {fieldConfig} onChange={updateFieldProperty} />
  </div>

  <SingleFieldInput
    metricName={metricsView}
    id={`${key}-field`}
    type={isDimension ? "dimension" : "measure"}
    includeTime
    selectedItem={fieldConfig?.field}
    onSelect={async (field) => {
      updateFieldConfig(field);
    }}
  />
</div>
