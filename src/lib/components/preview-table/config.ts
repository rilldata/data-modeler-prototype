import type { VirtualizedTableConfig } from "$lib/components/virtualized-table/types";

export const config: VirtualizedTableConfig = {
  defaultColumnWidth: 200,
  maxColumnWidth: 320,
  minColumnWidth: 120,
  minHeaderWidthWhenColumsAreSmall: 160,
  rowHeight: 36,
  columnHeaderHeight: 32,
  indexWidth: 60,
  columnHeaderFontWeightClass: "font-bold",
  defaultFontWeightClass: "font-semibold",
  table: "PreviewTable",
};
