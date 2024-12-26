import * as defaults from "./constants";
import type { PositionedItem, RowGroup, Vector, GridItem } from "./types";
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

// Ensure items are within the grid and not overlapping
export function validateItemPositions(items: V1CanvasItem[]): void {
  // First group items by row
  const rows = groupItemsByRow(items);

  // Process each row
  rows.forEach((row) => {
    leftAlignRow(row);
  });

  // Validate x positions are within bounds
  items.forEach((item) => {
    if (item.x !== undefined && item.width !== undefined) {
      item.x = Math.min(
        Math.max(0, item.x),
        defaults.COLUMN_COUNT - item.width,
      );
    }
  });
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

export function convertToGridItems(yamlItems: any[]): GridItem[] {
  return yamlItems.map((item) => ({
    position: [item.get("x"), item.get("y")],
    size: [item.get("width"), item.get("height")],
    node: item,
  }));
}

export function sortItemsByPosition(items: GridItem[]): GridItem[] {
  return items.sort((a, b) => {
    // Sort by Y first, then X for items in the same row
    if (a.position[1] === b.position[1]) {
      return a.position[0] - b.position[0];
    }
    return a.position[1] - b.position[1];
  });
}

export function leftAlignRow(row: RowGroup) {
  const startPosition: Vector = [0, row.y];

  row.items
    .sort((a, b) => (a.x ?? 0) - (b.x ?? 0))
    .forEach((item) => {
      if (item.x !== undefined) {
        item.x = startPosition[0];
        startPosition[0] = vector.add(
          [startPosition[0], 0],
          [item.width ?? 0, 0],
        )[0];
      }
    });
}

export function compactGrid(items: GridItem[]) {
  let currentY = 0;
  let lastRowHeight = 0;
  let lastY = -1;
  let currentRowItems: GridItem[] = [];

  // Process all items
  items.forEach((item, index) => {
    if (item.position[1] !== lastY) {
      // When we hit a new row, process the previous row
      if (currentRowItems.length > 0) {
        // Sort by X position and compact
        currentRowItems.sort((a, b) => a.position[0] - b.position[0]);
        let currentX = 0;
        currentRowItems.forEach((rowItem) => {
          rowItem.node.set("x", currentX);
          currentX += rowItem.size[0];
        });
      }

      // Start new row
      currentY += lastRowHeight;
      lastRowHeight = item.size[1];
      lastY = item.position[1];
      currentRowItems = [item];
    } else {
      // Same row - update max height if needed
      lastRowHeight = Math.max(lastRowHeight, item.size[1]);
      currentRowItems.push(item);
    }

    // Update Y position
    item.node.set("y", currentY);
  });

  // Process the last row
  if (currentRowItems.length > 0) {
    currentRowItems.sort((a, b) => a.position[0] - b.position[0]);
    let currentX = 0;
    currentRowItems.forEach((rowItem) => {
      rowItem.node.set("x", currentX);
      currentX += rowItem.size[0];
    });
  }
}

export function getRowIndex(item: V1CanvasItem, items: V1CanvasItem[]): number {
  const rows = groupItemsByRow(items);
  return rows.findIndex((row) =>
    row.items.some((rowItem) => rowItem.x === item.x && rowItem.y === item.y),
  );
}

export function getColumnIndex(
  item: V1CanvasItem,
  items: V1CanvasItem[],
): number {
  const rows = groupItemsByRow(items);
  const row = rows.find((r) =>
    r.items.some((rowItem) => rowItem.x === item.x && rowItem.y === item.y),
  );

  if (!row) return 0;

  // Sort items in the row by x position and find index
  const sortedItems = [...row.items].sort((a, b) => (a.x ?? 0) - (b.x ?? 0));
  return sortedItems.findIndex(
    (rowItem) => rowItem.x === item.x && rowItem.y === item.y,
  );
}
