<script lang="ts">
  import { page } from "$app/stores";
  import BillingBannerManager from "@rilldata/web-admin/features/billing/banner/BillingBannerManager.svelte";
  import {
    isBillingUpgradePage,
    isProjectInvitePage,
    isProjectsBreakdownPage,
    isPublicReportPage,
    withinOrganization,
    withinProject,
  } from "@rilldata/web-admin/features/navigation/nav-utils";
  import OrganizationTabs from "@rilldata/web-admin/features/organizations/OrganizationTabs.svelte";
  import { initCloudMetrics } from "@rilldata/web-admin/features/telemetry/initCloudMetrics";
  import BannerCenter from "@rilldata/web-common/components/banner/BannerCenter.svelte";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { initPylonWidget } from "@rilldata/web-common/features/help/initPylonWidget";
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import ErrorBoundary from "../features/errors/ErrorBoundary.svelte";
  import { createGlobalErrorCallback } from "../features/errors/error-utils";
  import TopNavigationBar from "../features/navigation/TopNavigationBar.svelte";

  export let data;

  $: ({ projectPermissions, organizationPermissions } = data);
  $: organization = $page.params.organization;

  // Motivation:
  // - https://tkdodo.eu/blog/breaking-react-querys-api-on-purpose#a-bad-api
  // - https://tkdodo.eu/blog/react-query-error-handling#the-global-callbacks
  queryClient.getQueryCache().config.onError =
    createGlobalErrorCallback(queryClient);

  // The admin server enables some dashboard features like scheduled reports and alerts
  // Set read-only mode so that the user can't edit the dashboard
  featureFlags.set(true, "adminServer", "readOnly");

  let removeJavascriptListeners: () => void;

  initCloudMetrics()
    .then(() => {
      removeJavascriptListeners =
        errorEventHandler?.addJavascriptErrorListeners();
    })
    .catch(console.error);
  initPylonWidget();

  onMount(() => {
    return () => removeJavascriptListeners?.();
  });

  $: isEmbed = $page.url.pathname === "/-/embed";

  $: hideTopBar =
    // invite page shouldn't show the top bar because it is considered an onboard step
    isProjectInvitePage($page) ||
    // upgrade callback landing page shouldn't show any rill identifications
    isBillingUpgradePage($page) ||
    // public reports are shared to external users who shouldn't be shown any rill related stuff
    isPublicReportPage($page);
  $: hideBillingManager =
    // billing manager needs organization
    !organization ||
    // invite page shouldn't show the banner since the illusion is that the user is not on cloud yet.
    isProjectInvitePage($page);

  $: showOrgTabs =
    withinOrganization($page) &&
    !withinProject($page) &&
    !isProjectsBreakdownPage($page);
</script>

<svelte:head>
  <meta content="Rill Cloud" name="description" />
</svelte:head>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <main class="flex flex-col min-h-screen h-screen">
      <BannerCenter />
      {#if !hideBillingManager}
        <BillingBannerManager {organization} {organizationPermissions} />
      {/if}
      {#if !isEmbed && !hideTopBar}
        <TopNavigationBar
          manageOrganization={organizationPermissions?.manageOrg}
          createMagicAuthTokens={projectPermissions?.createMagicAuthTokens}
          manageProjectMembers={projectPermissions?.manageProjectMembers}
        />

        {#if showOrgTabs}
          <OrganizationTabs />
        {/if}
      {/if}
      <ErrorBoundary>
        <slot />
      </ErrorBoundary>
    </main>
  </QueryClientProvider>

  <NotificationCenter />
</RillTheme>
