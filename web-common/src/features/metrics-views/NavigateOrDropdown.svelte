<script lang="ts">
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { removeLeadingSlash } from "../entity-management/entity-mappers";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import type { Builder } from "bits-ui";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import {
    displayResourceKind,
    ResourceKind,
  } from "../entity-management/resource-selectors";

  export let resources: V1Resource[];
  export let builder: Builder;

  $: firstResource = resources?.[0];
  $: firstResourceType = displayResourceKind(
    firstResource?.meta?.name?.kind as ResourceKind | undefined,
  );
</script>

{#if resources?.length === 1 && firstResource.meta?.filePaths?.[0]}
  <div
    class="border-primary-300 flex items-center border h-7 rounded-[2px] bg-transparent text-primary-600"
  >
    <a
      href={`/files/${removeLeadingSlash(firstResource.meta?.filePaths?.[0])}`}
      class="text-inherit font-medium flex items-center border-r px-3 size-full hover:bg-primary-50 border-primary-300"
    >
      Go to {firstResourceType}
    </a>
    <button
      aria-label="Create resource menu"
      use:builder.action
      {...builder}
      class="text-inherit h-full aspect-square grid place-content-center hover:bg-primary-50"
    >
      <CaretDownIcon />
    </button>
  </div>
{:else}
  <Button type="secondary" builders={[builder]}>
    Go to {firstResourceType}
    <CaretDownIcon />
  </Button>
{/if}
