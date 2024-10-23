import { extractFileExtension, splitFolderAndFileName } from "@rilldata/utils";
import {
  ResourceKind,
  useProjectParser,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  V1ReconcileStatus,
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGetFile,
  runtimeServicePutFile,
  type V1ParseError,
  type V1Resource,
  type V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient, QueryFunction } from "@tanstack/svelte-query";
import {
  derived,
  get,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import {
  DIRECTORIES_WITHOUT_AUTOSAVE,
  FILES_WITHOUT_AUTOSAVE,
} from "../editor/config";
import { fileArtifacts } from "./file-artifacts";
import { inferResourceKind } from "./infer-resource-kind";

const UNSUPPORTED_EXTENSIONS = [
  // Data formats
  ".db",
  ".db.wal",
  ".parquet",
  ".xls",
  ".xlsx",

  // Image formats
  ".png",
  ".jpg",
  ".jpeg",
  ".gif",
  ".svg",

  // Document formats
  ".pdf",
  ".doc",
  ".docx",
  ".ppt",
  ".pptx",
];

export class FileArtifact {
  readonly path: string;
  readonly resourceName = writable<V1ResourceName | undefined>(undefined);
  readonly inferredResourceKind = writable<ResourceKind | null | undefined>(
    undefined,
  );
  readonly reconciling = writable(false);
  readonly localContent: Writable<string | null> = writable(null);
  readonly remoteContent: Writable<string | null> = writable(null);
  readonly hasUnsavedChanges = writable(false);
  readonly fileExtension: string;
  readonly ready: Promise<boolean>;
  readonly fileTypeUnsupported: boolean;
  readonly folderName: string;
  readonly fileName: string;
  readonly disableAutoSave: boolean;
  readonly autoSave: Writable<boolean>;

  private remoteCallbacks = new Set<(content: string) => void>();
  private localCallbacks = new Set<(content: string | null) => void>();

  // Last time the state of the resource `kind/name` was updated.
  // This is updated in watch-resources and is used there to avoid
  // unnecessary calls to GetResource API.
  lastStateUpdatedOn: string | undefined;

  constructor(filePath: string) {
    const [folderName, fileName] = splitFolderAndFileName(filePath);

    this.path = filePath;
    this.folderName = folderName;
    this.fileName = fileName;

    this.disableAutoSave =
      FILES_WITHOUT_AUTOSAVE.includes(filePath) ||
      DIRECTORIES_WITHOUT_AUTOSAVE.includes(folderName);

    if (this.disableAutoSave) {
      this.autoSave = writable(false);
    } else {
      this.autoSave = localStorageStore<boolean>(`autoSave::${filePath}`, true);
    }

    this.fileExtension = extractFileExtension(filePath);
    this.fileTypeUnsupported = UNSUPPORTED_EXTENSIONS.includes(
      this.fileExtension,
    );

    this.onRemoteContentChange((content) => {
      const inferred = inferResourceKind(filePath, content);

      if (inferred) this.inferredResourceKind.set(inferred);
    });
  }

  updateRemoteContent = (newContent: string, alert = true) => {
    const hasNewContent = newContent !== get(this.remoteContent);
    this.remoteContent.set(newContent);
    if (alert && hasNewContent) {
      for (const callback of this.remoteCallbacks) {
        callback(newContent);
      }
    }
  };

  async fetchContent(invalidate = false) {
    const instanceId = get(runtime).instanceId;
    const queryParams = {
      path: this.path,
    };
    const queryKey = getRuntimeServiceGetFileQueryKey(instanceId, queryParams);

    if (invalidate) await queryClient.invalidateQueries(queryKey);

    const queryFn: QueryFunction<
      Awaited<ReturnType<typeof runtimeServiceGetFile>>
    > = ({ signal }) => runtimeServiceGetFile(instanceId, queryParams, signal);

    const { blob } = await queryClient.fetchQuery({
      queryKey,
      queryFn,
      staleTime: Infinity,
    });

    if (blob === undefined) {
      throw new Error("Content undefined");
    }

    this.updateRemoteContent(blob, true);
  }

  updateLocalContent = (content: string | null, alert = false) => {
    const hasUnsavedChanges = get(this.hasUnsavedChanges);
    const autoSave = get(this.autoSave);

    if (content === null) {
      this.hasUnsavedChanges.set(false);
      fileArtifacts.unsavedFiles.update((files) => {
        files.delete(this.path);
        return files;
      });
    } else if (!hasUnsavedChanges && !autoSave) {
      this.hasUnsavedChanges.set(true);
      fileArtifacts.unsavedFiles.update((files) => {
        files.add(this.path);
        return files;
      });
    }

    this.localContent.set(content);

    if (alert) {
      for (const callback of this.localCallbacks) {
        callback(content);
      }
    }
  };

  onRemoteContentChange = (callback: (content: string) => void) => {
    this.remoteCallbacks.add(callback);
    return () => this.remoteCallbacks.delete(callback);
  };

  onLocalContentChange = (callback: (content: string | null) => void) => {
    this.localCallbacks.add(callback);
    return () => this.localCallbacks.delete(callback);
  };

  revert = () => {
    this.updateLocalContent(null, true);
  };

  saveLocalContent = async () => {
    const blob = get(this.localContent);
    if (blob === null) return;

    await this.saveContent(blob);
  };

  saveContent = async (blob: string) => {
    const instanceId = get(runtime).instanceId;
    const key = getRuntimeServiceGetFileQueryKey(instanceId, {
      path: this.path,
    });

    queryClient.setQueryData(key, {
      blob,
    });

    try {
      await runtimeServicePutFile(instanceId, {
        path: this.path,
        blob,
      }).catch(console.error);

      // Optimistically update the remote content
      this.remoteContent.set(blob);

      this.updateLocalContent(null);
    } catch (e) {
      console.error(e);
    }
  };

  updateResource(resource: V1Resource) {
    this.updateResourceNameIfChanged(resource);
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
    this.reconciling.set(
      resource.meta?.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    );
  }

  hardDeleteResource() {
    // To avoid a workspace flicker, first infer the *intended* resource kind
    const inferred = inferResourceKind(
      this.path,
      get(this.remoteContent) ?? "",
    );

    const curName = get(this.resourceName);
    if (inferred) {
      this.inferredResourceKind.set(inferred);
    } else if (curName && curName.kind) {
      this.inferredResourceKind.set(curName.kind as ResourceKind);
    }

    this.resourceName.set(undefined);
    this.reconciling.set(false);
    this.lastStateUpdatedOn = undefined;
  }

  getResource = (queryClient: QueryClient, instanceId: string) => {
    return derived(this.resourceName, (name, set) =>
      useResource(
        instanceId,
        name?.name as string,
        name?.kind as ResourceKind,
        {
          queryClient,
        },
      ).subscribe(set),
    ) as ReturnType<typeof useResource<V1Resource>>;
  };

  getParseError = (queryClient: QueryClient, instanceId: string) => {
    return derived(
      useProjectParser(queryClient, instanceId),
      (projectParser) => {
        return projectParser.data?.projectParser?.state?.parseErrors?.find(
          (e) => e.filePath === this.path,
        );
      },
    );
  };

  getAllErrors = (
    queryClient: QueryClient,
    instanceId: string,
  ): Readable<V1ParseError[]> => {
    const store = derived(
      [
        useProjectParser(queryClient, instanceId),
        this.getResource(queryClient, instanceId),
      ],
      ([projectParser, resource]) => {
        if (projectParser.isFetching || resource.isFetching) {
          // to avoid flicker during a re-fetch, retain the previous value
          return get(store);
        }

        return [
          ...(
            projectParser.data?.projectParser?.state?.parseErrors ?? []
          ).filter((e) => e.filePath === this.path),
          ...(resource.data?.meta?.reconcileError
            ? [
                {
                  filePath: this.path,
                  message: resource.data.meta.reconcileError,
                },
              ]
            : []),
        ];
      },
      [],
    );
    return store;
  };

  getHasErrors(queryClient: QueryClient, instanceId: string) {
    return derived(
      this.getAllErrors(queryClient, instanceId),
      (errors) => errors.length > 0,
    );
  }

  private updateResourceNameIfChanged(resource: V1Resource) {
    const isSubResource = !!resource.component?.spec?.definedInCanvas;
    if (isSubResource) return;

    const curName = get(this.resourceName);

    // Much code currently assumes that a file is associated with 0 or 1 resource.
    // However, files for legacy Metrics Views generate 2 resources: a Metrics View and an Explore.
    // HACK: for files for legacy Metrics Views, ignore the Explore resource.
    if (
      curName?.kind === ResourceKind.MetricsView &&
      resource.meta?.name?.kind === ResourceKind.Explore
    ) {
      return;
    }

    const didResourceNameChange =
      curName?.name !== resource.meta?.name?.name ||
      curName?.kind !== resource.meta?.name?.kind;

    if (didResourceNameChange) {
      this.resourceName.set({
        kind: resource.meta?.name?.kind,
        name: resource.meta?.name?.name,
      });
    }
  }
}
