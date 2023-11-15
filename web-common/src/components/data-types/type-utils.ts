/**
 * Describes the king of missing data that is being displayed.
 *
 * FIXME: rename to eg. MISSING_DATA, since it's not just for
 * percent diffs
 */
export enum PERC_DIFF {
  PREV_VALUE_ZERO = "prev_value_zero",
  PREV_VALUE_NULL = "prev_value_null",
  PREV_VALUE_NO_DATA = "prev_value_no_data",
  CURRENT_VALUE_NO_DATA = "current_value_no_data",
}
