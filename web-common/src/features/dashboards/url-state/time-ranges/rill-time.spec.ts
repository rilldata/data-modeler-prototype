import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

describe("rill time", () => {
  describe("positive cases", () => {
    const Cases: [
      rillTime: string,
      label: string,
      rangeGrain: V1TimeGrain,
      bucketGrain: V1TimeGrain,
    ][] = [
      [
        "m : |s|",
        "Minute to date, incomplete",
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_SECOND,
      ],
      [
        "-5m : |m|",
        "Last 5 minutes, incomplete",
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_MINUTE,
      ],
      [
        "-5m, 0m : |m|",
        "Last 5 minutes",
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_MINUTE,
      ],
      [
        "-7d, 0d : |h|",
        "Last 7 days",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-7d, now/d : |h|",
        "Last 7 days",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-6d, now : |h|",
        "Last 6 days, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-6d, now : h",
        "Last 6 days, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "d : h",
        "Today, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],

      // TODO: correct label for the below
      [
        "-7d, -5d : h",
        "-7d, -5d : h",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-2d, now/d : h @ -5d",
        "-2d, now/d : h @ -5d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-2d, now/d @ -5d",
        "-2d, now/d @ -5d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],

      [
        "-7d, now/d : h @ {Asia/Kathmandu}",
        "-7d, now/d : h @ {Asia/Kathmandu}",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-7d, now/d : |h| @ {Asia/Kathmandu}",
        "-7d, now/d : |h| @ {Asia/Kathmandu}",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-7d, now/d : |h| @ -5d {Asia/Kathmandu}",
        "-7d, now/d : |h| @ -5d {Asia/Kathmandu}",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],

      // TODO: should these be something different when end is latest vs now?
      [
        "-7d, latest/d : |h|",
        "Last 7 days",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-6d, latest : |h|",
        "Last 6 days, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],
      [
        "-6d, latest : h",
        "Last 6 days, incomplete",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ],

      [
        "2024-01-01, 2024-03-31",
        "2024-01-01, 2024-03-31",
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "2024-01-01 10:00, 2024-03-31 18:00",
        "2024-01-01 10:00, 2024-03-31 18:00",
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],

      [
        "-7W+2d",
        "-7W+2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "-7W,latest+2d",
        "-7W,latest+2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "-7W,-7W+2d",
        "-7W,-7W+2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "-7W+2d",
        "-7W+2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "-7W + 2d",
        "-7W + 2d",
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
      [
        "2024-01-01-1W,2024-01-01",
        "2024-01-01-1W,2024-01-01",
        V1TimeGrain.TIME_GRAIN_WEEK,
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      ],
    ];

    const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
    for (const [rillTime, label, rangeGrain, bucketGrain] of Cases) {
      it(rillTime, () => {
        const parser = new nearley.Parser(compiledGrammar);
        parser.feed(rillTime);
        // assert that there is only match. this ensures unambiguous grammar.
        expect(parser.results).length(1);

        const rt = parseRillTime(rillTime);
        expect(rt.getLabel()).toEqual(label);
        expect(rt.getRangeGrain()).toEqual(rangeGrain);
        expect(rt.getBucketGrain()).toEqual(bucketGrain);

        const convertedRillTime = rt.toString();
        const convertedRillTimeParsed = parseRillTime(convertedRillTime);
        expect(convertedRillTimeParsed.toString()).toEqual(convertedRillTime);
      });
    }
  });

  describe("Update timezone", () => {
    const Cases: [rillTime: string, timezone: string, replaced: string][] = [
      ["-6d, latest : |h|", "UTC", "-6d,latest:|h|@{UTC}"],
      ["-6d, latest : |h| @ {IST}", "UTC", "-6d,latest:|h|@{UTC}"],
      ["-6d, latest : |h| @ now", "UTC", "-6d,latest:|h|@now {UTC}"],
      ["-6d, latest : |h| @ now {IST}", "UTC", "-6d,latest:|h|@now {UTC}"],
    ];

    for (const [rillTime, timezone, replaced] of Cases) {
      it(`'${rillTime}'.addTimezone(${timezone})=${replaced}`, () => {
        const rt = parseRillTime(rillTime);
        rt.addTimezone(timezone);
        expect(rt.toString()).toEqual(replaced);
      });
    }
  });
});
