import {
  ResourceKind,
  fetchResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  getRuntimeServiceGetResourceQueryKey,
  type V1Resource,
  type V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, get, writable } from "svelte/store";
import { FileArtifact } from "./file-artifact";

export class FileArtifacts {
  private readonly artifacts: Map<string, FileArtifact> = new Map();
  readonly unsavedFiles = writable(new Set<string>());

  async init(queryClient: QueryClient, instanceId: string) {
    const resources = await fetchResources(queryClient, instanceId);
    for (const resource of resources) {
      switch (resource.meta?.name?.kind) {
        case ResourceKind.Source:
        case ResourceKind.Connector:
        case ResourceKind.Model:
        case ResourceKind.MetricsView:
        case ResourceKind.Component:
        case ResourceKind.Dashboard:
          // set query data for GetResource to avoid refetching data we already have
          queryClient.setQueryData(
            getRuntimeServiceGetResourceQueryKey(instanceId, {
              "name.name": resource.meta?.name?.name,
              "name.kind": resource.meta?.name?.kind,
            }),
            {
              resource,
            },
          );
          this.updateArtifacts(resource);
          break;
      }
    }
  }

  removeFile(filePath: string) {
    this.artifacts.delete(filePath);
  }

  resourceDeleted(name: V1ResourceName) {
    const artifact = this.findFileArtifact(
      (name.kind ?? "") as ResourceKind,
      name.name ?? "",
    );
    if (!artifact) return;

    this.removeFile(artifact.path);
  }

  updateArtifacts(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.getFileArtifact(filePath)?.updateAll(resource);
    });
  }

  updateReconciling(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.getFileArtifact(filePath)?.updateReconciling(resource);
    });
  }

  updateLastUpdated(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.getFileArtifact(filePath)?.updateLastUpdated(resource);
    });
  }

  /**
   * This is called when a resource is deleted either because file was deleted or it errored out.
   */
  softDeleteResource(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.getFileArtifact(filePath)?.softDeleteResource();
    });
  }

  getFileArtifact = (filePath: string) => {
    let artifact = this.artifacts.get(filePath);

    if (!artifact) {
      artifact = new FileArtifact(filePath);
      this.artifacts.set(filePath, artifact);
    }

    return artifact;
  };

  findFileArtifact(resKind: ResourceKind, resName: string) {
    for (const filePath in this.artifacts) {
      const artifact = this.artifacts.get(filePath);
      if (!artifact) continue;
      const name = get(artifact.name);
      if (name?.kind === resKind && name?.name === resName) {
        return this.artifacts.get(filePath);
      }
    }
    return undefined;
  }

  /**
   * Best effort list of all reconciling resources.
   */
  getReconcilingResourceNames() {
    const artifacts = Array.from(this.artifacts.values());
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
    return Array.from(this.artifacts.values())
      .filter((artifact) => get(artifact.name)?.kind === kind)
      .map((artifact) => get(artifact.name)?.name ?? "");
  }

  async saveAll() {
    await Promise.all(
      Array.from(this.artifacts.values()).map((artifact) =>
        artifact.saveLocalContent(),
      ),
    );
  }
}

export const fileArtifacts = new FileArtifacts();
