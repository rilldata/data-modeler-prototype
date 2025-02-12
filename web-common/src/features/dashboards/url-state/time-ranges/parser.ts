import type { RillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
export function parseRillTime(rillTimeRange: string): RillTime {
  const parser = new nearley.Parser(compiledGrammar);
  parser.feed(rillTimeRange);
  const rt = parser.results[0] as RillTime;
  rt.timeRange = rillTimeRange;
  return rt;
}

export function normaliseRillTime(rillTimeRange: string) {
  let normalisedRillTime = rillTimeRange;
  try {
    normalisedRillTime = parseRillTime(rillTimeRange).toString();
  } catch {
    // validation doesnt happen here
  }
  return normalisedRillTime;
}

export function validateRillTime(rillTime: string): Error | undefined {
  try {
    const parser = parseRillTime(rillTime);
    if (!parser) return new Error("Unknown error");
  } catch (err) {
    return err;
  }
  return undefined;
}
