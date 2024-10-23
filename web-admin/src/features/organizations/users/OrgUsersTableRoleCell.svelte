<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    createAdminServiceSetOrganizationMemberUserRole,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import { page } from "$app/stores";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/events";

  export let email: string;
  export let role: string;
  export let isCurrentUser: boolean;

  let isDropdownOpen = false;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const setOrganizationMemberUserRole =
    createAdminServiceSetOrganizationMemberUserRole();

  async function handleSetRole(role: string) {
    try {
      await $setOrganizationMemberUserRole.mutateAsync({
        organization: organization,
        email: email,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationInvitesQueryKey(organization),
      );

      eventBus.emit("notification", {
        message: "User role updated",
      });
    } catch (error) {
      console.error("Error updating user role", error);
      eventBus.emit("notification", {
        message: "Error updating user role",
        type: "error",
      });
    }
  }
</script>

{#if !isCurrentUser}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      {role ? `Org ${role}` : "-"}
      {#if isDropdownOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={role === "admin"}
        on:click={() => {
          handleSetRole("admin");
        }}
      >
        <span>Admin</span>
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.CheckboxItem
        class="font-normal flex items-center"
        checked={role === "viewer"}
        on:click={() => {
          handleSetRole("viewer");
        }}
      >
        <span>Viewer</span>
      </DropdownMenu.CheckboxItem>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <div class="w-18 flex flex-row gap-1 items-center rounded-sm px-2 py-1">
    <span>Org {role}</span>
  </div>
{/if}
