import {
  extractFileExtension,
  getTableNameFromFile,
} from "$lib/util/extract-table-name";
import { FILE_EXTENSION_TO_TABLE_TYPE } from "$lib/types";
import notifications from "$lib/components/notifications";
import {
  config,
  DuplicateActions,
  duplicateSourceAction,
  duplicateSourceName,
} from "$lib/application-state-stores/application-store";
import { importOverlayVisible } from "$lib/application-state-stores/layout-store";

/**
 * uploadTableFiles
 * --------
 * Attempts to upload all files passed in.
 * Will return the list of files that are not valid.
 */
export function uploadTableFiles(files, apiBase: string) {
  const invalidFiles = [];
  const validFiles = [];

  [...files].forEach((file: File) => {
    const fileExtension = extractFileExtension(file.name);
    if (fileExtension in FILE_EXTENSION_TO_TABLE_TYPE) {
      validFiles.push(file);
    } else {
      invalidFiles.push(file);
    }
  });

  validFiles.forEach((validFile) => validateFile(validFile, apiBase));
  return invalidFiles;
}

export function validateFile(file: File, apiBase: string) {
  const tableUploadURL = `${apiBase}/table-upload`;
  const tableValidateURL = `${apiBase}/validate-table`;

  const currentTableName = getTableNameFromFile(file.name);
  fetch(tableValidateURL + `?tableName=${currentTableName}`)
    .then((response) => response.json())
    .then(async (d) => {
      if (d.isDuplicate) {
        const userResponse = await getResponseFromModal(currentTableName);
        duplicateSourceAction.set(DuplicateActions.None);
        if (userResponse == DuplicateActions.Cancel) {
          return;
        } else if (userResponse == DuplicateActions.KeepBoth) {
          uploadFile(file, tableUploadURL, d.name);
        } else if (userResponse == DuplicateActions.Overwrite) {
          uploadFile(file, tableUploadURL);
        }
      } else {
        uploadFile(file, tableUploadURL);
      }
    })
    .catch((...args) => console.error(...args));
}

export function uploadFile(file: File, url: string, tableName?: string) {
  importOverlayVisible.set(true);

  const formData = new FormData();
  formData.append("file", file);

  if (tableName) {
    formData.append("tableName", tableName);
  }

  fetch(url, {
    method: "POST",
    body: formData,
  })
    .then((...args) => console.error(...args))
    .catch((...args) => console.error(...args))
    .finally(() => importOverlayVisible.set(false));
}

function reportFileErrors(invalidFiles: File[]) {
  notifications.send({
    message: `${invalidFiles.length} file${
      invalidFiles.length !== 1 ? "s are" : " is"
    } invalid: \n${invalidFiles.map((file) => file.name).join("\n")}`,
    options: {
      width: 400,
    },
  });
}

/** Handles the uploading of the datasets. Any invalid files will be reported
 * through reportFileErrors.
 */
export function handleFileUploads(filesArray: File[]) {
  let invalidFiles = [];
  if (filesArray) {
    invalidFiles = uploadTableFiles(
      filesArray,
      `${config.server.serverUrl}/api`
    );
  }
  if (invalidFiles.length) {
    importOverlayVisible.set(false);
    reportFileErrors(invalidFiles);
  }
}

/** a drag and drop callback to kick off a source table import */
export function onSourceDrop(e: DragEvent) {
  const files = e?.dataTransfer?.files;
  if (files) {
    handleFileUploads(Array.from(files));
  }
}

/** an event callback when a source table file is chosen manually */
export function onManualSourceUpload(e: Event) {
  const files = (<HTMLInputElement>e.target)?.files as FileList;
  if (files) {
    handleFileUploads(Array.from(files));
  }
}

export async function uploadFilesWithDialog() {
  const input = document.createElement("input");
  input.multiple = true;
  input.type = "file";
  input.onchange = onManualSourceUpload;
  input.click();
}

async function getResponseFromModal(
  currentTableName
): Promise<DuplicateActions> {
  duplicateSourceName.set(currentTableName);

  return new Promise((resolve) => {
    const unsub = duplicateSourceAction.subscribe((action) => {
      if (action !== DuplicateActions.None) {
        unsub();
        resolve(action);
      }
    });
  });
}
