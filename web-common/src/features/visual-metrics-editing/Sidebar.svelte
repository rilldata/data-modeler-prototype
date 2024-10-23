<script lang="ts">
  import { YAMLDimension, YAMLMeasure, type MenuOption } from "./lib";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import type { V1ProfileColumn } from "@rilldata/web-common/runtime-client";
  import { FileArtifact } from "../entity-management/file-artifact";
  import { parseDocument, YAMLMap, YAMLSeq } from "yaml";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import SimpleSqlExpression from "./SimpleSQLExpression.svelte";
  import { eventBus } from "@rilldata/events";
  import { NUMERICS } from "@rilldata/web-common/lib/duckdb-data-types";

  export let item: YAMLMeasure | YAMLDimension;
  export let fileArtifact: FileArtifact;
  export let index: number;
  export let type: "measures" | "dimensions";
  export let columns: V1ProfileColumn[];
  export let editing: boolean;
  export let unsavedChanges: boolean;
  export let switchView: () => void;
  export let onDelete: () => void;
  export let onCancel: (unsavedChanges: boolean) => void;
  export let resetEditing: () => void;

  let editingClone = structuredClone(item);

  let columnOptions: MenuOption[] = columns.map(({ name, type }) => ({
    value: name ?? "",
    label: name ?? "",
    type,
  }));

  let properties: Record<
    string,
    Array<{
      label: string;
      selected: number;
      optional?: true;
      fontFamily?: string;
      fields: Array<{
        key: string;
        hint?: string;
        label: string;
        options?: MenuOption[];
        placeholder?: string;
        boolean?: true;
      }>;
    }>
  > = {
    measures: [
      {
        label: "SQL expression",
        fontFamily: `"Source Code Variable", monospace`,
        fields: [
          {
            key: "expression",
            label: "Simple",
            options: columnOptions,
            placeholder: "Column from model",
          },
          {
            key: "expression",
            label: "Advanced",
          },
        ],
        selected: editingClone.expression ? 1 : 0,
      },
      {
        label: "Name",
        fontFamily: `"Source Code Variable", monospace`,
        fields: [
          {
            key: "name",
            hint: "A stable identifier for use in code",
            label: "Name",
          },
        ],
        selected: 0,
      },
      {
        optional: true,
        label: "Label",
        fields: [
          {
            key: "label",
            hint: "Used on dashboards and charts. Inferred from name when not provided",

            label: "Label",
          },
        ],
        selected: 0,
      },
      {
        label: "Format",
        fields: [
          {
            label: "Simple",
            key: "format_preset",
            options: Object.values(FormatPreset).map((value) => ({
              value,
              label: value,
            })),
            placeholder: "Select a format",
          },
          {
            label: "d3-format",
            key: "format_d3",
          },
        ],
        selected: editingClone["format_d3"] ? 1 : 0,
      },
      {
        optional: true,
        label: "Description",
        fields: [
          {
            key: "description",

            label: "Description",
          },
        ],
        selected: 0,
      },
      {
        label: "Summable metric",
        optional: true,
        fields: [
          {
            key: "valid_percent_of_total",
            hint: "Values can be added to yield a valid sum",
            label: "Summable metric",
            boolean: true,
          },
        ],
        selected: 0,
      },
    ],
    dimensions: [
      {
        label: "SQL expression",
        fontFamily: `"Source Code Variable", monospace`,
        fields: [
          {
            placeholder: "Column from model",
            options: columnOptions,
            key: "column",
            label: "Simple",
          },
          {
            label: "Advanced",
            key: "expression",
          },
        ],
        selected: editingClone.expression ? 1 : 0,
      },
      {
        label: "Name",
        fontFamily: `"Source Code Variable", monospace`,
        fields: [
          {
            key: "name",
            hint: "A stable identifier for use in code",
            label: "Name",
          },
        ],
        selected: 0,
      },
      {
        label: "Label",
        optional: true,
        fields: [
          {
            key: "label",
            hint: "Used on dashboards and charts. Inferred from name when not provided",

            label: "Label",
          },
        ],
        selected: 0,
      },

      {
        label: "Description",
        optional: true,
        fields: [
          {
            key: "description",

            label: "Description",
          },
        ],
        selected: 0,
      },
    ],
  };

  $: numericColumns = columnOptions.filter(
    ({ type }) => type && NUMERICS.has(type),
  );

  $: ({ remoteContent, localContent, saveContent } = fileArtifact);

  $: requiredPropertiesUnfilled = properties[type]
    .filter(({ optional, fields, selected }) => {
      const value = editingClone[fields[selected].key];
      return !optional && (value === undefined || value === "");
    })
    .map(({ label }) => label);

  $: unsavedChanges = Object.keys(editingClone).some(
    (key) => editingClone[key] !== item?.[key],
  );

  async function saveChanges() {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    let sequence = parsedDocument.get(type);

    if (!(sequence instanceof YAMLSeq) || sequence.items.length === 0) {
      sequence = new YAMLSeq();
      parsedDocument.set(type, sequence);
    }

    if (!(sequence instanceof YAMLSeq)) {
      throw new Error("Invalid YAML document");
    }

    const items = sequence.items as YAMLMap[];

    const newItem = new YAMLMap();

    properties[type].forEach(({ selected, fields }) => {
      const { key } = fields[selected];
      if (editingClone[key] || editingClone[key] === false)
        newItem.set(key, editingClone[key]);
    });

    if (editing) {
      items[index] = newItem;
    } else {
      items.push(newItem);
    }

    await saveContent(parsedDocument.toString());
    resetEditing();

    eventBus.emit("notification", { message: "Item saved", type: "success" });
  }
