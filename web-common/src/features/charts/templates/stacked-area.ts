import { multiLayerBaseSpec } from "./utils";

export function buildStackedArea(
  timeField: string,
  quantitativeField: string,
  nominalField: string,
) {
  const baseSpec = multiLayerBaseSpec();

  baseSpec.encoding = {
    x: { field: timeField, type: "temporal" },
    y: {
      field: quantitativeField,
      type: "quantitative",
      stack: "zero",
    },
  };
  baseSpec.layer = [
    {
      mark: { type: "area", clip: true },
      encoding: {
        color: { field: nominalField, type: "nominal", legend: null },
      },
    },
    {
      mark: { type: "line", strokeWidth: 1, clip: true },
      encoding: {
        x: { field: timeField, type: "temporal" },
        y: {
          field: quantitativeField,
          type: "quantitative",
          stack: "zero",
        },
        stroke: { field: nominalField, type: "nominal", legend: null },
      },
    },
    {
      mark: { type: "rule", color: "transparent" },
      encoding: {
        x: { type: "temporal", field: timeField },
        tooltip: [
          { field: quantitativeField, type: "quantitative" },
          { field: "ts", type: "temporal", title: "Time" },
          { field: nominalField, type: "nominal" },
        ],
      },
      params: [
        {
          name: "x_hover",
          select: {
            type: "point",
            encodings: ["x"],
            nearest: true,
            on: "pointerover",
            clear: "pointerout",
          },
        },
      ],
    },
    {
      transform: [{ filter: { param: "x-hover", empty: false } }],
      mark: { type: "point", filled: true, opacity: 1, size: 40 },
      encoding: {
        x: { type: "temporal", field: timeField },
        y: { type: "quantitative", field: quantitativeField },
        color: { type: "nominal", field: nominalField },
      },
    },
  ];

  return baseSpec;
}
