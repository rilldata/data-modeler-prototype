export interface HeaderData<T> {
  value: T | null | undefined;
  spark?: string;
}

export type TDDCellData = string | number | null | undefined;

export interface TableData {
  rowCount: number;
  fixedColCount: number;
  rowHeaderData: HeaderData<string>[][];
  columnCount: number;
  columnHeaderData: HeaderData<Date>[][];
  body: TDDCellData[][];
  selectedValues: (string | null)[];
}

export interface HighlightedCell {
  dimensionValue: string | undefined | null;
  time: Date | undefined;
}

export interface ChartInteractionColumns {
  hover: number | undefined;
  scrubStart: number | undefined;
  scrubEnd: number | undefined;
}

export type TDDComparison = "time" | "none" | "dimension";

export interface TablePosition {
  x0?: number;
  x1?: number;
  y0?: number;
  y1?: number;
}
