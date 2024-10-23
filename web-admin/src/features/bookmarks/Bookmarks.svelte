<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceRemoveBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import BookmarkDialog from "@rilldata/web-admin/features/bookmarks/BookmarkDialog.svelte";
  import BookmarksContent from "@rilldata/web-admin/features/bookmarks/BookmarksDropdownMenuContent.svelte";
  import { createBookmarkApplier } from "@rilldata/web-admin/features/bookmarks/applyBookmark";
  import { createHomeBookmarkModifier } from "@rilldata/web-admin/features/bookmarks/createOrUpdateHomeBookmark";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { useExploreStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { eventBus } from "@rilldata/events";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkIcon } from "lucide-svelte";

  export let metricsViewName: string;
  export let exploreName: string;

  let showDialog = false;
  let bookmark: BookmarkEntry | null = null;

  $: bookmarkApplier = createBookmarkApplier(
    $runtime?.instanceId,
    metricsViewName,
    exploreName,
  );

  $: exploreStore = useExploreStore(exploreName);
  $: projectId = useProjectId($page.params.organization, $page.params.project);

  const queryClient = useQueryClient();
  $: homeBookmarkModifier = createHomeBookmarkModifier(
    $runtime?.instanceId,
    metricsViewName,
    exploreName,
  );
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  function selectBookmark(bookmark: BookmarkEntry) {
    bookmarkApplier(bookmark.resource);
  }

  async function createHomeBookmark() {
    await homeBookmarkModifier(getBookmarkDataForDashboard($exploreStore));
    eventBus.emit("notification", {
      message: "Home bookmark created",
    });
    return queryClient.refetchQueries(
      getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.Explore,
        resourceName: exploreName,
      }),
    );
  }

  async function deleteBookmark(bookmark: BookmarkEntry) {
    // TODO: add confirmation?
    await $bookmarkDeleter.mutateAsync({
      bookmarkId: bookmark.resource.id,
    });
    eventBus.emit("notification", {
      message: `Bookmark ${bookmark.resource.displayName} deleted`,
    });
    return queryClient.refetchQueries(
      getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.Explore,
        resourceName: exploreName,
      }),
    );
  }

  let open = false;
</script>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger asChild let:builder>
    <Button builders={[builder]} compact type="secondary">
      <BookmarkIcon
        class="inline-flex"
        fill={open ? "black" : "none"}
        size="16px"
      />
    </Button>
  </DropdownMenuTrigger>
  <BookmarksContent
    on:create={() => (showDialog = true)}
    on:create-home={() => createHomeBookmark()}
    on:delete={({ detail }) => deleteBookmark(detail)}
    on:edit={({ detail }) => {
      showDialog = true;
      bookmark = detail;
    }}
    on:select={({ detail }) => selectBookmark(detail)}
    {metricsViewName}
    {exploreName}
  />
</DropdownMenu>

{#if showDialog}
  <BookmarkDialog
    {bookmark}
    {metricsViewName}
    {exploreName}
    onClose={() => {
      showDialog = false;
      bookmark = null;
    }}
  />
{/if}
