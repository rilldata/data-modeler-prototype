export function removeIfExists<T>(array: Array<T>, checker: (e: T) => boolean) {
  const index = array.findIndex(checker);
  if (index >= 0) {
    array.splice(index, 1);
    return true;
  }
  return false;
}

export function getMapFromArray<T, K, V = T>(
  array: T[],
  keyGetter: (entity: T) => K,
  valGetter: (entity: T) => V = (e) => e as unknown as V,
): Map<K, V> {
  const map = new Map<K, V>();
  for (const entity of array) {
    map.set(keyGetter(entity), valGetter(entity));
  }
  return map;
}

export function createBatches<T>(array: T[], batchSize: number): T[][] {
  const batches: T[][] = [];
  for (let i = 0; i < array.length; i += batchSize) {
    batches.push(array.slice(i, i + batchSize));
  }
  return batches;
}
