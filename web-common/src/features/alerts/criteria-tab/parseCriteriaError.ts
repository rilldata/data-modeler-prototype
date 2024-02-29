// criteria[0].value must be a `number` type, but the…inal value was: `NaN` (cast from the value `""`).

const criteriaParserRegex = /criteria\[(\d)*]\.(.*)/;

export function parseCriteriaError(errStr: string, index: number): string {
  const match = criteriaParserRegex.exec(errStr);
  if (!match) return "";
  const [, matchedIndex, matchedErr] = match;
  return Number(matchedIndex) === index ? matchedErr : "";
}
