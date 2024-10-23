import { cubicOut } from "svelte/easing";
import type { TransitionConfig } from "svelte/transition";

export default function slideRight(
  node: Element,
  {
    delay = 0,
    duration = 200,
    easing = cubicOut,
    rightOffset = 0,
    leftOffset = 0,
  },
): TransitionConfig {
  const style = getComputedStyle(node);
  const width = parseFloat(style.width);
  const opacity = +style.opacity;
  const paddingLeft = parseFloat(style.paddingLeft);
  const paddingRight = parseFloat(style.paddingRight);
  const marginLeft = parseFloat(style.marginTop);
  const marginRight = parseFloat(style.marginBottom);
  const borderLeftWidth = parseFloat(style.borderLeftWidth);
  const borderRightWidth = parseFloat(style.borderRightWidth);

  return {
    delay,
    duration,
    easing,
    css: (t) => `
      overflow: hidden;
      white-space: nowrap;
      opacity: ${Math.min(t * 20, 1) * opacity};
      width: ${t * width}px;
      padding-right: ${t * paddingRight}px;
      padding-left: ${t * paddingLeft}px;
      margin-left: ${t * marginLeft - (1 - t) * leftOffset}px;
      margin-right: ${t * marginRight - (1 - t) * rightOffset}px;
      border-left-width: ${t * borderLeftWidth}px;
      border-right-width: ${t * borderRightWidth}px;
    `,
  };
}
