import { goto } from "$app/navigation";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { getName } from "@rilldata/utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  PossibleFileExtensions,
  PossibleZipExtensions,
  fileHasValidExtension,
} from "@rilldata/web-common/features/sources/modal/possible-file-extensions";
import { overlay } from "@rilldata/web-common/layout/overlay-store";
import { eventBus } from "@rilldata/events";
import { runtimeServiceFileUpload } from "@rilldata/web-common/runtime-client/manual-clients";
import { getTableNameFromFile } from "../extract-file-name";
import {
  DuplicateActions,
  duplicateSourceAction,
  duplicateSourceName,
} from "../sources-store";

/**
 * Uploads all valid files.
 * If any file exists, a prompt is shown to resolve the duplicates.
 * Returns table name and file paths of all uploaded files.
 * Note: actual creation of the table with the file is not done by this method.
 */
export async function* uploadTableFiles(
  files: Array<File>,
  instanceId: string,
  goToIfSuccessful = true,
): AsyncGenerator<{ tableName: string; filePath: string }> {
  if (!files?.length) return;
  const { validFiles, invalidFiles } = filterValidFileExtensions(files);

  let lastTableName: string | undefined = undefined;
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];

  for (const validFile of validFiles) {
    // check if the file is already present. get the file and
    const resolvedTableName = await checkForDuplicate(
      validFile,
      (name) => {
        const lowerName = name.toLowerCase();
        return allNames.some((allName) => allName.toLowerCase() === lowerName);
      },
      (name) => getName(name, allNames),
    );
    // if there was a duplicate and cancel was clicked then we do not upload
    if (!resolvedTableName) continue;

    overlay.setDebounced({ title: `Uploading ${validFile.name}` });

    const filePath = await uploadFile(instanceId, validFile);
    // if upload failed for any reason continue
    if (filePath) {
      lastTableName = resolvedTableName;
      yield { tableName: resolvedTableName, filePath };
    }

    overlay.clear();
  }

  if (lastTableName && goToIfSuccessful) {
    await goto(`/files/sources/${lastTableName}.yaml`);
  }

  if (invalidFiles.length) {
    reportFileErrors(invalidFiles);
  }
}

function filterValidFileExtensions(files: Array<File>): {
  validFiles: Array<File>;
  invalidFiles: Array<File>;
} {
  const validFiles: File[] = [];
  const invalidFiles: File[] = [];

  files.forEach((file: File) => {
    if (fileHasValidExtension(file.name)) {
      validFiles.push(file);
    } else {
      invalidFiles.push(file);
    }
  });

  return { validFiles, invalidFiles };
}

/**
 * Checks if the file already exists.
 * If it does then prompt the user on what to do.
 * Return next available name with a number appended if user decides to keep both.
 * Return the table name extracted from file name in all other cases.
 */
async function checkForDuplicate(
  file: File,
  duplicateValidator: (name: string) => boolean,
  incrementedNameGetter: (name: string) => string,
): Promise<string | undefined> {
  const currentTableName = getTableNameFromFile(file.name);

  try {
    const isDuplicate = duplicateValidator(currentTableName);
    if (isDuplicate) {
      const userResponse = await getResponseFromModal(currentTableName);
      if (userResponse == DuplicateActions.Cancel) {
        return undefined;
      } else if (userResponse == DuplicateActions.KeepBoth) {
        return incrementedNameGetter(currentTableName);
      } else if (userResponse == DuplicateActions.Overwrite) {
        return currentTableName;
      }
    } else {
      return currentTableName;
    }
  } catch (err) {
    console.error(err);
  }

  return undefined;
}

export async function uploadFile(
  instanceId: string,
  file: File,
): Promise<string | undefined> {
  const formData = new FormData();
  formData.append("file", file);

  const filePath = `data/${file.name}`;

  try {
    await runtimeServiceFileUpload(instanceId, filePath, formData);
    return filePath;
  } catch (err) {
    console.error(err);
  }

  return undefined;
}

function reportFileErrors(invalidFiles: File[]) {
  eventBus.emit("notification", {
    message: `${invalidFiles.length} file${
      invalidFiles.length !== 1 ? "s are" : " is"
    } invalid: \n${invalidFiles.map((file) => file.name).join("\n")}`,
    detail:
      "Only .parquet, .csv, .tsv, .json, and .ndjson files are supported, along with their gzipped (.gz) counterparts",
    options: {
      persisted: true,
    },
  });
}

async function getResponseFromModal(
  currentTableName,
): Promise<DuplicateActions> {
  duplicateSourceName.set(currentTableName);

  return new Promise((resolve) => {
    const unsub = duplicateSourceAction.subscribe((action) => {
      if (action !== DuplicateActions.None) {
        setTimeout(unsub);
        duplicateSourceAction.set(DuplicateActions.None);
        resolve(action);
      }
    });
  });
}

export function openFileUploadDialog(multiple = true) {
  return new Promise<Array<File>>((resolve) => {
    const input = document.createElement("input");
    input.multiple = true;
    input.type = "file";
    /** an event callback when a source table file is chosen manually */
    input.onchange = (e: Event) => {
      const files = (<HTMLInputElement>e.target)?.files as FileList;
      if (files) {
        resolve(Array.from(files));
      } else {
        resolve([]);
      }
    };
    const focusHandler = () => {
      window.removeEventListener("focus", focusHandler);
      setTimeout(() => {
        resolve([]);
      }, 1000);
    };
    window.addEventListener("focus", focusHandler);
    input.multiple = multiple;
    input.accept = [...PossibleFileExtensions, ...PossibleZipExtensions].join(
      ",",
    );
    input.click();
  });
}
