<script lang="ts">
  import { cn, flyAndScale } from "@rilldata/web-common/lib/shadcn";
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";
  import Cross2 from "svelte-radix/Cross2.svelte";
  import * as AlertDialog from "web-common/src/components/alert-dialog/index.js";

  type $$Props = AlertDialogPrimitive.ContentProps;

  let className: $$Props["class"] = undefined;
  export let transition: $$Props["transition"] = flyAndScale;
  export let transitionConfig: $$Props["transitionConfig"] = undefined;
  export let noCancel = false;
  export { className as class };
</script>

<AlertDialog.Portal>
  <AlertDialog.Overlay />
  <AlertDialogPrimitive.Content
    {transition}
    {transitionConfig}
    class={cn(
      "fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border border-slate-300 bg-surface p-6 shadow-lg sm:rounded-lg md:w-full",
      className,
    )}
    {...$$restProps}
  >
    <slot />
    {#if !noCancel}
      <AlertDialogPrimitive.Cancel
        class="absolute right-4 top-4 rounded-md opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
      >
        <Cross2 class="h-4 w-4" />
        <span class="sr-only">Close</span>
      </AlertDialogPrimitive.Cancel>
    {/if}
  </AlertDialogPrimitive.Content>
</AlertDialog.Portal>
