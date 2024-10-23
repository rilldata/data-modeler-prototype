import { splitFolderAndFileName } from "@rilldata/utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
import { getPaddingFromPath } from "@rilldata/web-common/features/file-explorer/nav-tree-spacing";
import { get, writable } from "svelte/store";

export type NavDragData = {
  id: string;
  filePath: string;
  fileName?: string;
  isDir: boolean;
  kind?: ResourceKind | undefined;
};

export class NavEntryDragDropStore {
  private static readonly MIN_DRAG_DISTANCE = 9;
  /**
   * Timeout to wait while drag-drop and hovering over a folder to expand the folder
   */
  private static readonly DROP_HOVER_FOLDER_EXPAND_TIMEOUT = 1000;

  public readonly dragData = writable<null | NavDragData>(null);
  public initialPosition = { left: 0, top: 0 };
  public readonly position = writable({ left: 0, top: 0 });
  public offset = { x: 0, y: 0 };
  public readonly dropDirs = writable<Array<string>>([]);

  private newDragData: NavDragData | null;

  public onMouseDown(e: MouseEvent, dragData: NavDragData) {
    e.preventDefault();
    e.stopPropagation();

    const offsets = this.getOffsets(e, dragData);
    if (!offsets) return;
    const { left, top, x, y } = offsets;

    this.initialPosition = { left, top };

    this.offset = { x, y };

    const [, fileName] = splitFolderAndFileName(dragData.filePath);
    this.newDragData = {
      ...dragData,
      fileName,
    };
  }

  public async onMouseUp(
    e: MouseEvent,
    dropSuccess: (fromPath: string, toDir: string) => Promise<void>,
  ) {
    e.preventDefault();
    e.stopPropagation();

    const curDragData = get(this.dragData);
    const dropDirs = get(this.dropDirs);
    const toDir = dropDirs[dropDirs.length - 1];
    if (curDragData?.filePath && toDir) {
      await dropSuccess(curDragData.filePath, toDir);
    }

    this.resetDrag();
  }

  public onMouseMove(e: MouseEvent) {
    if (!this.newDragData) return;
    const left = e.clientX - this.offset.x;
    const top = e.clientY - this.offset.y;
    this.position.set({ left, top });

    if (get(this.dragData)) return;
    const dist = Math.sqrt(
      Math.pow(left - this.initialPosition.left, 2) +
        Math.pow(top - this.initialPosition.top, 2),
    );
    if (dist < NavEntryDragDropStore.MIN_DRAG_DISTANCE) return;
    this.dragData.set(this.newDragData);
  }

  public onMouseEnter(dir: string) {
    this.dropDirs.update((d) => {
      d.push(dir);
      return d;
    });
    if (!get(this.dragData)) return;
    setTimeout(
      () => this.expandDirectory(dir),
      NavEntryDragDropStore.DROP_HOVER_FOLDER_EXPAND_TIMEOUT,
    );
  }

  public onMouseLeave() {
    this.dropDirs.update((d) => {
      d.pop();
      return d;
    });
  }

  public resetDrag() {
    this.newDragData = null;
    this.dragData.set(null);
  }

  private getOffsets(e: MouseEvent, dragData: NavDragData) {
    const dragItem = document.getElementById(dragData.id);
    if (!dragItem) return;

    const { left, top } = dragItem.getBoundingClientRect();
    // 14 is the temporary offset for icon. we should add the icon eventually
    const effectiveLeft = left + getPaddingFromPath(dragData.filePath);

    return {
      left: effectiveLeft,
      top,
      x: e.clientX - effectiveLeft,
      y: e.clientY - top,
    };
  }

  private expandDirectory(dir: string) {
    const dropFolders = get(this.dropDirs);
    if (dir !== dropFolders[dropFolders.length - 1]) return;
    directoryState.expand(dir);
  }
}

export const navEntryDragDropStore = new NavEntryDragDropStore();
