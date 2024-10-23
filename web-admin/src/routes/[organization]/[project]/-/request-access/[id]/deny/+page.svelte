<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceDenyProjectAccess,
    createAdminServiceGetProjectAccessRequest,
  } from "@rilldata/web-admin/client";
  import { parseAccessRequestError } from "@rilldata/web-admin/features/access-request/utils";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import AccessRequestContainer from "@rilldata/web-admin/features/access-request/AccessRequestContainer.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { eventBus } from "@rilldata/events";
  import type { AxiosError } from "axios";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: id = $page.params.id;

  let requested = false;
  $: denyAccess = createAdminServiceDenyProjectAccess();
  $: requestAccess = createAdminServiceGetProjectAccessRequest(id);

  async function onDeny() {
    if ($requestAccess.error) {
      eventBus.emit("notification", {
        type: "error",
        message: parseAccessRequestError(
          project,
          $requestAccess.error as unknown as AxiosError,
        ),
        options: {
          persisted: true,
        },
      });
      return goto(`/${organization}/${project}`);
    }

    requested = true;
    try {
      await $denyAccess.mutateAsync({
        id,
        data: {},
      });
      eventBus.emit("notification", {
        type: "success",
        message: `${$requestAccess.data.email} has been denied access to ${project}`,
      });
    } catch {
      eventBus.emit("notification", {
        type: "error",
        message: parseAccessRequestError(
          project,
          $requestAccess.error as unknown as AxiosError,
        ),
        options: {
          persisted: true,
        },
      });
    }
    return goto(`/${organization}/${project}`);
  }
  $: if (
    organization &&
    project &&
    id &&
    !$requestAccess.isLoading &&
    !requested
  ) {
    onDeny();
  }
</script>

<AccessRequestContainer>
  {#if $denyAccess.isLoading && $requestAccess.data}
    <Spinner status={EntityStatus.Running} size="2rem" duration={725} />
    <div>
      Denying <b>{$requestAccess.data.email}</b> access to
      <b>{project}</b>
    </div>
  {/if}
</AccessRequestContainer>
