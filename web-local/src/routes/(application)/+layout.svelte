<script lang="ts">
  import { page } from "$app/stores";
  import BannerCenter from "@rilldata/web-common/components/banner/BannerCenter.svelte";
  import RepresentingUserBanner from "@rilldata/web-common/features/authentication/RepresentingUserBanner.svelte";
  import AddDataModal from "@rilldata/web-common/features/sources/modal/AddDataModal.svelte";
  import FileDrop from "@rilldata/web-common/features/sources/modal/FileDrop.svelte";
  import SourceImportedModal from "@rilldata/web-common/features/sources/modal/SourceImportedModal.svelte";
  import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store";
  import ApplicationHeader from "@rilldata/web-common/layout/ApplicationHeader.svelte";

  let showDropOverlay = false;

  $: ({ route } = $page);
  $: mode = route.id?.includes("(preview)") ? "Preview" : "Developer";

  function isEventWithFiles(event: DragEvent) {
    let types = event?.dataTransfer?.types;
    return types && types.indexOf("Files") != -1;
  }
</script>

<BannerCenter />
<RepresentingUserBanner />
<ApplicationHeader {mode} />

<main
  role="application"
  class="index-body relative size-full flex flex-col overflow-hidden"
  on:drag|preventDefault|stopPropagation
  on:drop|preventDefault|stopPropagation
  on:dragenter|preventDefault|stopPropagation
  on:dragleave|preventDefault|stopPropagation
  on:dragover|preventDefault|stopPropagation={(e) => {
    if (isEventWithFiles(e)) showDropOverlay = true;
  }}
>
  <slot />
</main>

{#if showDropOverlay}
  <FileDrop bind:showDropOverlay />
{/if}

<AddDataModal />
<SourceImportedModal sourcePath={$sourceImportedPath} />
