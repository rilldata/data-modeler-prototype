<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateProjectWhitelistedDomain,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-invite/CopyInviteLinkButton.svelte";
  import {
    getUserDomain,
    userDomainIsPublic,
  } from "@rilldata/web-admin/features/projects/user-invite/selectors";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-invite/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { eventBus } from "@rilldata/events";
  import type { AxiosError } from "axios";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  let allowDomain = false;
  $: userDomain = getUserDomain();
  $: isPublicDomain = userDomainIsPublic();
  const addToAllowlist = createAdminServiceCreateProjectWhitelistedDomain();
  async function onContinue() {
    if (allowDomain) {
      try {
        await $addToAllowlist.mutateAsync({
          organization,
          project,
          data: {
            domain: $userDomain.data,
            role: "viewer",
          },
        });
      } catch (e) {
        eventBus.emit("notification", {
          type: "error",
          message:
            (e as AxiosError<RpcStatus>).response.data?.message ?? e.message,
          options: {
            persisted: true,
          },
        });
      }
    }
    return goto(`/${organization}/${project}/-/status`);
  }

  $: copyLink = `${$page.url.protocol}//${$page.url.host}/${organization}/${project}`;

  let invited = false;
</script>

<div class="flex flex-col gap-5 w-[600px] my-16 sm:my-32 md:my-64 mx-auto">
  <div class="text-xl text-center w-full">Invite teammates to your project</div>
  <div class="flex flex-col gap-y-1">
    <div class="flex flex-row items-center">
      <div class="text-sm font-medium">Invite by email</div>
      <div class="grow"></div>
      <CopyInviteLinkButton {copyLink} />
    </div>
    <UserInviteForm
      {organization}
      {project}
      onInvite={() => (invited = true)}
    />
  </div>
  {#if $userDomain.data && !$isPublicDomain.data}
    <div class="flex flex-col gap-y-1">
      <div class="text-sm font-medium">Allow domain access</div>
      <div class="flex flex-row gap-x-2">
        <Switch
          small
          bind:checked={allowDomain}
          id="allow-domain"
          class="mt-1"
        />
        <Label for="allow-domain" class="font-normal text-gray-700 text-sm">
          Allow any user with a <b>@{$userDomain.data}</b> email address to join
          this project as a <b>Viewer</b>.
          <a
            target="_blank"
            href="https://docs.rilldata.com/reference/cli/user/whitelist"
          >
            Learn more
          </a>
        </Label>
      </div>
    </div>
  {/if}
  <Button
    type="primary"
    on:click={onContinue}
    loading={$addToAllowlist.isLoading}
    wide
    class="mx-auto"
  >
    {invited ? "Continue" : "Skip"}
  </Button>
</div>
