<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { Trash2Icon } from "lucide-svelte";
  import RemoveUserFromOrgConfirmDialog from "./RemoveUserFromOrgConfirmDialog.svelte";
  import {
    createAdminServiceRemoveOrganizationMemberUser,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/events";
  import { page } from "$app/stores";

  export let email: string;
  export let isCurrentUser: boolean;

  let isDropdownOpen = false;
  let isRemoveConfirmOpen = false;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const removeOrganizationMemberUser =
    createAdminServiceRemoveOrganizationMemberUser();

  async function handleRemove(email: string) {
    try {
      await $removeOrganizationMemberUser.mutateAsync({
        organization: organization,
        email: email,
        // Uncomment if `keepProjectRoles` is needed
        // See: https://github.com/rilldata/rill/pull/2231
        // params: {
        //   keepProjectRoles: false,
        // },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationInvitesQueryKey(organization),
      );

      eventBus.emit("notification", {
        message: "User removed from organization",
      });
    } catch (error) {
      console.error("Error removing user from organization", error);
      eventBus.emit("notification", {
        message: "Error removing user from organization",
        type: "error",
      });
    }
  }
</script>

{#if !isCurrentUser}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        type="destructive"
        on:click={() => {
          isRemoveConfirmOpen = true;
        }}
      >
        <Trash2Icon size="12px" />
        <span class="ml-2">Remove</span>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<RemoveUserFromOrgConfirmDialog
  bind:open={isRemoveConfirmOpen}
  {email}
  onRemove={handleRemove}
/>
