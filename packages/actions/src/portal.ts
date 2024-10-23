import { tick } from "svelte";

type Target = HTMLElement | string;

export function portal(node: HTMLElement, target: Target = "#rill-portal") {
  let targetElement: HTMLElement;

  async function update(newTarget: Target) {
    let possibleTarget: HTMLElement | null;

    if (typeof newTarget === "string") {
      possibleTarget = document.querySelector(newTarget);
      if (possibleTarget === null) {
        await tick();
        possibleTarget = document.querySelector(newTarget);
      }
      if (possibleTarget === null) {
        possibleTarget = document.body;
      }
      targetElement = possibleTarget;
    } else {
      const window = newTarget.ownerDocument?.defaultView || false;
      if (window && newTarget instanceof window.HTMLElement) {
        targetElement = newTarget;
      } else {
        throw new TypeError(`Unknown portal type.`);
      }
    }

    targetElement.appendChild(node);
    node.hidden = false;
  }

  function destroy() {
    if (node.parentNode) {
      node.parentNode.removeChild(node);
    }
  }

  void update(target);
  return {
    update,
    destroy,
  };
}
