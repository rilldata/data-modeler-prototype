<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { PlusCircle } from "lucide-svelte";
  import type { ComponentType, SvelteComponent } from "svelte";
  import type { CanvasComponentType } from "./components/types";
  import BigNumberIcon from "./icons/BigNumberIcon.svelte";
  import ChartIcon from "./icons/ChartIcon.svelte";
  import TableIcon from "./icons/TableIcon.svelte";
  import TextIcon from "./icons/TextIcon.svelte";

  type MenuItem = {
    id: CanvasComponentType;
    label: string;
    icon: ComponentType<SvelteComponent>;
  };

  export const menuItems: MenuItem[] = [
    { id: "bar_chart", label: "Chart", icon: ChartIcon },
    { id: "table", label: "Table", icon: TableIcon },
    { id: "markdown", label: "Text", icon: TextIcon },
    { id: "kpi_grid", label: "KPI", icon: BigNumberIcon },
    { id: "image", label: "Image", icon: ChartIcon },
  ];

  export let disabled = false;
  export let componentForm = false;
  export let onItemClick: (type: CanvasComponentType) => void;
  export let onMouseEnter: () => void;

  let open = false;
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    {#if componentForm}
      <button
        {...builder}
        use:builder.action
        class="pointer-events-auto shadow-sm hover:shadow-md flex bg-white h-[84px] flex-col justify-center gap-2 items-center rounded-md border border-slate-200 w-full"
      >
        <PlusCircle class="w-6 h-6 text-slate-500" />
        <span class="text-sm font-medium text-slate-500">Add a component</span>
      </button>
    {:else}
      <button
        {disabled}
        on:mouseenter={onMouseEnter}
        use:builder.action
        class="pointer-events-auto disabled:pointer-events-none h-7 px-2 grid place-content-center z-50 hover:bg-gray-100 text-slate-500 disabled:opacity-50"
      >
        <PlusCircle size="15px" />
      </button>
    {/if}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align={componentForm ? "center" : "start"}>
    <div class="flex flex-col" role="presentation" on:mouseenter={onMouseEnter}>
      {#each menuItems as { id, label, icon } (id)}
        <DropdownMenu.Item
          on:click={() => {
            open = false;
            onItemClick(id);
          }}
        >
          <div class="flex flex-row gap-x-2">
            <svelte:component this={icon} />
            {label}
          </div>
        </DropdownMenu.Item>
      {/each}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
