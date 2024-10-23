export const VALID_NAME_PATTERN = /^[^<>:"/\\|?*]+$/;

export const INVALID_CHARS = /[^a-zA-Z_\d]/g;

export const INVALID_NAME_MESSAGE =
  'Filename cannot contain special characters like /, <, >, :, ", \\, |, ?, or *. Please choose a different name.';

// sourced from https://www.steveruiz.me/posts/incrementing-name

// Will return "1" from "table_name_1"
const INCREMENT = new RegExp(/(\d+)$/);

/**
 * Get an incremented name (e.g. new_table_2) from a name (e.g. new_table), based on an array of
 * existing names.
 *
 * @param name The name to increment.
 * @param others The array of existing names.
 */
export function getName(name: string, others: string[]): string {
  const set = new Set(others.map((other) => other.toLowerCase()));

  let result = name;

  while (set.has(result.toLowerCase())) {
    result = INCREMENT.exec(result)?.[1]
      ? result.replace(INCREMENT, (m) => (+m + 1).toString())
      : `${result}_1`;
  }

  return result;
}

export function isDuplicateName(
  name: string,
  fromName: string,
  names: Array<string>,
) {
  if (name.toLowerCase() === fromName.toLowerCase()) return false;
  return names.findIndex((n) => n?.toLowerCase() === name.toLowerCase()) >= 0;
}

export function sanitizeEntityName(entityName: string): string {
  return entityName.replace(INVALID_CHARS, "_");
}

export function duplicateNameChecker(
  name: string,
  models: Array<string>,
  sources: Array<string>,
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
  sources: Array<string>,
) {
  return getName(name, [...models, ...sources]);
}

const nameSanitiserRegex = /[^\w-]/g;

export function sanitizeName(name: string) {
  return name.replace(nameSanitiserRegex, "-");
}
