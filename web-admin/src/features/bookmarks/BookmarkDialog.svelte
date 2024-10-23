<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateBookmark,
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import BaseBookmarkForm from "@rilldata/web-admin/features/bookmarks/BaseBookmarkForm.svelte";
  import type { BookmarkFormValues } from "@rilldata/web-admin/features/bookmarks/form-utils";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useExploreStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { eventBus } from "@rilldata/events";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";

  export let metricsViewName: string;
  export let exploreName: string;
  export let bookmark: BookmarkEntry | null = null;
  export let onClose = () => {};

  const StateManagers = getStateManagers();

  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();
  const timeControlsStore = useTimeControlStore(StateManagers);

  const formState = createForm<BookmarkFormValues>({
    initialValues: {
      displayName: bookmark?.resource.displayName ?? "Default Label",
      description: bookmark?.resource.description ?? "",
      shared: bookmark?.resource.shared ? "true" : "false",
      filtersOnly: bookmark?.filtersOnly ?? false,
      absoluteTimeRange: bookmark?.absoluteTimeRange ?? false,
    },
    validationSchema: yup.object({
      displayName: yup.string().required("Required"),
      description: yup.string(),
    }),
    onSubmit: async (values) => {
      if (bookmark) {
        await $bookmarkUpdater.mutateAsync({
          data: {
            bookmarkId: bookmark.resource.id,
            displayName: values.displayName,
            description: values.description,
            shared: values.shared === "true",
            data: getBookmarkDataForDashboard(
              $exploreStore,
              values.filtersOnly,
              values.absoluteTimeRange,
              $timeControlsStore,
            ),
          },
        });
      } else {
        await $bookmarkCreator.mutateAsync({
          data: {
            displayName: values.displayName,
            description: values.description,
            projectId: $projectId.data ?? "",
            resourceKind: ResourceKind.Explore,
            resourceName: exploreName,
            shared: values.shared === "true",
            data: getBookmarkDataForDashboard(
              $exploreStore,
              values.filtersOnly,
              values.absoluteTimeRange,
              $timeControlsStore,
            ),
          },
        });
        handleReset();
      }
      onClose();

      await queryClient.refetchQueries(
        getAdminServiceListBookmarksQueryKey({
          projectId: $projectId.data ?? "",
          resourceKind: ResourceKind.Explore,
          resourceName: exploreName,
        }),
      );
      eventBus.emit("notification", {
        message: bookmark ? "Bookmark updated" : "Bookmark created",
      });
    },
  });

  const { handleSubmit, handleReset } = formState;

  $: ({ params } = $page);
  $: exploreStore = useExploreStore(exploreName);
  $: projectId = useProjectId(params.organization, params.project);
</script>

<Dialog.Root
  open
  onOpenChange={(o) => {
    if (!o) onClose();
  }}
>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>
        {bookmark ? "Edit bookmark" : "Bookmark current view"}
      </Dialog.Title>
    </Dialog.Header>

    <BaseBookmarkForm {formState} {metricsViewName} {exploreName} />

    <div class="flex flex-row mt-4 gap-2">
      <div class="grow" />
      <Button on:click={onClose} type="secondary">Cancel</Button>
      <Button on:click={handleSubmit} type="primary">Save</Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
