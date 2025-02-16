<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListBookmarks,
  } from "@rilldata/web-admin/client";
  import BookmarkItem from "@rilldata/web-admin/features/bookmarks/BookmarksDropdownMenuItem.svelte";
  import {
    type BookmarkEntry,
    categorizeBookmarks,
    searchBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/selectors";
  import {
    getProjectPermissions,
    useProjectId,
  } from "@rilldata/web-admin/features/projects/selectors";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import HomeBookmarkPlus from "@rilldata/web-common/components/icons/HomeBookmarkPlus.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { BookmarkPlusIcon } from "lucide-svelte";

  export let metricsViewName: string;
  export let exploreName: string;
  export let onCreate: (isHome: boolean) => void;
  export let onEdit: (bookmark: BookmarkEntry) => void;
  export let onDelete: (bookmark: BookmarkEntry) => Promise<void>;

  $: ({ instanceId } = $runtime);

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: exploreState = useExploreState(exploreName);
  $: validExploreSpec = useExploreValidSpec(instanceId, exploreName);
  $: metricsViewSpec = $validExploreSpec.data?.metricsView ?? {};
  $: exploreSpec = $validExploreSpec.data?.explore ?? {};
  $: metricsViewTimeRange = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
  );
  $: defaultExplorePreset = getDefaultExplorePreset(
    exploreSpec,
    $metricsViewTimeRange.data,
  );
  $: schemaResp = createQueryServiceMetricsViewSchema(
    instanceId,
    metricsViewName,
  );

  $: projectIdResp = useProjectId(organization, project);
  const userResp = createAdminServiceGetCurrentUser();
  $: bookamrksResp = createAdminServiceListBookmarks(
    {
      projectId: $projectIdResp.data,
      resourceKind: ResourceKind.Explore,
      resourceName: exploreName,
    },
    {
      query: {
        enabled: !!$projectIdResp.data && !!$userResp.data.user,
      },
    },
  );

  let searchText: string;
  $: categorizedBookmarks = categorizeBookmarks(
    $bookamrksResp.data?.bookmarks ?? [],
    metricsViewSpec,
    exploreSpec,
    $schemaResp.data?.schema,
    $exploreState,
    defaultExplorePreset,
    $metricsViewTimeRange.data?.timeRangeSummary,
  );
  $: filteredBookmarks = searchBookmarks(categorizedBookmarks, searchText);

  $: projectPermissions = getProjectPermissions(organization, project);
  $: manageProject = $projectPermissions.data?.manageProject;
</script>

<DropdownMenu.Content class="w-[450px]">
  <DropdownMenu.Item on:click={() => onCreate(false)}>
    <div class="flex flex-row gap-x-2.5 items-center h-7">
      <BookmarkPlusIcon size="18px" strokeWidth={1.5} />
      <div>Bookmark current view</div>
    </div>
  </DropdownMenu.Item>

  {#if manageProject}
    <DropdownMenu.Item on:click={() => onCreate(true)} slot="manage-project">
      <div class="flex flex-row gap-x-2.5 items-center h-7">
        <HomeBookmarkPlus size="18px" />
        <div>
          <div class="text-xs font-medium text-gray-700 h-4">
            Bookmark current view as Home.
          </div>
          <div class="text-[11px] font-normal text-gray-500 h-4">
            This will be everyone’s main view for this dashboard.
          </div>
        </div>
      </div>
    </DropdownMenu.Item>
  {/if}

  <DropdownMenu.Separator />

  <div class="p-2">
    <Search
      autofocus={false}
      bind:value={searchText}
      showBorderOnFocus={false}
    />
  </div>

  {#if filteredBookmarks}
    <DropdownMenu.Separator />
    <DropdownMenu.Group class="gap-y-1 flex flex-col">
      <DropdownMenu.Label class="text-gray-500 text-[10px] h-6 uppercase">
        Your bookmarks
      </DropdownMenu.Label>

      {#each filteredBookmarks.personal as bookmark (bookmark.resource.id)}
        <BookmarkItem
          {bookmark}
          {onEdit}
          {onDelete}
          selected={$page.url.toString() === bookmark.url}
        />
      {:else}
        <div class="my-2 ui-copy-disabled text-center">
          You have no bookmarks for this dashboard.
        </div>
      {/each}
    </DropdownMenu.Group>
    <DropdownMenu.Separator />

    <DropdownMenu.Group class="gap-y-1">
      <DropdownMenu.Label class="text-gray-500">
        <div class="text-[10px] h-4 uppercase">Managed bookmarks</div>
        <div class="text-[11px] font-normal">Created by project admin</div>
      </DropdownMenu.Label>

      {#if filteredBookmarks.shared?.length || filteredBookmarks.home}
        {#if filteredBookmarks.home}
          <BookmarkItem
            bookmark={filteredBookmarks.home}
            {onEdit}
            {onDelete}
            readOnly={!manageProject}
            selected={$page.url.toString() === filteredBookmarks.home.url}
          />
        {/if}

        {#each filteredBookmarks.shared as bookmark (bookmark.resource.id)}
          <BookmarkItem
            {bookmark}
            readOnly={!manageProject}
            selected={$page.url.toString() === bookmark.url}
            {onEdit}
            {onDelete}
          />
          {bookmark.url}
        {/each}
      {:else}
        <div class="my-2 ui-copy-disabled text-center">
          There are no shared bookmarks for this dashboard.
        </div>
      {/if}
    </DropdownMenu.Group>
  {/if}
</DropdownMenu.Content>
