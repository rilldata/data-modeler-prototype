<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceRemoveBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { createBookmarkApplier } from "@rilldata/web-admin/features/bookmarks/applyBookmark";
  import BookmarksContent from "@rilldata/web-admin/features/bookmarks/BookmarksDropdownMenuContent.svelte";
  import BookmarkDialog from "@rilldata/web-admin/features/bookmarks/BookmarkDialog.svelte";
  import { createHomeBookmarkModifier } from "@rilldata/web-admin/features/bookmarks/createOrUpdateHomeBookmark";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkIcon } from "lucide-svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { suppressFilterMenus } from "@rilldata/web-common/features/dashboards/filters/Filters.svelte";

  export let metricsViewName: string;

  let bookmark: BookmarkEntry | null = null;

  $: bookmarkApplier = createBookmarkApplier(
    $runtime?.instanceId,
    metricsViewName,
  );

  $: dashboardStore = useDashboardStore(metricsViewName);
  $: projectId = useProjectId($page.params.organization, $page.params.project);

  const queryClient = useQueryClient();
  $: homeBookmarkModifier = createHomeBookmarkModifier($runtime?.instanceId);
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  function selectBookmark(bookmark: BookmarkEntry) {
    bookmarkApplier(bookmark.resource);
  }

  async function createHomeBookmark() {
    await homeBookmarkModifier(getBookmarkDataForDashboard($dashboardStore));
    eventBus.emit("notification", {
      message: "Home bookmark created",
    });
    return queryClient.refetchQueries(
      getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.MetricsView,
        resourceName: metricsViewName,
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
        resourceKind: ResourceKind.MetricsView,
        resourceName: metricsViewName,
      }),
    );
  }

  let open = false;
</script>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger asChild let:builder>
    <Button builders={[builder]} compact type="secondary">
      <BookmarkIcon
        class="inline-flex {open && 'fill-primary-600'}"
        size="16px"
      />
    </Button>
  </DropdownMenuTrigger>
  <BookmarksContent
    on:create={() => suppressFilterMenus.set(true)}
    on:create-home={() => createHomeBookmark()}
    on:delete={({ detail }) => deleteBookmark(detail)}
    on:edit={({ detail }) => {
      suppressFilterMenus.set(true);
      bookmark = detail;
      selectBookmark(detail);
    }}
    on:select={({ detail }) => selectBookmark(detail)}
    {metricsViewName}
  />
</DropdownMenu>

{#if $suppressFilterMenus}
  <BookmarkDialog
    {bookmark}
    {metricsViewName}
    onClose={() => {
      suppressFilterMenus.set(false);
      bookmark = null;
    }}
  />
{/if}
