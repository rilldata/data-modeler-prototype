import { parseKindAndNameFromFile } from "@rilldata/web-common/features/entity-management/file-content-utils";
import {
  extractFileName,
  extractFileExtension,
} from "@rilldata/web-common/features/entity-management/file-path-utils";
import {
  fetchFileContent,
  fetchMainEntityFiles,
} from "@rilldata/web-common/features/entity-management/file-selectors";
import {
  ResourceKind,
  fetchResources,
  useProjectParser,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  V1ReconcileStatus,
  type V1ParseError,
  type V1Resource,
  type V1ResourceName,
  getRuntimeServiceGetFileQueryKey,
  V1GetFileResponse,
  runtimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient, QueryObserverResult } from "@tanstack/svelte-query";
import {
  derived,
  get,
  type Readable,
  writable,
  type Writable,
} from "svelte/store";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";

const UNSUPPORTED_EXTENSIONS = [".parquet", ".db", ".db.wal"];

export class FileArtifact {
  readonly path: string;

  readonly name = writable<V1ResourceName | undefined>(undefined);

  readonly reconciling = writable<boolean>(false);

  /**
   * Used to check if a file has finished renaming.
   *
   * Reconciler uses meta.renamedFrom internally to track it.
   * It is unset once rename is complete.
   */
  public renaming = false;

  /**
   * Last time the state of the resource `kind/name` was updated.
   * This is updated in watch-resources and is used there to avoid unnecessary calls to GetResource API.
   */
  public lastStateUpdatedOn: string | undefined;

  public hasTable = false;

  public deleted = false;

  public localContent: Writable<string | null> = writable(null);
  public remoteContent: Writable<string | null> = writable(null);
  public hasUnsavedChanges = derived(this.localContent, (localContent) => {
    return localContent !== null;
  });
  public fileExtension: string;
  public fileQuery: Readable<QueryObserverResult<V1GetFileResponse, HTTPError>>;
  public ready: Promise<boolean>;
  private remoteCallbacks = new Set<(content: string) => void>();

  constructor(filePath: string) {
    this.path = filePath;

    this.fileExtension = extractFileExtension(filePath);
    const fileTypeUnsupported = UNSUPPORTED_EXTENSIONS.includes(
      this.fileExtension,
    );

    const queryKey = getRuntimeServiceGetFileQueryKey(get(runtime).instanceId, {
      path: filePath,
    });

    // Initial data fetch
    this.ready = fileTypeUnsupported
      ? Promise.resolve(false)
      : queryClient
          .fetchQuery({
            queryKey,
            queryFn: () =>
              runtimeServiceGetFile(get(runtime).instanceId, {
                path: filePath,
              }),
          })
          .then(({ blob }) => {
            if (blob === undefined) return false;
            this.updateRemoteContent(blob, false);
            return true;
          })
          .catch((e) => {
            console.error(e);
            return false;
          });
  }

  private updateRemoteContent(content: string, alert = true) {
    console.log("update remote", this.path);
    this.remoteContent.set(content);
    console.log({ alert, content });
    if (alert) {
      for (const callback of this.remoteCallbacks) {
        callback(content);
      }
    }
    // this.hasUnsavedChanges.set(get(this.localContent) !== content);
  }

  public onRemoteContentChange = (callback: (content: string) => void) => {
    this.remoteCallbacks.add(callback);
    return () => this.remoteCallbacks.delete(callback);
  };

  public updateLocalContent = (content: string | null) => {
    console.log("update local", this.path);
    this.localContent.set(content);
  };

  public updateAll(resource: V1Resource) {
    this.updateNameIfChanged(resource);
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
    this.reconciling.set(
      resource.meta?.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    );
    this.renaming = !!resource.meta?.renamedFrom;
    this.hasTable = resourceHasTable(resource);
    this.deleted = false;
  }

  public updateReconciling(resource: V1Resource) {
    this.updateNameIfChanged(resource);
    this.reconciling.set(
      resource.meta?.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    );
  }

  public updateLastUpdated(resource: V1Resource) {
    this.updateNameIfChanged(resource);
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
  }

  public softDeleteResource() {
    this.reconciling.set(false);
  }

