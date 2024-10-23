import {
  FolderNameToResourceKind,
  ResourceShortNameToResourceKind,
  removeLeadingSlash,
} from "./entity-mappers";
import { extractFileExtension } from "@rilldata/utils";
import { ResourceKind } from "./resource-selectors";

/**
 * When a file's reconciliation does not yield a resource, we guess the intended resource kind
 * in order to show the relevant Workspace UI.
 */
export function inferResourceKind(
  path: string,
  blob: string,
): ResourceKind | null {
  const fileExtension = extractFileExtension(path);

  // If it's not a YAML or SQL file, we don't know what it is
  if (fileExtension !== ".yaml" && fileExtension !== ".sql") {
    return null;
  }

  // If it's a SQL file, it's a model
  if (fileExtension === ".sql") {
    return ResourceKind.Model;
  }

  // If it is a YAML file...

  // Look at the file's content to see if it includes `type: <resource-kind>`
  const kindInText = findResourceKindInYAML(blob);
  if (kindInText) return kindInText;

  // Look at the first folder in the path
  const kindInPath = findResourceKindInFilePath(path);
  if (kindInPath) return kindInPath;

  return null;
}

const resourceKindRegex = /type\s*:\s*(\w+)/i;
function findResourceKindInYAML(text: string): ResourceKind | null {
  const match = text.match(resourceKindRegex);

  if (match) {
    const shortName = match[1].toLowerCase();
    return ResourceShortNameToResourceKind[shortName] ?? null;
  }

  return null;
}

function findResourceKindInFilePath(filePath: string): ResourceKind | null {
  const dirName = removeLeadingSlash(filePath).split("/")[0];
  return FolderNameToResourceKind[dirName] ?? null;
}
