import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { extractFileName, getTopLevelFolder } from "@rilldata/utils";
import { fileIsMainEntity } from "@rilldata/web-common/features/entity-management/file-selectors";
import { eventBus } from "@rilldata/events";
import {
  runtimeServiceDeleteFile,
  runtimeServiceRenameFile,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { get } from "svelte/store";
import {
  FolderNameToResourceKind,
  addLeadingSlash,
  removeLeadingSlash,
} from "./entity-mappers";

export async function renameFileArtifact(
  instanceId: string,
  fromPath: string,
  toPath: string,
) {
  const fromName = extractFileName(fromPath);
  const toName = extractFileName(toPath);

  const fromFileArtifact = fileArtifacts.getFileArtifact(
    addLeadingSlash(fromPath),
  );
  const fromResName = get(fromFileArtifact.resourceName);

  if (fileIsMainEntity(fromPath)) {
    // try and copy over kind+name proactively for main entities (.sql,.yml,.yaml)
    // this fixes a flicker when renamed
    const toFileArtifact = fileArtifacts.getFileArtifact(
      addLeadingSlash(toPath),
    );
    if (!get(toFileArtifact.resourceName)) {
      // if there is no name set yet copy over from the source
      toFileArtifact.resourceName.set(fromResName);
    }
  }

  try {
    await runtimeServiceRenameFile(instanceId, {
      fromPath,
      toPath,
    });

    httpRequestQueue.removeByName(fromName);
    const topLevelFromFolder = getTopLevelFolder(addLeadingSlash(fromPath));
    const topLevelToFolder = getTopLevelFolder(addLeadingSlash(toPath));

    if (
      fromResName?.kind &&
      topLevelFromFolder !== topLevelToFolder &&
      FolderNameToResourceKind[removeLeadingSlash(topLevelFromFolder)] ===
        fromResName?.kind &&
      !toPath.endsWith(".sql")
    ) {
      eventBus.emit("notification", {
        message: `Moving ${fromName} out of its native folder. Make sure to specify the resource type with the "type" key.`,
      });
    }
  } catch (err) {
    eventBus.emit("notification", {
      message: `Failed to rename ${fromName} to ${toName}: ${extractMessage(err.response?.data?.message ?? err.message)}`,
    });
  }
}

export async function deleteFileArtifact(
  instanceId: string,
  filePath: string,
  force = false,
) {
  const name = extractFileName(filePath);
  try {
    await runtimeServiceDeleteFile(instanceId, {
      path: filePath,
      force,
    });

    httpRequestQueue.removeByName(name);
  } catch (err) {
    eventBus.emit("notification", {
      message: `Failed to delete ${name}: ${extractMessage(err.response?.data?.message ?? err.message)}`,
    });
  }
}

function extractMessage(msg: string) {
  if (msg.endsWith("directory not empty")) return "directory not empty";
  return msg;
}
