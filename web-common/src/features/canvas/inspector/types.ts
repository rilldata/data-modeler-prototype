type NativeInputTypes = "text" | "number" | "boolean" | "textArea";
type SemanticInputTypes =
  | "metrics"
  | "measure"
  | "dimension"
  | "multi_measures"
  | "multi_dimensions";
type ChartInputTypes = "positional" | "mark" | "tooltip" | "config";
type CustomInputTypes = "rill_time" | "sparkline" | "comparison_options";
type PositionalInputTypes = "alignment";

export type InputType =
  | NativeInputTypes
  | SemanticInputTypes
  | ChartInputTypes
  | CustomInputTypes
  | PositionalInputTypes;

export type FilterInputTypes = "time_filters" | "dimension_filters";

export interface ComponentInputParam {
  type: InputType;
  label?: string;
  showInUI?: boolean; // If not specified, can assume true
  optional?: boolean;
  description?: string; // Tooltip description for the input
  meta?: Record<string, string>; // Any additional metadata
}

export interface FilterInputParam {
  type: FilterInputTypes;
  meta?: Record<string, any>;
}

export interface InputParams<T> {
  options: Partial<Record<keyof T, ComponentInputParam>>;
  filter: Partial<Record<FilterInputTypes, FilterInputParam>> | [];
}