  public getResource(queryClient: QueryClient, instanceId: string) {
    return derived(this.name, (name, set) =>
      useResource(
        instanceId,
        name?.name as string,
        name?.kind as ResourceKind,
        undefined,
        queryClient,
      ).subscribe(set),
    ) as ReturnType<typeof useResource<V1Resource>>;
  }

  public getAllErrors = (
    queryClient: QueryClient,
    instanceId: string,
  ): Readable<V1ParseError[]> => {
    const store = derived(
      [
        this.name,
        useProjectParser(queryClient, instanceId),
        this.getResource(queryClient, instanceId),
      ],
      ([name, projectParser, resource]) => {
        if (
          projectParser.isFetching ||
          resource.isFetching ||
          // retain old state while reconciling
          resource.data?.meta?.reconcileStatus ===
            V1ReconcileStatus.RECONCILE_STATUS_RUNNING ||
          resource.data?.meta?.reconcileStatus ===
            V1ReconcileStatus.RECONCILE_STATUS_PENDING
        ) {
          // to avoid flicker during a re-fetch, retain the previous value
          return get(store);
        }
        if (
          // not having a name will signify a non-entity file
          !name?.kind ||
          projectParser.isError
        ) {
          return [];
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

  public getHasErrors(queryClient: QueryClient, instanceId: string) {
    return derived(
      this.getAllErrors(queryClient, instanceId),
      (errors) => errors.length > 0,
    );
  }

  public revert = () => {
    this.updateLocalContent(null);
  };

  public saveLocalContent = async () => {
    const local = get(this.localContent);
    if (local === null) return;

    const blob = get(this.localContent);

    const instanceId = get(runtime).instanceId;
    const key = getRuntimeServiceGetFileQueryKey(instanceId, {
      path: this.path,
    });

    queryClient.setQueryData(key, {
      blob,
    });

    await runtimeServicePutFile(instanceId, {
      path: this.path,
      blob: local,
    }).catch(console.error);

    this.localContent.set(null);

    // this.hasUnsavedChanges.set(false);
  };

  public getEntityName() {
    return get(this.name)?.name ?? extractFileName(this.path);
  }

  private updateNameIfChanged(resource: V1Resource) {
    const isSubResource = !!resource.component?.spec?.definedInDashboard;
    if (isSubResource) return;
    const curName = get(this.name);
    if (
      curName?.name !== resource.meta?.name?.name ||
      curName?.kind !== resource.meta?.name?.kind
    ) {
      this.name.set({
        kind: resource.meta?.name?.kind,
        name: resource.meta?.name?.name,
      });
    }
  }
}

export class FileArtifacts {
  /**
   * Map of all files and its individual store
   */
  private readonly artifacts: Record<string, FileArtifact> = {};

  // Actions

  async init(queryClient: QueryClient, instanceId: string) {
    const resources = await fetchResources(queryClient, instanceId);
    for (const resource of resources) {
      switch (resource.meta?.name?.kind) {
        case ResourceKind.Source:
        case ResourceKind.Model:
        case ResourceKind.MetricsView:
        case ResourceKind.Component:
        case ResourceKind.Dashboard:
          this.updateArtifacts(resource);
          break;
      }
    }

    const files = await fetchMainEntityFiles(queryClient, instanceId);
    const missingFiles = files.filter(
      (f) => !this.artifacts[f] || !get(this.artifacts[f].name)?.kind,
    );

    await Promise.all(
      missingFiles.map((filePath) =>
        fetchFileContent(queryClient, instanceId, filePath).then(
          (fileContents) => {
            const artifact =
              this.artifacts[filePath] ?? new FileArtifact(filePath);
            const newName = parseKindAndNameFromFile(filePath, fileContents);
            if (newName) artifact.name.set(newName);
            this.artifacts[filePath] ??= artifact;
          },
        ),
      ),
    );
  }

  async fileUpdated(filePath: string) {
    // this.artifacts[filePath] ??= new FileArtifact(filePath);
    const fileContents = await fetchFileContent(
      queryClient,
      get(runtime).instanceId,
      filePath,
    );
    const newName = parseKindAndNameFromFile(filePath, fileContents);
    if (newName) this.artifacts[filePath].name.set(newName);
    console.log("FILE UPDATED");
    this.artifacts[filePath].updateRemoteContent(fileContents);
  }

  /**
   * This is called when an artifact is deleted.
   */
  fileDeleted(filePath: string) {
    // 2-way delete to handle race condition with delete from file and resource watchers
    // TODO: avoid this if `name` is undefined - event from resource will not be present
    if (this.artifacts[filePath]?.deleted) {
      // already marked for delete in resourceDeleted, delete from cache
      delete this.artifacts[filePath];
    } else if (this.artifacts[filePath]) {
      // seeing delete for the 1st time, mark for delete
      this.artifacts[filePath].deleted = true;
    }
  }

  resourceDeleted(name: V1ResourceName) {
    const artifact = this.findFileArtifact(
      (name.kind ?? "") as ResourceKind,
      name.name ?? "",
    );
    if (!artifact) return;
    // 2-way delete to handle race condition with delete from file and resource watchers
    if (artifact.deleted) {
      // already marked for delete in fileDeleted, delete from cache
      delete this.artifacts[artifact.path];
    } else {
      // seeing delete for the 1st time, mark for delete
      artifact.deleted = true;
    }
  }

  updateArtifacts(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      // this.artifacts[filePath] ??= new FileArtifact(filePath);
      this.artifacts[filePath].updateAll(resource);
    });
  }

  updateReconciling(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      // this.artifacts[filePath] ??= new FileArtifact(filePath);
      this.artifacts[filePath].updateReconciling(resource);
    });
  }

