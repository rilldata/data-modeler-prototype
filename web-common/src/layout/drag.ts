import type { Writable } from "svelte/store";
import type { LayoutElement } from "./workspace/types";

interface DragParams {
  store: Writable<LayoutElement>;
  minSize?: number;
  maxSize?: number;
  reverse?: boolean;
  orientation?: "horizontal" | "vertical";
}

export function drag(node, params: DragParams) {
  const underlyingStore = params.store;
  const minSize_ = params?.minSize || 300;
  const maxSize_ = params?.maxSize || 440;
  const reverse_ = params?.reverse || false;
  const orientation_ = params?.orientation || "horizontal";

  let moving = false;
  let space = minSize_;

  node.style.cursor = "move";
  node.style.userSelect = "none";

  function mousedown() {
    moving = true;
  }

  function mousemove(e: MouseEvent) {
    e.preventDefault();

    if (moving) {
      let size;
      if (orientation_ === "horizontal") {
        size = reverse_ ? innerWidth - e.pageX : e.pageX;
      } else if (orientation_ === "vertical") {
        size = reverse_ ? innerHeight - e.pageY : e.pageY;
      }
      if (size > minSize_ && size < maxSize_) {
        space = size;
      }
      /** update the store passed in as a parameter */
      underlyingStore.update((state) => {
        state.value = space;
        return state;
      });
    }
  }

  function mouseup() {
    moving = false;
  }

  node.addEventListener("mousedown", mousedown);
  window.addEventListener("mousemove", mousemove);
  window.addEventListener("mouseup", mouseup);
  return {
    update() {
      moving = false;
    },
  };
}
