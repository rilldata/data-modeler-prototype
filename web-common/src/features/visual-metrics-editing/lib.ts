import { writable } from "svelte/store";
import type { YAMLMap } from "yaml";
import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";

export const types: ItemType[] = ["measures", "dimensions"];

export type ItemType = "measures" | "dimensions";

export type MenuOption = { value: string; label: string; type?: string };

export const editingItem = writable<{
  item: YAMLMeasure | YAMLDimension;
  type: ItemType;
} | null>(null);

export const editingIndex = writable<number | null>(null);

export type Confirmation = {
  action: "cancel" | "delete" | "switch";
  type?: ItemType;
  model?: string;
  index?: number;
  field?: string;
};

export class YAMLDimension {
  column: string;
  expression: string;
  name: string;
  label: string;
  description: string;
  unnest: boolean | undefined;
  resourceName: string;

  constructor(
    item?: YAMLMap<string, string>,
    dimension?: MetricsViewSpecDimensionV2,
  ) {
    this.column = item?.get("column") ?? "";
    this.expression = item?.get("expression") ?? "";
    this.name = item?.get("name") ?? "";
    this.label = item?.get("label") ?? "";
    this.description = item?.get("description") ?? "";
    this.unnest =
      item?.get("unnest") === undefined
        ? undefined
        : item?.get("unnest") === "true";
    this.resourceName = dimension?.name ?? "";
  }
}

export class YAMLMeasure {
  expression: string;
  name: string;
  label: string;
  description: string;
  valid_percent_of_total: boolean;
  format_d3: string;
  format_preset: FormatPreset | "";

  constructor(item?: YAMLMap<string, string>) {
    this.expression = item?.get("expression") ?? "";
    this.name = item?.get("name") ?? "";
    this.label = item?.get("label") ?? item?.get("name") ?? "";
    this.description = item?.get("description") ?? "";
    this.valid_percent_of_total =
      item?.get("valid_percent_of_total") === undefined
        ? true
        : Boolean(item?.get("valid_percent_of_total"));
    this.format_d3 = item?.get("format_d3") ?? "";
    this.format_preset =
      (item?.get("format_preset") as unknown as FormatPreset) ?? this.format_d3
        ? ""
        : FormatPreset.HUMANIZE;
  }
}

export const ROW_HEIGHT = 40;
