<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/events";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";

  export let organization: string;

  const deleteOrgMutation = createAdminServiceDeleteOrganization();

  async function deleteOrg() {
    await $deleteOrgMutation.mutateAsync({
      name: organization,
    });

    void goto(`/`);
    queryClient.removeQueries(
      getAdminServiceGetOrganizationQueryKey(organization),
    );
    eventBus.emit("notification", {
      message: "Deleted organization",
    });
  }
</script>

<SettingsContainer title="Delete org">
  <svelte:fragment slot="body">
    Permanently remove all contents of this org.
  </svelte:fragment>

  <AlertDialogGuardedConfirmation
    slot="action"
    title="Delete this org?"
    description={`The org ${organization} will be deleted permanently. This action cannot be undone.`}
    confirmText={`delete ${organization}`}
    loading={$deleteOrgMutation.isLoading}
    error={$deleteOrgMutation.error?.message}
    onConfirm={deleteOrg}
  >
    <svelte:fragment let:builder>
      <Button builders={[builder]} type="primary">Delete org</Button>
    </svelte:fragment>
  </AlertDialogGuardedConfirmation>
</SettingsContainer>
