<script lang="ts">
  import { page } from "$app/stores";
  import BaseBookmarkForm from "@rilldata/web-admin/features/bookmarks/BaseBookmarkForm.svelte";
  import type { BookmarkFormValues } from "@rilldata/web-admin/features/bookmarks/form-utils";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import {
    createAdminServiceCreateBookmark,
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  export let metricsViewName: string;
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
              $dashboardStore,
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
            resourceKind: ResourceKind.MetricsView,
            resourceName: metricsViewName,
            shared: values.shared === "true",
            data: getBookmarkDataForDashboard(
              $dashboardStore,
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
          resourceKind: ResourceKind.MetricsView,
          resourceName: metricsViewName,
        }),
      );
      eventBus.emit("notification", {
        message: bookmark ? "Bookmark updated" : "Bookmark created",
      });
    },
  });

  const { handleSubmit, handleReset } = formState;

  $: ({ params } = $page);
  $: dashboardStore = useDashboardStore(metricsViewName);
  $: projectId = useProjectId(params.organization, params.project);
</script>

<Dialog.Root
  open
  onOpenChange={(o) => {
    if (!o) onClose();
  }}
  closeOnOutsideClick={false}
>
  <Dialog.Content class="w-full max-w-[620px]">
    <Dialog.Header>
      <Dialog.Title>
        {bookmark ? "Edit bookmark" : "Bookmark current view"}
      </Dialog.Title>
    </Dialog.Header>

    <BaseBookmarkForm {formState} />

    <div class="flex flex-row mt-4 gap-2">
      <div class="grow" />
      <Button on:click={onClose} type="secondary">Cancel</Button>
      <Button on:click={handleSubmit} type="primary">Save</Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
