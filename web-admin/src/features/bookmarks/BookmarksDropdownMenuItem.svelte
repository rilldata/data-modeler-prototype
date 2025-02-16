<script lang="ts">
  import type { BookmarkEntry } from "@rilldata/web-admin/features/bookmarks/selectors";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { BookmarkIcon, Edit2 as EditIcon } from "lucide-svelte";

  export let bookmark: BookmarkEntry;
  export let readOnly = false;
  export let onEdit: (bookmark: BookmarkEntry) => void;
  export let onDelete: (bookmark: BookmarkEntry) => Promise<void>;
  export let selected: boolean;

  let disableDelete = false;

  $: ({
    resource: { displayName, description, default: isDefault },
    filtersOnly,
  } = bookmark);

  function editBookmark() {
    onEdit(bookmark);
  }

  async function deleteBookmark(e: MouseEvent) {
    e.preventDefault();
    try {
      await onDelete(bookmark);
    } catch {
      // no-op
    }
    disableDelete = false;
  }
</script>

<DropdownMenu.Item
  class="flex gap-x-2.5 h-10 px-2.5 text-primary-600 group hover:bg-gray-100 {selected &&
    'outline outline-1 outline-gray-200 bg-slate-100'}"
  href={bookmark.url}
>
  {#if isDefault}
    <HomeBookmark size="18px" className="text-primary-600" />
  {:else if filtersOnly}
    <Filter size="18px" className="text-primary-600" />
  {:else}
    <BookmarkIcon size="18px" class="text-primary-600" />
  {/if}

  <div class="flex flex-col gap-y-0.5">
    <p class="text-xs font-medium text-gray-700 truncate">
      {displayName}
    </p>

    {#if description}
      <p class="text-[11px] text-gray-500 truncate">
        {description}
      </p>
    {/if}
  </div>

  {#if !readOnly}
    <div
      class="group-hover:flex hidden ml-auto flex-row justify-end gap-x-2 items-start w-20"
    >
      <Button type="ghost" square gray onClick={editBookmark}>
        <EditIcon size="15px" />
      </Button>
      <Button
        type="ghost"
        square
        gray
        disabled={disableDelete}
        onClick={deleteBookmark}
      >
        <Trash size="16px" />
      </Button>
    </div>
  {/if}
</DropdownMenu.Item>
