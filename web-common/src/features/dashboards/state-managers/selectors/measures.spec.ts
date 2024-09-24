import { getFilteredMeasuresAndDimensions } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

describe("measures selectors", () => {
  describe("getFilteredMeasuresAndDimensions", () => {
    const TestCases = [
      {
        title: "with unspecified grains, selected DAY",
        measures: ["mes", "mes_time_no_grain"],
        timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
        expected: {
          measures: ["mes", "mes_time_no_grain"],
          dimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
            },
          ],
        },
      },
      {
        title: "with unspecified and specified grains, selected DAY",
        measures: [
          "mes",
          "mes_time_no_grain",
          "mes_time_day_grain",
          "mes_time_week_grain",
        ],
        timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
        expected: {
          measures: ["mes", "mes_time_no_grain", "mes_time_day_grain"],
          dimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
            },
          ],
        },
      },
      {
        title: "with unspecified and specified grains, selected WEEK",
        measures: [
          "mes",
          "mes_time_no_grain",
          "mes_time_day_grain",
          "mes_time_week_grain",
        ],
        timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
        expected: {
          measures: ["mes", "mes_time_no_grain", "mes_time_week_grain"],
          dimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
            },
          ],
        },
      },
      {
        title: "with unspecified and specified grains, selected MONTH",
        measures: [
          "mes",
          "mes_time_no_grain",
          "mes_time_day_grain",
          "mes_time_week_grain",
        ],
        timeGrain: V1TimeGrain.TIME_GRAIN_MONTH,
        expected: {
          measures: ["mes", "mes_time_no_grain"],
          dimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
            },
          ],
        },
      },
    ];
    const MetricsView: V1MetricsViewSpec = {
      dimensions: [],
      measures: [
        {
          name: "mes",
          expression: "count(*)",
        },
        {
          name: "mes_time_no_grain",
          expression: "",
          requiredDimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
            },
          ],
        },
        {
          name: "mes_time_day_grain",
          expression: "",
          requiredDimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
            },
          ],
        },
        {
          name: "mes_time_week_grain",
          expression: "",
          requiredDimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
            },
          ],
        },
      ],
    };
    for (const { title, measures, timeGrain, expected } of TestCases) {
      it(title, () => {
        expect(
          getFilteredMeasuresAndDimensions({
            dashboard: {
              selectedTimeRange: {
                interval: timeGrain,
              },
            } as MetricsExplorerEntity,
          })(MetricsView, measures),
        ).toEqual(expected);
      });
    }
  });
});
