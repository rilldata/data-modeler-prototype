export function clickOutside(
  node: Node,
  [trigger, cb]: [Node | null, (event: Event) => void],
) {
  function handleClick(e: MouseEvent) {
    if (!(e.target instanceof Node)) return;
    if (node === e.target) return;
    if (node.contains(e.target)) return;
    if (trigger !== null) {
      if (!trigger) return;
      if (trigger.contains(e.target)) return;
    }

    cb(e);
  }
  document.addEventListener("click", handleClick);
  return {
    destroy() {
      document.removeEventListener("click", handleClick);
    },
  };
}
