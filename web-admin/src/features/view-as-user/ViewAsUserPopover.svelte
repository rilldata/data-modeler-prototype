<script lang="ts">
  import {
    DropdownMenuGroup,
    DropdownMenuItem,
  } from "@rilldata/web-common/components/dropdown-menu";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import { matchSorter } from "match-sorter";
  import { createEventDispatcher } from "svelte";
  import {
    createAdminServiceSearchProjectUsers,
    type V1User,
  } from "../../client";
  import { errorStore } from "../../components/errors/error-store";
  import { viewAsUserStore } from "./viewAsUserStore";

  export let organization: string;
  export let project: string;

  // Note: this approach will break down if/when there are more than 1000 users in a project
  $: projectUsers = createAdminServiceSearchProjectUsers(
    organization,
    project,
    { emailQuery: "%", pageSize: 1000, pageToken: undefined },
  );

  const dispatch = createEventDispatcher();

  function handleViewAsUser(user: V1User) {
    viewAsUserStore.set(user);
    errorStore.reset();
    dispatch("select");
  }

  let searchText = "";

  $: clientSideUsers = $projectUsers.data?.users ?? [];

  $: visibleUsers = searchText
    ? matchSorter(clientSideUsers, searchText, { keys: ["email"] })
    : clientSideUsers;
</script>

<div class="px-0.5 pt-1 pb-2 text-[10px] text-gray-500 text-left">
  Test your <a target="_blank" href="https://docs.rilldata.com/manage/security"
    >security policies</a
  > by viewing this project from the perspective of another user.
</div>
<div class="pb-1">
  <Search bind:value={searchText} autofocus={false} />
</div>
{#if visibleUsers.length > 0}
  <DropdownMenuGroup class="overflow-auto pb-1">
    {#each visibleUsers as user}
      <DropdownMenuItem on:click={() => handleViewAsUser(user)}>
        {#if user === $viewAsUserStore}
          <Check size="20px" color="#15141A" />
        {:else}
          <Spacer size="20px" />
        {/if}
        {user.email}
      </DropdownMenuItem>
    {/each}
  </DropdownMenuGroup>
{:else}
  <div class="mt-5 ui-copy-disabled text-center">no results</div>
{/if}
