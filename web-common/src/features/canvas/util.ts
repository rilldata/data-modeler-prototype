import * as defaults from "./constants";
import type { PositionedItem, Vector } from "./types";
import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";

export const vector = {
  add: (add: Vector, initial: Vector): Vector => {
    return [add[0] + initial[0], add[1] + initial[1]];
  },
  multiply: (vector: Vector, multiplier: Vector): Vector => {
    return [vector[0] * multiplier[0], vector[1] * multiplier[1]];
  },
  subtract: (minuend: Vector, subtrahend: Vector): Vector => {
    return [minuend[0] - subtrahend[0], minuend[1] - subtrahend[1]];
  },
  absolute: (vector: Vector): Vector => {
    return [Math.abs(vector[0]), Math.abs(vector[1])];
  },
  divide: (vector: Vector, divisor: Vector): Vector => {
    return [vector[0] / divisor[0], vector[1] / divisor[1]];
  },
};

export function isString(value: unknown): value is string {
  return typeof value === "string";
}

// Allowed widths for components
const ALLOWED_WIDTHS = [3, 4, 6, 8, 9, 12];

// Snap to the closest valid width
function getValidWidth(newWidth: number): number {
  return ALLOWED_WIDTHS.reduce((closest, width) =>
    Math.abs(width - newWidth) < Math.abs(closest - newWidth) ? width : closest,
  );
}

// Check if a position is free of collisions
function isPositionFree(
  existingItems: PositionedItem[],
  x: number,
  y: number,
  width: number,
  height: number,
): boolean {
  return !existingItems.some((item) => {
    const overlapsInX = x < item.x + item.width && x + width > item.x;
    const overlapsInY = y < item.y + item.height && y + height > item.y;
    return overlapsInX && overlapsInY;
  });
}

// Row-based grouping with sequential placement with collision checks
export function findNextAvailablePosition(
  existingItems: PositionedItem[],
  newWidth: number,
  newHeight: number,
): [number, number] {
  const validWidth = getValidWidth(newWidth);

  if (!existingItems?.length) {
    return [0, 0];
  }

  // Group items by row (y coordinate)
  const rowGroups = new Map<number, PositionedItem[]>();
  existingItems.forEach((item) => {
    const items = rowGroups.get(item.y) || [];
    items.push(item);
    rowGroups.set(item.y, items);
  });

  // Sort rows top-to-bottom
  const rows = Array.from(rowGroups.entries()).sort(([y1], [y2]) => y1 - y2);

  // First pass: find space at the end of rows
  for (const [y, items] of rows) {
    const rightmostX = Math.max(...items.map((item) => item.x + item.width), 0);
    if (rightmostX + validWidth <= defaults.COLUMN_COUNT) {
      if (isPositionFree(existingItems, rightmostX, y, validWidth, newHeight)) {
        return [rightmostX, y];
      }
    }
  }

  // Second pass: find gaps within rows
  for (const [y, items] of rows) {
    const sortedItems = items.sort((a, b) => a.x - b.x);

    let x = 0;
    for (const item of sortedItems) {
      if (
        x + validWidth <= item.x &&
        isPositionFree(existingItems, x, y, validWidth, newHeight)
      ) {
        return [x, y];
      }
      x = item.x + item.width;
    }

    // Check after the last item in the row
    if (
      x + validWidth <= defaults.COLUMN_COUNT &&
      isPositionFree(existingItems, x, y, validWidth, newHeight)
    ) {
      return [x, y];
    }
  }

  // Final pass: add a new row
  const lastRowY = Math.max(
    ...existingItems.map((item) => item.y + item.height),
    0,
  );
  const newY = lastRowY; // Place the new row below the tallest existing item
  return [0, newY];
}

export function isValidItem(item: V1CanvasItem): item is V1CanvasItem & {
  x: number;
  y: number;
  width: number;
  height: number;
} {
  return (
    item?.x !== undefined &&
    item?.y !== undefined &&
    item?.width !== undefined &&
    item?.height !== undefined
  );
}

export function recalculateRowPositions(
  items: V1CanvasItem[],
  startingY: number,
) {
  const rows = groupItemsByRow(items);
  let currentY = startingY;

  rows.forEach((row) => {
    // Sort items in row by x position
    row.items.sort((a, b) => (a.x ?? 0) - (b.x ?? 0));

    // Adjust x positions within row
    let currentX = 0;
    row.items.forEach((item) => {
      item.y = currentY;
      item.x = currentX;
      currentX += (item?.width ?? 0) + Math.round(defaults.GAP_SIZE / 1000);

      // Ensure item doesn't exceed grid width
      if (currentX > defaults.COLUMN_COUNT) {
        item.y = currentY + row.height;
        item.x = 0;
        currentX = (item?.width ?? 0) + Math.round(defaults.GAP_SIZE / 1000);
      }
    });
    currentY += row.height;
  });
}

export function validateItemPositions(items: V1CanvasItem[]): void {
  items.forEach((item) => {
    if (item.x !== undefined && item.width !== undefined) {
      item.x = Math.min(
        Math.max(0, item.x),
        defaults.COLUMN_COUNT - item.width,
      );
    }
  });
}

export function reorderRows(
  items: V1CanvasItem[],
  sourceY: number | undefined,
  targetY: number,
): void {
  if (sourceY === undefined || sourceY === targetY) return;

  // First, identify rows and their items
  const rowMap = new Map<number, V1CanvasItem[]>();
  items.forEach((item) => {
    if (!isValidItem(item)) return;
    const row = rowMap.get(item.y) || [];
    row.push(item);
    rowMap.set(item.y, row);
  });

  // Get source and target rows
  const sourceRow = rowMap.get(sourceY) || [];
  const targetRow = rowMap.get(targetY) || [];

  // Simple row swap - update y values
  sourceRow.forEach((item) => {
    if (!isValidItem(item)) return;
    item.y = targetY;
  });

  targetRow.forEach((item) => {
    if (!isValidItem(item)) return;
    item.y = sourceY;
  });

  // If source row has a full-width item, ensure it stays full width
  const hasFullWidth = sourceRow.some(
    (item) => isValidItem(item) && item.width === defaults.COLUMN_COUNT,
  );
  if (hasFullWidth) {
    sourceRow.forEach((item) => {
      if (!isValidItem(item)) return;
      item.width = defaults.COLUMN_COUNT;
      item.x = 0;
    });
  }

  // Recalculate x positions for non-full-width rows
  if (!hasFullWidth) {
    recalculateRowPositions(items, targetY);
  }
  if (
    !targetRow.some(
      (item) => isValidItem(item) && item.width === defaults.COLUMN_COUNT,
    )
  ) {
    recalculateRowPositions(items, sourceY);
  }
}

interface RowGroup {
  y: number;
  height: number;
  items: V1CanvasItem[];
}

export function groupItemsByRow(items: V1CanvasItem[]): RowGroup[] {
  const rows: RowGroup[] = [];

  items.forEach((item) => {
    const existingRow = rows.find((row) => row.y === item.y);
    if (existingRow) {
      existingRow.items.push(item);
      existingRow.height = Math.max(existingRow.height ?? 0, item.height ?? 0);
    } else {
      rows.push({
        y: item.y ?? 0,
        height: item.height ?? 0,
        items: [item],
      });
    }
  });

  return rows.sort((a, b) => a.y - b.y);
}

export function flattenRowGroups(rows: RowGroup[]): V1CanvasItem[] {
  return rows.flatMap((row) => row.items);
}
