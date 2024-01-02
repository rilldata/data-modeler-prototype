import type { NumberParts } from "../humanizer-types";

export const numberPartsToString = (parts: NumberParts): string =>
  (parts.neg || "") +
  (parts.currencySymbol || "") +
  parts.int +
  parts.dot +
  parts.frac +
  parts.suffix +
  (parts.percent || "");

export function numStrToParts(numStr: string): NumberParts {
  const nonNumReMatch = numStr.match(/[a-zA-z ]/);
  let int = "";
  const dot: "" | "." = numStr.includes(".") ? "." : "";
  let frac = "";
  let suffix = "";
  if (nonNumReMatch) {
    const suffixIndex = nonNumReMatch.index;
    const numPart = numStr.slice(0, suffixIndex);
    suffix = numStr.slice(suffixIndex);

    if (numPart.split(".").length == 1) {
      int = numPart;
    } else {
      int = numPart.split(".")[0];
      frac = numPart.split(".")[1] ?? "";
    }
  } else {
    int = numStr.split(".")[0];
    frac = numStr.split(".")[1] ?? "";
  }
  if (suffix === undefined || int === undefined || frac === undefined) {
    console.error({ numStr, int, frac, suffix });
  }
  return { int, dot, frac, suffix };
}
