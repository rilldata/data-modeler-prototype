<!-- @component 
The FloatingElement component is the backbone of all of our floating UI element functionality.
It handles the setting of the location of the floating element relative to these possible options, set in the relationship prop:
- a direct DOM element passed in through target through the 'direct' prop
- the first child of target through the "parent" prop
display:contents. This is useful when nesting a floating element within a tooltip.
- a mouse click location through "mouse". This is an {x,y} coordinate that matches where the pointer is.
-->
<script lang="ts">
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import {
    mouseLocationToBoundingRect,
    placeElement,
  } from "../../lib/place-element";
  import type {
    Location,
    Alignment,
    FloatingElementRelationship,
  } from "../../lib/place-element";
  import { portal } from "@rilldata/actions";

  export let target: HTMLElement;
  export let relationship: FloatingElementRelationship = "parent"; // parent, mouse {x, y}
  export let location: Location = "bottom";
  export let alignment: Alignment = "middle";
  export let distance = 0;
  // edge padding
  export let pad = 8;
  // whether to flip the element's location (from bottom to top) or (from top to bottom)
  // if it overflows the window
  export let overflowFlipY = true;
  // mouse position to be used when relationship is `mouse`
  export let mousePos = { x: 0, y: 0 };

  let top = 0;
  let left = 0;
  let innerHeight: number;
  let innerWidth: number;
  let scrollY: number;
  let scrollX: number;
  let child: Element | undefined;
  let firstParentElement: Element | undefined;

  $: if (firstParentElement && child) {
    if (relationship === "parent") {
      setLocation(
        firstParentElement.getBoundingClientRect(),
        child.getBoundingClientRect(),
        scrollX,
        scrollY,
        innerWidth,
        innerHeight,
        overflowFlipY,
      );
    } else if (relationship === "direct") {
      setLocation(
        target.getBoundingClientRect(),
        child.getBoundingClientRect(),
        scrollX,
        scrollY,
        innerWidth,
        innerHeight,
        overflowFlipY,
      );
    } else {
      setLocation(
        mouseLocationToBoundingRect({ x: mousePos.x, y: mousePos.y }),
        child?.getBoundingClientRect(),
        scrollX,
        scrollY,
        innerWidth,
        innerHeight,
        overflowFlipY,
      );
    }
  }
  $: getFirstValidChildElement(target);

  onMount(() => {
    // we listen to the parent.
    // actually, we listen to the first chidl element!
    if (relationship === "parent") {
      firstParentElement = getFirstValidChildElement(target); // target?.children[0];
      const config = { attributes: true };
      const observer = new MutationObserver(() => {
        if (child && firstParentElement)
          setLocation(
            firstParentElement?.getBoundingClientRect(),
            child?.getBoundingClientRect(),
            scrollX,
            scrollY,
            innerWidth,
            innerHeight,
            overflowFlipY,
          );
      });
      if (firstParentElement) {
        observer.observe(firstParentElement, config);
      }
    }
  });

  function getFirstValidChildElement(element: Element) {
    // get this child.
    let possibleChild = element?.children[0];
    if (!(possibleChild instanceof Element)) return;
    // check for display: contents, which may indicate
    // another wrapped object.
    if (getComputedStyle(possibleChild).display === "contents") {
      return getFirstValidChildElement(possibleChild);
    } else {
      return possibleChild;
    }
  }

  function setLocation(
    parentBoundingClientRect: DOMRect,
    elementBoundingClientRect: DOMRect,
    scrollXValue: number,
    scrollYvalue: number,
    windowWidth: number,
    windowHeight: number,
    overflowFlipY: boolean,
  ) {
    if (!(parentBoundingClientRect && elementBoundingClientRect)) return;
    const [leftPos, topPos] = placeElement({
      location,
      alignment,
      distance,
      pad,
      parentPosition: parentBoundingClientRect,
      elementPosition: elementBoundingClientRect,
      y: scrollYvalue,
      x: scrollXValue,
      windowWidth,
      windowHeight,
      overflowFlipY,
    });
    top = topPos;
    left = leftPos;
  }
</script>

<svelte:window bind:scrollX bind:scrollY bind:innerHeight bind:innerWidth />

<div
  use:portal
  transition:fade={{ duration: 25 }}
  bind:this={child}
  class="absolute"
  style:z-index="200"
  style:left="{left}px"
  style:top="{top}px"
>
  <slot />
</div>
