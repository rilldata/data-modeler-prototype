<script lang="ts">
  import { type V1OrganizationPermissions } from "@rilldata/web-admin/client";
  import Tab from "@rilldata/web-admin/components/nav/Tab.svelte";
  import {
    width,
    position,
  } from "@rilldata/web-admin//components/nav/Tab.svelte";

  export let organization: string;
  export let organizationPermissions: V1OrganizationPermissions;
  export let pathname: string;

  $: tabs = [
    {
      route: `/${organization}`,
      label: "Projects",
      hasPermission: true,
    },
    {
      route: `/${organization}/-/users`,
      label: "Users",
      hasPermission: organizationPermissions.manageOrgMembers,
    },
    {
      route: `/${organization}/-/settings`,
      label: "Settings",
      hasPermission: organizationPermissions.manageOrg,
    },
  ];

  $: showTabs =
    organizationPermissions.manageOrg ||
    organizationPermissions.manageOrgMembers;

  $: selectedIndex = tabs?.findLastIndex((t) => pathname.startsWith(t.route));
</script>

<div>
  {#if showTabs}
    <nav>
      {#each tabs as tab, i (tab.route)}
        {#if tab.hasPermission}
          <Tab
            route={tab.route}
            label={tab.label}
            selected={selectedIndex === i}
            {organization}
          />
        {/if}
      {/each}
    </nav>
  {/if}
  {#if $width && $position}
    <span
      style:width="{$width}px"
      style:transform="translateX({$position}px) "
    />
  {/if}
</div>

<style lang="postcss">
  div {
    @apply border-b pt-1;
    @apply gap-y-[3px] flex flex-col;
  }

  nav {
    @apply flex w-fit;
    @apply gap-x-3 px-[17px];
  }

  span {
    @apply h-[3px] bg-primary-500 rounded transition-all;
  }
</style>
