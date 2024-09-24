<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";

  type $$Props = DropdownMenuPrimitive.LabelProps & {
    inset?: boolean;
  };

  let className: $$Props["class"] = undefined;
  export let inset: $$Props["inset"] = undefined;
  export { className as class };
</script>

<DropdownMenuPrimitive.Label
  class={cn(
    "px-2 py-1.5 text-xs text-gray-500 font-semibold",
    inset && "pl-8",
    className,
  )}
  {...$$restProps}
>
  <slot />
</DropdownMenuPrimitive.Label>
