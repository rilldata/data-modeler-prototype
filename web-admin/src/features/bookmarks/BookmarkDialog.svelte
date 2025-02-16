<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateBookmark,
    createAdminServiceUpdateBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import BaseBookmarkForm from "@rilldata/web-admin/features/bookmarks/BaseBookmarkForm.svelte";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { useProjectId } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { bool, object, string } from "yup";

  export let metricsViewName: string;
  export let exploreName: string;
  export let bookmark: BookmarkEntry | null = null;
  export let onClose = () => {};

  const StateManagers = getStateManagers();

  const bookmarkCreator = createAdminServiceCreateBookmark();
  const bookmarkUpdater = createAdminServiceUpdateBookmark();
  const timeControlsStore = useTimeControlStore(StateManagers);

  const schema = yup(
    object({
      displayName: string().required("Required"),
      description: string(),
      shared: string().oneOf(["true", "false"]),
      filtersOnly: bool(),
      absoluteTimeRange: bool(),
    }),
  );

  const initialValues = {
    displayName: bookmark?.resource.displayName || "Default Label",
    description: bookmark?.resource.description ?? "",
    shared: bookmark?.resource.shared ? "true" : "false",
    filtersOnly: bookmark?.filtersOnly ?? false,
    absoluteTimeRange: bookmark?.absoluteTimeRange ?? false,
  };

  const formState = superForm(defaults(initialValues, schema), {
    SPA: true,
    validators: schema,
    async onUpdate({ form }) {
      if (!form.valid) return;

      const values = form.data;

      if (bookmark) {
        await $bookmarkUpdater.mutateAsync({
          data: {
            bookmarkId: bookmark.resource.id,
            displayName: values.displayName,
            description: values.description,
            shared: values.shared === "true",
            data: getBookmarkDataForDashboard(
              $exploreState,
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
              $exploreState,
              values.filtersOnly,
              values.absoluteTimeRange,
              $timeControlsStore,
            ),
          },
        });
        formState.reset();
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
    validationMethod: "oninput",
  });

  $: ({ params } = $page);
  $: exploreState = useExploreState(exploreName);
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
      <Button on:click={formState.submit} type="primary">Save</Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
