import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../../runtime-client";

export function useProjectTitle(instanceId: string) {
  return createRuntimeServiceGetFile(
    instanceId,
    { path: "rill.yaml" },
    {
      query: {
        select: (data) => {
          if (!data.blob) return "";

          let projectData: { title?: string; name?: string } = {};
          try {
            projectData = parse(data.blob, {
              logLevel: "silent",
            }) as {
              title?: string;
              name?: string;
            };
          } catch (e) {
            // Ignore
          }

          return (
            projectData.title || projectData.name || "Untitled Rill Project"
          );
        },
      },
    },
  );
}
