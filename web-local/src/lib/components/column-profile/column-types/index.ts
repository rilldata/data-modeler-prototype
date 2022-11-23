import {
  CATEGORICALS,
  NUMERICS,
  TIMESTAMPS,
} from "@rilldata/web-local/lib/duckdb-data-types";
import NumericProfile from "./NumericProfile.svelte";
import TimestampProfile from "./TimestampProfile.svelte";
import VarcharProfile from "./VarcharProfile.svelte";

export function getColumnType(type) {
  if (CATEGORICALS.has(type)) return VarcharProfile;
  if (NUMERICS.has(type)) return NumericProfile;
  if (TIMESTAMPS.has(type)) return TimestampProfile;
}
