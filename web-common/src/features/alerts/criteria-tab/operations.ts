import { V1Operation } from "@rilldata/web-common/runtime-client";

// TODO: should match measure filter. remove this once that is merged to main
export const MeasureFilterOptions = [
  {
    value: V1Operation.OPERATION_LT,
    label: "Less Than",
    shortLabel: "<",
  },
  {
    value: V1Operation.OPERATION_LTE,
    label: "Less Than Or Equals",
    shortLabel: "<=",
  },
  {
    value: V1Operation.OPERATION_GT,
    label: "Greater Than",
    shortLabel: ">",
  },
  {
    value: V1Operation.OPERATION_GTE,
    label: "Greater Than Or Equals",
    shortLabel: ">=",
  },
];
