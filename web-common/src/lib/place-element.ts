import { clamp } from "@rilldata/utils";

export type Location = "left" | "right" | "top" | "bottom";
export type Alignment = "start" | "middle" | "end";
export type FloatingElementRelationship = "parent" | "direct" | "mouse";

export function mouseLocationToBoundingRect({
  x,
  y,
  width = 0,
  height = 0,
}: {
  x: number;
  y: number;
  width?: number;
  height?: number;
}) {
  return new DOMRect(x, y, width, height);
}

export function placeElement({
  location,
  alignment,
  parentPosition, // using getBoundingClientRect // DOMRect
  elementPosition, // using getBoundingClientRect // DOMRect
  distance = 0,
  x = 0,
  y = 0,
  windowWidth = window.innerWidth,
  windowHeight = window.innerHeight,
  pad = 16 * 2,
  overflowFlipY = true,
}: {
  location: Location;
  alignment: Alignment;
  parentPosition: DOMRect;
  elementPosition: DOMRect;
  distance?: number;
  x?: number;
  y?: number;
  windowWidth?: number;
  windowHeight?: number;
  pad?: number;
  overflowFlipY?: boolean;
}) {
  let left = 0;
  let top = 0;

  const elementWidth = elementPosition.width;
  const elementHeight = elementPosition.height;

  const parentRight = parentPosition.right + x;
  const parentLeft = parentPosition.left + x;
  const parentTop = parentPosition.top + y;
  const parentBottom = parentPosition.bottom + y;
  const parentWidth = parentPosition.width;
  const parentHeight = parentPosition.height;

  // Task 1: check if we need to reflect agains the location axis.
  if (location === "bottom") {
    if (
      overflowFlipY &&
      parentBottom + elementHeight + distance + pad > windowHeight + y
    ) {
      top = parentTop - elementHeight - distance;
    } else {
      top = parentBottom + distance;
    }
  } else if (location === "top") {
    if (overflowFlipY && parentTop - elementHeight - distance - pad < y) {
      top = parentBottom + distance;
    } else {
      top = parentTop - elementHeight - distance;
    }
  } else if (location === "left") {
    if (parentLeft - distance - elementWidth - pad < x) {
      // reflect
      left = parentRight + distance;
    } else {
      left = parentLeft - elementWidth - distance;
    }
  } else if (location === "right") {
    if (parentRight + elementWidth + distance + pad > windowWidth + x) {
      left = parentLeft - elementWidth - distance;
    } else {
      left = parentRight + distance;
    }
  }

  // OUR SECOND JOB IS RE-ALIGNMENT ALONG THE ALIGNMENT ACTION.
  let alignmentValue: number;

  const rightLeft = location === "right" || location === "left";

  switch (alignment) {
    case "start": {
      alignmentValue = rightLeft
        ? parentTop // right / left
        : parentLeft; // top / bottom
      break;
    }
    case "end": {
      alignmentValue = rightLeft
        ? parentBottom - elementHeight // right / left
        : parentRight - elementWidth; // top / bottom
      break;
    }
    default: {
      // 'middle'
      alignmentValue = rightLeft
        ? parentTop - (elementHeight - parentHeight) / 2 // right / left
        : parentLeft - (elementWidth - parentWidth) / 2; // top / bottom
      break;
    }
  }
  const alignMin = pad + (rightLeft ? y : x);
  const alignMax = rightLeft
    ? y + windowHeight - elementHeight - pad
    : x + windowWidth - elementWidth - pad;

  const value = clamp(alignMin, alignmentValue, alignMax);

  if (rightLeft) {
    top = value;
  } else {
    left = value;
  }

  return [left, top];
}
