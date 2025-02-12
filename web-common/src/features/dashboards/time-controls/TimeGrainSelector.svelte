<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
  import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "../../../runtime-client";
  import { getAllowedTimeGrains } from "@rilldata/web-common/lib/time/grains";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Switch from "@rilldata/web-common/components/button/Switch.svelte";

  export let tdd = false;
  export let activeTimeGrain: V1TimeGrain | undefined;
  export let timeStart: string | undefined;
  export let timeEnd: string | undefined;
  export let minTimeGrain: V1TimeGrain | undefined;
  export let onTimeGrainSelect: (timeGrain: V1TimeGrain) => void;
  export let complete: boolean;
  export let toggleComplete: () => void;

  let open = false;

  $: timeGrainOptions =
    timeStart && timeEnd
      ? getAllowedTimeGrains(new Date(timeStart), new Date(timeEnd))
      : [];
  $: activeTimeGrainLabel =
    activeTimeGrain && TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label;

  $: capitalizedLabel = activeTimeGrainLabel
    ?.split(" ")
    .map((word) => {
      return word.charAt(0).toUpperCase() + word.slice(1);
    })
    .join(" ");

  $: timeGrains = minTimeGrain
    ? timeGrainOptions
        .filter((timeGrain) => !isGrainBigger(minTimeGrain, timeGrain.grain))
        .map((timeGrain) => {
          return {
            main: timeGrain.label,
            key: timeGrain.grain,
          };
        })
    : [];

  import { featureFlags } from "../../feature-flags";

  const { rillTime } = featureFlags;
</script>

{#if activeTimeGrain && timeGrainOptions.length && minTimeGrain}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger asChild let:builder>
      <button
        use:builder.action
        {...builder}
        aria-label="Select a time grain"
        class="flex items-center gap-x-1"
      >
        <div class="items-center flex gap-x-1">
          <span>
            <svelte:element this={tdd ? "b" : "span"}>
              {tdd ? "Time" : "by"}
            </svelte:element>

            <svelte:element this={tdd ? "span" : "b"}>
              {capitalizedLabel}
            </svelte:element>

            {#if complete}
              <i class="ml-0.5">complete</i>
            {/if}
          </span>
          <span class="flex-none transition-transform" class:-rotate-180={open}>
            <CaretDownIcon />
          </span>
        </div>
      </button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="min-w-52" align="start">
      {#each timeGrains as option (option.key)}
        <DropdownMenu.CheckboxItem
          role="menuitem"
          checked={option.key === activeTimeGrain}
          class="text-xs cursor-pointer capitalize"
          on:click={() => onTimeGrainSelect(option.key)}
        >
          {option.main}
        </DropdownMenu.CheckboxItem>
      {/each}

      {#if $rillTime}
        <DropdownMenu.Separator />
        <div class="flex justify-between px-2 py-1">
          <label for="complete" class="select-none cursor-pointer">
            Complete periods
          </label>
          <Switch id="complete" checked={complete} on:click={toggleComplete} />
        </div>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
