<!-- <script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceAddOrganizationMemberUsergroup,
    createAdminServiceAddUsergroupMemberUser,
    createAdminServiceCreateUsergroup,
    createAdminServiceDeleteUsergroup,
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationMemberUsergroups,
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceRemoveOrganizationMemberUsergroup,
    createAdminServiceRemoveUsergroupMemberUser,
    createAdminServiceRenameUsergroup,
    createAdminServiceSetOrganizationMemberUsergroupRole,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    getAdminServiceListUsergroupMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgGroupsTable from "@rilldata/web-admin/features/organizations/users/OrgGroupsTable.svelte";
  import { eventBus } from "@rilldata/events";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import CreateUserGroupDialog from "@rilldata/web-admin/features/organizations/users/CreateUserGroupDialog.svelte";
  import { Search } from "@rilldata/web-common/components/search";

  let userGroupName = "";
  let isCreateUserGroupDialogOpen = false;
  let searchText = "";

  $: organization = $page.params.organization;
  // FIXME: Ideally, we have membersCount in this response
  $: listOrganizationMemberUsergroups =
    createAdminServiceListOrganizationMemberUsergroups(organization);
  $: listOrganizationMemberUsers =
    createAdminServiceListOrganizationMemberUsers(organization);

  const queryClient = useQueryClient();
  const currentUser = createAdminServiceGetCurrentUser();
  const createUserGroup = createAdminServiceCreateUsergroup();
  const renameUserGroup = createAdminServiceRenameUsergroup();
  const deleteUserGroup = createAdminServiceDeleteUsergroup();
  const addUserGroupRole = createAdminServiceAddOrganizationMemberUsergroup();
  const setUserGroupRole =
    createAdminServiceSetOrganizationMemberUsergroupRole();
  const revokeUserGroupRole =
    createAdminServiceRemoveOrganizationMemberUsergroup();
  const removeUserGroupMember = createAdminServiceRemoveUsergroupMemberUser();
  const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();

  $: filteredGroups = $listOrganizationMemberUsergroups.data?.members.filter(
    (group) => group.groupName.toLowerCase().includes(searchText.toLowerCase()),
  );

  async function handleCreate(newName: string) {
    try {
      await $createUserGroup.mutateAsync({
        organization: organization,
        data: {
          name: newName,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      userGroupName = "";
      isCreateUserGroupDialogOpen = false;

      eventBus.emit("notification", { message: "User group created" });
    } catch (error) {
      console.error("Error creating user group", error);
      eventBus.emit("notification", {
        message: "Error creating user group",
        type: "error",
      });
    }
  }

  async function handleRename(groupName: string, newName: string) {
    try {
      await $renameUserGroup.mutateAsync({
        organization: organization,
        usergroup: groupName,
        data: {
          name: newName,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group renamed" });
    } catch (error) {
      console.error("Error renaming user group", error);
      eventBus.emit("notification", {
        message: "Error renaming user group",
        type: "error",
      });
    }
  }

  async function handleDelete(deletedUserGroupName: string) {
    try {
      await $deleteUserGroup.mutateAsync({
        organization: organization,
        usergroup: deletedUserGroupName,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group deleted" });
    } catch (error) {
      console.error("Error deleting user group", error);
      eventBus.emit("notification", {
        message: "Error deleting user group",
        type: "error",
      });
    }
  }

  async function handleAddRole(groupName: string, role: string) {
    try {
      await $addUserGroupRole.mutateAsync({
        organization: organization,
        usergroup: groupName,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group role added" });
    } catch (error) {
      console.error("Error adding role to user group", error);
      eventBus.emit("notification", {
        message: "Error adding role to user group",
        type: "error",
      });
    }
  }

  async function handleSetRole(groupName: string, role: string) {
    try {
      await $setUserGroupRole.mutateAsync({
        organization: organization,
        usergroup: groupName,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group role updated" });
    } catch (error) {
      console.error("Error updating user group role", error);
      eventBus.emit("notification", {
        message: "Error updating user group role",
        type: "error",
      });
    }
  }

  async function handleRevokeRole(groupName: string) {
    try {
      await $revokeUserGroupRole.mutateAsync({
        organization: organization,
        usergroup: groupName,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group role revoked" });
    } catch (error) {
      console.error("Error revoking user group role", error);
      eventBus.emit("notification", {
        message: "Error revoking user group role",
        type: "error",
      });
    }
  }

  async function handleRemoveUser(groupName: string, email: string) {
    try {
      await $removeUserGroupMember.mutateAsync({
        organization: organization,
        usergroup: groupName,
        email: email,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListUsergroupMemberUsersQueryKey(
          organization,
          groupName,
        ),
      );

      eventBus.emit("notification", {
        message: "User removed from user group",
      });
    } catch (error) {
      console.error("Error removing user from user group", error);
      eventBus.emit("notification", {
        message: "Error removing user from user group",
        type: "error",
      });
    }
  }

  async function handleAddUsergroupMemberUser(
    email: string,
    usergroup: string,
  ) {
    try {
      await $addUsergroupMemberUser.mutateAsync({
        organization: organization,
        usergroup: usergroup,
        email: email,
        data: {},
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListUsergroupMemberUsersQueryKey(
          organization,
          usergroup,
        ),
      );

      eventBus.emit("notification", {
        message: "User added to user group",
      });
    } catch (error) {
      console.error("Error adding user to user group", error);
      eventBus.emit("notification", {
        message: "Error adding user to user group",
        type: "error",
      });
    }
  }
</script>

<div class="flex flex-col w-full">
  {#if $listOrganizationMemberUsergroups.isLoading}
    <DelayedSpinner
      isLoading={$listOrganizationMemberUsergroups.isLoading}
      size="1rem"
    />
  {:else if $listOrganizationMemberUsergroups.isError}
    <div class="text-red-500">
      Error loading organization user groups: {$listOrganizationMemberUsergroups.error}
    </div>
  {:else if $listOrganizationMemberUsergroups.isSuccess && $listOrganizationMemberUsers.isSuccess}
    <div class="flex flex-col gap-4">
      <div class="flex flex-row gap-x-4">
        <Search
          placeholder="Search"
          bind:value={searchText}
          large
          autofocus={false}
          showBorderOnFocus={false}
        />
        <Button
          type="primary"
          large
          on:click={() => (isCreateUserGroupDialogOpen = true)}
        >
          <Plus size="16px" />
          <span>Create group</span>
        </Button>
      </div>
      <OrgGroupsTable
        data={filteredGroups}
        currentUserEmail={$currentUser.data?.user.email}
        searchUsersList={$listOrganizationMemberUsers.data?.members ?? []}
        onRename={handleRename}
        onDelete={handleDelete}
        onAddRole={handleAddRole}
        onSetRole={handleSetRole}
        onRevokeRole={handleRevokeRole}
        onRemoveUser={handleRemoveUser}
        onAddUser={handleAddUsergroupMemberUser}
      />
    </div>
  {/if}
</div>

<CreateUserGroupDialog
  bind:open={isCreateUserGroupDialogOpen}
  groupName={userGroupName}
  onCreate={handleCreate}
/> -->
