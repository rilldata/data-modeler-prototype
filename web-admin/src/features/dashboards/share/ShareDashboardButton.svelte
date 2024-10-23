<script lang="ts">
  import CreatePublicURLForm from "@rilldata/web-admin/features/public-urls/CreatePublicURLForm.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Link from "@rilldata/web-common/components/icons/Link.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import { copyToClipboard } from "@rilldata/actions";

  export let createMagicAuthTokens: boolean;

  let isOpen = false;
</script>

<Popover bind:open={isOpen}>
  <PopoverTrigger asChild let:builder>
    <Button type="secondary" builders={[builder]} selected={isOpen}
      >Share</Button
    >
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[402px] p-0">
    <Tabs>
      <TabsList>
        <TabsTrigger value="tab1">Copy URL</TabsTrigger>
        {#if createMagicAuthTokens}
          <TabsTrigger value="tab2">Create public URL</TabsTrigger>
        {/if}
      </TabsList>
      <TabsContent value="tab1" class="mt-0 p-4">
        <div class="flex flex-col gap-y-4">
          <h3 class="text-xs text-gray-800 font-normal">
            Share your current view with another project member.
          </h3>
          <Button
            type="secondary"
            on:click={() => {
              copyToClipboard(window.location.href, "Link copied to clipboard");
            }}
          >
            <Link size="16px" className="text-primary-500" />
            Copy URL
          </Button>
        </div>
      </TabsContent>
      <TabsContent value="tab2" class="mt-0 p-4">
        <CreatePublicURLForm />
      </TabsContent>
    </Tabs>
  </PopoverContent>
</Popover>

<style lang="postcss">
  h3 {
    @apply font-semibold;
  }
</style>
