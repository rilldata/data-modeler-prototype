<script lang="ts">
  import { page } from "$app/stores";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import NoUser from "@rilldata/web-common/components/icons/NoUser.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { initPylonChat } from "@rilldata/web-common/features/help/initPylonChat";
  import {
    createLocalServiceGetCurrentUser,
    createLocalServiceGetMetadata,
  } from "@rilldata/web-common/runtime-client/local-service";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  $: user = createLocalServiceGetCurrentUser({
    query: {
      // refetch in case user does a login/logout from outside of rill developer UI
      refetchOnWindowFocus: true,
    },
  });
  $: metadata = createLocalServiceGetMetadata();

  let loginUrl: string;
  $: if ($metadata.data?.loginUrl) {
    const u = new URL($metadata.data.loginUrl);
    u.searchParams.set(
      "redirect",
      `${window.location.origin}${window.location.pathname}`,
    );
    loginUrl = u.toString();
  }

  let logoutUrl: string;
  $: if ($metadata.data?.loginUrl) {
    const u = new URL($metadata.data.loginUrl + "/logout");
    u.searchParams.set("redirect", $page.url.href);
    logoutUrl = u.toString();
  }

  $: loggedIn = $user.isSuccess && $user.data?.user;

  $: if ($user.data?.user) {
    initPylonChat($user.data.user);
  }
  function handlePylon() {
    window.Pylon("show");
  }

  let photoUrlErrored = false;
</script>

<!-- {#if ($user.isLoading || $metadata.isLoading) && !$user.error && !$metadata.error}
  <div class="flex flex-row items-center h-7 mx-1.5">
    <Spinner size="16px" status={EntityStatus.Running} />
  </div>
{:else if $user.data && $metadata.data} -->
<DropdownMenu.Root>
  <DropdownMenu.Trigger class="flex-none w-7">
    {#if loggedIn && !photoUrlErrored}
      <Avatar
        src={$user.data?.user?.photoUrl}
        alt={$user.data?.user?.displayName || $user.data?.user?.email}
        avatarSize="h-7 w-7"
      />
    {:else}
      <NoUser />
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content>
    <DropdownMenu.Item
      href="https://docs.rilldata.com"
      target="_blank"
      rel="noreferrer noopener"
      class="text-gray-800 font-normal"
    >
      Documentation
    </DropdownMenu.Item>
    <DropdownMenu.Item
      href="https://discord.com/invite/ngVV4KzEGv?utm_source=rill&utm_medium=rill-cloud-avatar-menu"
      target="_blank"
      rel="noreferrer noopener"
      class="text-gray-800 font-normal"
    >
      Join us on Discord
    </DropdownMenu.Item>
    {#if loggedIn}
      <DropdownMenu.Item on:click={handlePylon}>
        Contact Rill support
      </DropdownMenu.Item>
      <DropdownMenu.Item
        href={logoutUrl}
        rel="external"
        class="text-gray-800 font-normal"
      >
        Logout
      </DropdownMenu.Item>
    {:else}
      <DropdownMenu.Item
        href={loginUrl}
        rel="external"
        class="text-gray-800 font-normal"
      >
        Log in / Sign up
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
<!-- {/if} -->
