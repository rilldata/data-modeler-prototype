<script lang="ts">
  import { InfoIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import * as Select from "@rilldata/web-common/components/select";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";

  const dispatch = createEventDispatcher();

  export let value: string | string[];
  export let id: string;
  export let label: string;
  export let options: {
    value: string;
    label: string;
    disabled?: boolean;
    tooltip?: string;
  }[];
  export let placeholder: string = "";
  export let optional: boolean = false;
  export let tooltip: string = "";
  export let width: number | null = null;
  export let multiple = false;

  $: selected = multiple
    ? options.filter((option) => value.includes(option.value))
    : options.find((option) => option.value === value);
</script>

<div class="flex flex-col gap-y-2">
  {#if label?.length}
    <label for={id} class="text-sm flex items-center gap-x-1">
      <span class="text-gray-800 font-medium">
        {label}
      </span>
      {#if optional}
        <span class="text-gray-500">(optional)</span>
      {/if}
      {#if tooltip}
        <Tooltip.Root portal="body">
          <Tooltip.Trigger>
            <InfoIcon class="text-gray-500" size="14px" strokeWidth={2} />
          </Tooltip.Trigger>
          <Tooltip.Content side="right">
            {tooltip}
          </Tooltip.Content>
        </Tooltip.Root>
      {/if}
    </label>
  {/if}

  <Select.Root
    {selected}
    onSelectedChange={(newSelection) => {
      if (!newSelection) return;
      if (Array.isArray(newSelection)) {
        value = newSelection.map((v) => v.value);
      } else {
        value = newSelection.value;
      }
      dispatch("change", value);
    }}
    items={options}
    {multiple}
  >
    <Select.Trigger class="px-3 gap-x-2 {width && `w-[${width}px]`}">
      <Select.Value
        {placeholder}
        class="text-[12px] {!selected ? 'text-gray-400' : ''}"
      />
    </Select.Trigger>

    <Select.Content
      sameWidth={false}
      align="start"
      class="max-h-80 overflow-y-auto"
    >
      {#each options as { value, label, disabled, tooltip } (value)}
        <Select.Item {value} {label} {disabled} class="text-[12px]">
          {#if tooltip}
            <Tooltip.Root portal="body">
              <Tooltip.Trigger class="select-tooltip cursor-default">
                {label ?? value}
              </Tooltip.Trigger>
              <Tooltip.Content side="right" sideOffset={8}>
                {tooltip}
              </Tooltip.Content>
            </Tooltip.Root>
          {:else}
            {label ?? value}
          {/if}
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>
</div>
