<script lang="ts">
  import { invalidate } from "$app/navigation";
  import {
    createAdminServiceUpdateOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import UploadImagePopover from "@rilldata/web-admin/features/organizations/settings/UploadImagePopover.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let organization: string;
  export let organizationLogoUrl: string | undefined;

  const orgUpdater = createAdminServiceUpdateOrganization();
  $: ({ error, isLoading, mutateAsync } = $orgUpdater);

  async function onSave(assetId: string) {
    await mutateAsync({
      name: organization,
      data: {
        logoAssetId: assetId,
      },
    });
    void queryClient.invalidateQueries(
      getAdminServiceGetOrganizationQueryKey(organization),
    );
    void invalidate("root");
  }

  async function onRemove() {
    await mutateAsync({
      name: organization,
      data: {
        logoAssetId: "",
      },
    });
    void queryClient.invalidateQueries(
      getAdminServiceGetOrganizationQueryKey(organization),
    );
    void invalidate("root");
  }
</script>

<SettingsContainer title="Logo" suppressFooter={!organizationLogoUrl}>
  <div slot="body" class="flex flex-col gap-y-2">
    <div>
      Click to upload your logo and customize Rill for your organization.
    </div>
    <UploadImagePopover
      imageUrl={organizationLogoUrl}
      accept="image/png, image/ico, image/x-ico, image/icon, image/x-icon"
      label="favicon"
      {organization}
      loading={isLoading}
      error={getRpcErrorMessage(error)}
      {onSave}
      {onRemove}
    >
      <Rill width="64" height="40" />
    </UploadImagePopover>
  </div>
  <svelte:fragment slot="action">
    {#if organizationLogoUrl}
      <Button
        type="secondary"
        on:click={onRemove}
        loading={isLoading}
        disabled={isLoading}
      >
        Remove
      </Button>
    {/if}
  </svelte:fragment>
</SettingsContainer>
