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

export function moveItemToNewRow(
  item: V1CanvasItem,
  targetY: number,
  targetHeight: number,
): void {
  if (!isValidItem(item)) return;
  item.y = targetY + targetHeight;
  item.x = 0;
}

export function isValidItem(item: V1CanvasItem): item is V1CanvasItem & {
  x: number;
  y: number;
  width: number;
  height: number;
} {
  return (
    item.x !== undefined &&
    item.y !== undefined &&
    item.width !== undefined &&
    item.height !== undefined
  );
}

export function recalculateRowPositions(
  items: V1CanvasItem[],
  targetY: number | undefined,
): void {
  let currentX = 0;

  items.forEach((item, index) => {
    if (!isValidItem(item)) return;
    if (item.y !== targetY) return;

    if (
      index === 0 ||
      !isValidItem(items[index - 1]) ||
      items[index - 1].y !== item.y
    ) {
      item.x = 0;
      currentX = item.width;
    } else {
      const newX = Math.round(currentX + defaults.GAP_SIZE / 1000);

      if (newX + item.width > defaults.COLUMN_COUNT) {
        moveItemToNewRow(item, targetY, item.height);
        currentX = item.width;
      } else {
        item.x = Math.min(newX, defaults.COLUMN_COUNT - item.width);
        currentX = item.x + item.width;
      }
    }
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
