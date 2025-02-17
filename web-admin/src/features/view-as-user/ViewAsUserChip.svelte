<script lang="ts">
  import { page } from "$app/stores";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { errorStore } from "../../components/errors/error-store";
  import ViewAsUserPopover from "./ViewAsUserPopover.svelte";
  import { viewAsUserStore } from "./viewAsUserStore";

  let active: boolean;
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger asChild let:builder>
    <Chip
      removable
      {active}
      builders={[builder]}
      removeTooltipText="Clear view"
      on:remove={() => {
        viewAsUserStore.set(null);
        errorStore.reset();
      }}
    >
      <div slot="body">
        Viewing as <b>{$viewAsUserStore.email}</b>
      </div>
    </Chip>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="start"
    class="flex flex-col min-w-[150px] max-w-[300px] min-h-[150px] max-h-[190px]"
  >
    <ViewAsUserPopover
      organization={$page.params.organization}
      project={$page.params.project}
      on:select={() => (active = false)}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
