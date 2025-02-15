import type { ComponentType, SvelteComponent } from "svelte";

export type ChartSortDirection = "x" | "y" | "-x" | "-y";

export interface FieldConfig {
  field: string;
  showAxisTitle?: boolean; // Default is false
  zeroBasedOrigin?: boolean; // Default is false
  type: "quantitative" | "ordinal" | "nominal" | "temporal";
  timeUnit?: string; // For temporal fields
  sort?: ChartSortDirection;
}

export interface ChartConfig {
  metrics_view: string;
  x?: FieldConfig;
  y?: FieldConfig;
  color?: FieldConfig | string;
  tooltip?: FieldConfig;
  vl_config?: string;
}

export type ChartType =
  | "line_chart"
  | "bar_chart"
  | "stacked_bar"
  | "area_chart";

export interface ChartMetadata {
  type: ChartType;
  icon: ComponentType<SvelteComponent>;
  title: string;
}

/** Temporary solution for the lack of vega lite type exports */
export interface TooltipValue {
  title?: string;
  field: string;
  format?: string;
  formatType?: string;
  type: "quantitative" | "ordinal" | "temporal" | "nominal";
}
