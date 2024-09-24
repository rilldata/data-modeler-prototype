<script lang="ts">
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { BookmarkIcon } from "lucide-svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import { createEventDispatcher } from "svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import { DropdownMenuItem } from "@rilldata/web-common/components/dropdown-menu";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";

  export let bookmark: BookmarkEntry;
  export let readOnly = false;

  const dispatch = createEventDispatcher();

  function selectBookmark() {
    dispatch("select", bookmark);
  }

  function editBookmark() {
    dispatch("edit", bookmark);
  }

  function deleteBookmark() {
    dispatch("delete", bookmark);
  }

  let hovered = false;
</script>

<DropdownMenuItem class="py-2">
  <div
    class="flex justify-between gap-x-2 w-full h-7 items-center"
    on:click={selectBookmark}
    on:keydown={(e) => e.key === "Enter" && e.currentTarget.click()}
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
    role="menuitem"
    tabindex="-1"
  >
    <div class="flex flex-row gap-x-2 items-center">
      {#if bookmark.resource.default}
        <HomeBookmark size="16px" />
      {:else if bookmark.filtersOnly}
        <Filter size="16px" />
      {:else}
        <BookmarkIcon size="16px" />
      {/if}
      <div class="flex flex-col">
        <div
          class="text-xs font-medium text-gray-700 h-4 text-ellipsis overflow-hidden"
        >
          {bookmark.resource.displayName}
        </div>
        {#if bookmark.resource.description}
          <div
            class="text-[11px] font-normal text-gray-500 h-4 text-ellipsis overflow-hidden"
          >
            {bookmark.resource.description}
          </div>
        {/if}
      </div>
    </div>
    {#if !readOnly}
      <div class="flex flex-row justify-end gap-x-2 items-start w-20">
        {#if hovered}
          <button
            on:click={editBookmark}
            class="bg-gray-100 hover:bg-primary-100 px-2 h-7 text-gray-400 hover:text-gray-500"
          >
            <EditIcon size="16px" />
          </button>
          <button
            on:click={deleteBookmark}
            class="bg-gray-100 hover:bg-primary-100 px-2 h-7 text-gray-400 hover:text-gray-500"
          >
            <Trash size="16px" />
          </button>
        {/if}
      </div>
    {/if}
  </div>
</DropdownMenuItem>