</script>

<svelte:window
  on:keydown={(e) => {
    if (e.key === "Escape") onCancel(unsavedChanges);
  }}
/>

<div
  class="h-full w-[320px] bg-background flex-none flex flex-col border select-none rounded-[2px]"
>
  <h1 class="pt-6 px-5">{editing ? "Edit" : "Add"} {type.slice(0, -1)}</h1>

  <div
    class="px-5 flex flex-col gap-y-3 w-full h-fit overflow-y-auto overflow-x-visible"
  >
    {#each properties[type] as { fields, selected, label, optional, fontFamily } (label)}
      {@const { hint, key, options, placeholder, boolean } = fields[selected]}
      {#if boolean}
        <div class="flex gap-x-2 items-center h-full bg-white rounded-full">
          <Switch bind:checked={editingClone[key]} id="auto-save" medium />
          <Label class="font-medium text-sm" for="auto-save">{label}</Label>
          {#if hint}
            <Tooltip location="left">
              <div class="text-gray-500">
                <InfoCircle size="13px" />
              </div>
              <TooltipContent slot="tooltip-content">
                {hint}
              </TooltipContent>
            </Tooltip>
          {/if}
        </div>
      {:else if key === "expression" && type === "measures"}
        <SimpleSqlExpression
          {editing}
          columns={columnOptions}
          {numericColumns}
          bind:expression={editingClone.expression}
          bind:name={editingClone.name}
        />
      {:else}
        <Input
          id="vme-{label}"
          textClass="text-sm"
          {label}
          {hint}
          {options}
          {optional}
          {fontFamily}
          {placeholder}
          multiline={key === "description"}
          enableSearch={key === "column"}
          bind:selected
          bind:value={editingClone[key]}
          fields={fields.map(({ label }) => label)}
          onChange={(e) => {
            if (!editing && key === "column" && type === "dimensions") {
              editingClone.name = e;
            }
          }}
          sameWidth={true}
        />
      {/if}
    {/each}

    <span />
  </div>

  <div class="flex flex-col gap-y-3 mt-auto border-t px-5 pb-6 pt-3">
    <!-- <h2>Preview</h2> -->

    <p>
      For more options,
      <button on:click={switchView} class="text-primary-600 font-medium">
        edit in YAML
      </button>
    </p>
    <div class="flex justify-{editing ? 'between' : 'end'}">
      {#if editing}
        <Button type="text" on:click={onDelete}>Delete</Button>
      {/if}
      <div class="flex gap-x-2 self-end">
        <Button
          type="secondary"
          on:click={() => {
            onCancel(unsavedChanges);
          }}
        >
          Cancel
        </Button>
        <Tooltip
          location="top"
          distance={8}
          suppress={!requiredPropertiesUnfilled.length}
        >
          <Button
            type="primary"
            on:click={saveChanges}
            disabled={requiredPropertiesUnfilled.length > 0 || !unsavedChanges}
          >
            {editing ? "Save changes" : "Add " + type.slice(0, -1)}
          </Button>

          <TooltipContent slot="tooltip-content">
            Required: {requiredPropertiesUnfilled.join(", ")}
          </TooltipContent>
        </Tooltip>
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  h1 {
    @apply text-lg font-semibold mb-4;
  }

  /* h2 {
    @apply text-sm font-medium;
  } */
</style>
