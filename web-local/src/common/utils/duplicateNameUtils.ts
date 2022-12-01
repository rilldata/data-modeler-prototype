import { getName } from "@rilldata/web-local/common/utils/incrementName";

export function duplicateNameChecker(
  name: string,
  models: Array<string>,
  sources: Array<string>
) {
  const lowerName = name.toLowerCase();
  return (
    models.some((model) => model.toLowerCase() === lowerName) ||
    sources.some((source) => source.toLowerCase() === lowerName)
  );
}

export function incrementedNameGetter(
  name: string,
  models: Array<string>,
  sources: Array<string>
) {
  return getName(name, [...models, ...sources]);
}