  updateLastUpdated(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      // this.artifacts[filePath] ??= new FileArtifact(filePath);
      this.artifacts[filePath].updateLastUpdated(resource);
    });
  }

  tableStatusChanged(resource: V1Resource) {
    const hadTable =
      resource.meta?.filePaths?.some((filePath) => {
        return this.artifacts[filePath].hasTable;
      }) ?? false;
    const hasTable = resourceHasTable(resource);
    return hadTable !== hasTable;
  }

  wasRenaming(resource: V1Resource) {
    const finishedRename = !resource.meta?.renamedFrom;
    return (
      resource.meta?.filePaths?.some((filePath) => {
        return this.artifacts[filePath].renaming && finishedRename;
      }) ?? false
    );
  }

  /**
   * This is called when a resource is deleted either because file was deleted or it errored out.
   */
  softDeleteResource(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.artifacts[filePath]?.softDeleteResource();
    });
  }

  // Selectors

  getFileArtifact(filePath: string) {
    this.artifacts[filePath] ??= new FileArtifact(filePath);
    return this.artifacts[filePath];
  }

  findFileArtifact(resKind: ResourceKind, resName: string) {
    for (const filePath in this.artifacts) {
      const name = get(this.artifacts[filePath].name);
      if (name?.kind === resKind && name?.name === resName) {
        return this.artifacts[filePath];
      }
    }
    return undefined;
  }

  /**
   * Best effort list of all reconciling resources.
   */
  getReconcilingResourceNames() {
    const artifacts = Object.values(this.artifacts);
    return derived(
      artifacts.map((a) => a.reconciling),
      (reconcilingArtifacts) => {
        const currentlyReconciling = new Array<V1ResourceName>();
        reconcilingArtifacts.forEach((reconcilingArtifact, i) => {
          if (reconcilingArtifact) {
            currentlyReconciling.push(get(artifacts[i].name) as V1ResourceName);
          }
        });
        return currentlyReconciling;
      },
    );
  }

  /**
   * Filters all fileArtifacts based on kind param and returns the file paths.
   * This can be expensive if the project gets large.
   * If we ever need this reactively then we should look into caching this list.
   */
  getNamesForKind(kind: ResourceKind): string[] {
    return Object.values(this.artifacts)
      .filter((artifact) => get(artifact.name)?.kind === kind)
      .map((artifact) => get(artifact.name)?.name ?? "");
  }
}

function resourceHasTable(resource: V1Resource) {
  return (
    (!!resource.model && !!resource.model.state?.resultTable) ||
    (!!resource.source && !!resource.source.state?.table)
  );
}

export const fileArtifacts = new FileArtifacts();
