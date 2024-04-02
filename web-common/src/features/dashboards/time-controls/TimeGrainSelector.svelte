<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
  import type {
    AvailableTimeGrain,
    TimeGrain,
  } from "@rilldata/web-common/lib/time/types";

  import type { V1TimeGrain } from "../../../runtime-client";
  import { page } from "$app/stores";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";

  export let timeGrainOptions: TimeGrain[];
  export let minTimeGrain: V1TimeGrain;

  $: searchParams = $page.url.searchParams;

  let open = false;

  $: activeTimeGrain = searchParams.get("timeGrain") ?? "TIME_GRAIN_DAY";
  $: activeTimeGrainLabel =
    activeTimeGrain && TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label;

  $: timeGrains = timeGrainOptions
    .filter((timeGrain) => !isGrainBigger(minTimeGrain, timeGrain.grain))
    .map((timeGrain) => {
      return {
        main: timeGrain.label,
        key: timeGrain.grain,
      };
    });

  $: createLink = (key: string, value: string) => {
    const newParams = new URLSearchParams(searchParams);
    newParams.set(key, value);
    return `?${newParams.toString()}`;
  };
</script>

{#if timeGrainOptions.length}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger asChild let:builder>
      <button
        use:builder.action
        {...builder}
        aria-label="Select a time grain"
        class:bg-gray-200={open}
        class="px-3 py-2 rounded flex flex-row gap-x-2 items-center hover:bg-gray-200 hover:dark:bg-gray-600"
      >
        <div>
          Metric trends by <span class="font-bold">{activeTimeGrainLabel}</span>
        </div>
        <IconSpaceFixer pullRight>
          <div class="transition-transform" class:-rotate-180={open}>
            <CaretDownIcon size="14px" />
          </div>
        </IconSpaceFixer>
      </button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="min-w-40 flex flex-col" align="start">
      {#each timeGrains as option}
        <DropdownMenu.Item
          href={createLink("timeGrain", option.key)}
          role="menuitem"
          class="text-xs cursor-pointer text-black gap-x-2"
        >
          <span class="w-4 h-4">
            {#if option.key === activeTimeGrain}
              <Check size="16px" />
            {/if}
          </span>
          <p class:font-bold={option.key === activeTimeGrain}>{option.main}</p>
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
