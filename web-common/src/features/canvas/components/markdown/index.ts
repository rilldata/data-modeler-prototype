import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import { commonOptions } from "@rilldata/web-common/features/canvas/components/util";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { type ComponentCommonProperties } from "../types";

export { default as Markdown } from "./Markdown.svelte";

export interface MarkdownSpec extends ComponentCommonProperties {
  content: string;
}

export class MarkdownCanvasComponent extends BaseCanvasComponent<MarkdownSpec> {
  minSize = { width: 1, height: 1 };
  defaultSize = { width: 4, height: 2 };

  constructor(
    fileArtifact: FileArtifact,
    path: (string | number)[] = [],
    initialSpec: Partial<MarkdownSpec> = {},
  ) {
    const defaultSpec: MarkdownSpec = {
      title: "",
      description: "",
      content: "Your text",
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: MarkdownSpec): boolean {
    return typeof spec.content === "string" && spec.content.trim().length > 0;
  }

  inputParams(): Record<keyof MarkdownSpec, ComponentInputParam> {
    return {
      content: { type: "textArea" },
      ...commonOptions,
    };
  }

  newComponentSpec(): MarkdownSpec {
    return {
      content: "Markdown Text",
    };
  }
}
