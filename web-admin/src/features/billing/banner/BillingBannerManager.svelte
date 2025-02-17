<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import type { V1OrganizationPermissions } from "@rilldata/web-admin/client";
  import BillingBannerManagerForAdmins from "@rilldata/web-admin/features/billing/banner/BillingBannerManagerForAdmins.svelte";
  import BillingBannerManagerForViewers from "@rilldata/web-admin/features/billing/banner/BillingBannerManagerForViewers.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let organization: string;
  export let organizationPermissions: V1OrganizationPermissions;

  onNavigate(({ from, to }) => {
    const changedOrganization =
      !from || !to || from.params.organization !== to.params.organization;
    if (changedOrganization) {
      eventBus.emit("banner", null);
    }
  });
</script>

{#if organizationPermissions.manageOrg}
  <BillingBannerManagerForAdmins {organization} />
{:else if organizationPermissions.readOrg}
  <BillingBannerManagerForViewers {organization} />
{/if}
