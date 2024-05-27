import { markdown } from "@codemirror/lang-markdown";
import { yaml } from "@rilldata/web-common/components/editor/presets/yaml";
import { extractFileExtension } from "@rilldata/web-common/features/entity-management/file-path-utils";

export const FileExtensionToEditorExtension = {
  ".yaml": yaml(),
  ".yml": yaml(),
  ".md": [markdown()],
};

export function getExtensionsForFile(filePath: string) {
  const extension = extractFileExtension(filePath);
  return FileExtensionToEditorExtension[extension] ?? [];
}
