const SAMPLE_SIZE = 30;
import { clamp } from "@rilldata/utils";
import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import { type V1MetricsViewColumn } from "@rilldata/web-common/runtime-client";

export function extractSamples<T>(arr: T[], sampleSize: number = 30) {
  if (arr.length <= sampleSize) {
    return arr.slice();
  }

  const sectionSize = Math.floor(sampleSize / 3);

  const lastSectionSize = sampleSize - sectionSize * 2;

  const first = arr.slice(0, sectionSize);

  const middleStartIndex = Math.floor((arr.length - sectionSize) / 2);
  const middle = arr.slice(middleStartIndex, middleStartIndex + sectionSize);

  const last = arr.slice(-lastSectionSize);

  return [...first, ...middle, ...last];
}

export function initColumnWidths<K>(params: {
  columns: (VirtualizedTableColumns | V1MetricsViewColumn)[];
  rows: K[];
  columnAccessor: keyof VirtualizedTableColumns;
  minWidth: number;
  maxWidth: number;
  resizableColumns: boolean;
}) {
  const {
    columns,
    rows,
    columnAccessor,
    minWidth,
    maxWidth,
    resizableColumns,
  } = params;
  const samples = extractSamples(rows, SAMPLE_SIZE);

  return columns.map((column) => {
    if (!resizableColumns) return minWidth;

    const columnName = String(column[columnAccessor]);

    const columnNameLength = columnName.length ?? 8;

    let totalValueCharacters = 0;

    samples.forEach((row) => {
      totalValueCharacters += String(row[columnName]).length;
    });

    const averageCharacterLength = Math.ceil(
      totalValueCharacters / SAMPLE_SIZE,
    );

    const lengthBasis = Math.max(columnNameLength, averageCharacterLength);

    const pixelLength = Math.ceil(lengthBasis * 7);

    if (columnNameLength > averageCharacterLength) {
      return clamp(minWidth, pixelLength + 60, maxWidth);
    } else {
      return clamp(minWidth, pixelLength + 60, maxWidth);
    }
  });
}
