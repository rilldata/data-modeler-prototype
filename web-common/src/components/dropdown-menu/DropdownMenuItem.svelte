<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";

  type ItemType = "default" | "destructive";

  type $$Props = DropdownMenuPrimitive.ItemProps & {
    inset?: boolean;
    type?: ItemType;
  };
  // type $$Events = DropdownMenuPrimitive.ItemEvents;

  let className: $$Props["class"] = undefined;
  export { className as class };

  export let href: $$Props["href"] = undefined;
  export let inset: $$Props["inset"] = undefined;
  export let type: $$Props["type"] = "default";
</script>

<DropdownMenuPrimitive.Item
  class={cn(
    "relative flex gap-x-2 cursor-default select-none items-center rounded-sm px-2 py-1.5 text-xs outline-none data-[highlighted]:bg-slate-100 data-[highlighted]:text-accent-foreground data-[highlighted]:cursor-pointer data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
    href && "text-gray-900 font-normal",
    inset && "pl-8",
    type === "destructive" &&
      "text-red-500 hover:text-red-600 data-[highlighted]:text-red-600",
    className,
  )}
  {...$$restProps}
  {href}
  on:click
  on:keydown
  on:focusin
  on:focusout
  on:pointerdown
  on:pointerleave
  on:pointermove
>
  <slot />
</DropdownMenuPrimitive.Item>
