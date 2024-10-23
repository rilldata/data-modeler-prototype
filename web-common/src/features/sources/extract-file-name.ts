import { extractFileName, sanitizeEntityName } from "@rilldata/utils";

export function getTableNameFromFile(filePath: string, name?: string) {
  return name ?? sanitizeEntityName(extractFileName(filePath));
}
