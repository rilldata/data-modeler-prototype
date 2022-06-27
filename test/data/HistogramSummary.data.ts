import type { NumericHistogramBin } from "$lib/types";

interface NumericHistogramTestCase {
    name: string,
    input: string,
    output: NumericHistogramBin[]
}

export const numericHistograms:NumericHistogramTestCase[] = [
    {
        name: 'three equidistant values',
        input: `
SELECT 1 as column
UNION ALL SELECT 4
UNION ALL SELECT 7`,
        output: [
            { bucket: 0, low: 1, high: 3, count: 1 },
            { bucket: 1, low: 3, high: 5, count: 1 },
            { bucket: 2, low: 5, high: 7, count: 1 }
          ]
    },
    {
    name: 'more values',
    input: `
SELECT 1 as column
UNION ALL SELECT 1
UNION ALL SELECT 1
UNION ALL SELECT 4
UNION ALL SELECT 4
UNION ALL SELECT 4
UNION ALL SELECT 4
UNION ALL SELECT 5
UNION ALL SELECT 7
UNION ALL SELECT 7
UNION ALL SELECT 7
    `,
        output: [
            { bucket: 0, low: 1, high: 2.5, count: 3 },
            { bucket: 1, low: 2.5, high: 4, count: 0 },
            { bucket: 2, low: 4, high: 5.5, count: 5 },
            { bucket: 3, low: 5.5, high: 7, count: 3 }
          ]
    }
];


export const dateHistograms:NumericHistogramTestCase[] = [
    {
    name: 'three DATES with more high',
        input: `
    SELECT DATE '1970-01-01'
    UNION ALL SELECT DATE '1970-01-02'
    UNION ALL SELECT DATE '1970-01-02'
        `,
        output: [
            { bucket: 0, low: 0, high: 43200, count: 1 },
            { bucket: 1, low: 43200, high: 86400, count: 2 }
          ]
    },
    {
        name: 'three DATES with more low',
        input: `
    SELECT DATE '1970-01-01'
    UNION ALL SELECT DATE '1970-01-01'
    UNION ALL SELECT DATE '1970-01-02'
        `,
        output: [
            { bucket: 0, low: 0, high: 43200, count: 2 },
            { bucket: 1, low: 43200, high: 86400, count: 1 }
          ]
    },
    {
        name: "many dates",
        input: `
        SELECT DATE '1970-01-01'
        UNION ALL SELECT DATE '1970-01-01'
        UNION ALL SELECT DATE '1970-01-02'
        UNION ALL SELECT DATE '1970-01-03'
        UNION ALL SELECT DATE '1970-01-03'
        UNION ALL SELECT DATE '1970-01-03'
        UNION ALL SELECT DATE '1970-01-03'
        UNION ALL SELECT DATE '1970-01-04'
        UNION ALL SELECT DATE '1970-01-05'
        UNION ALL SELECT DATE '1970-01-06'
        UNION ALL SELECT DATE '1970-01-06'
        `,
        output: [
            { bucket: 0, low: 0, high: 72000, count: 2 },
            { bucket: 1, low: 72000, high: 144000, count: 1 },
            { bucket: 2, low: 144000, high: 216000, count: 4 },
            { bucket: 3, low: 216000, high: 288000, count: 1 },
            { bucket: 4, low: 288000, high: 360000, count: 1 },
            { bucket: 5, low: 360000, high: 432000, count: 2 }
        ]
    },
    {
        name: "many timestamps",
        input: `
        SELECT DATE '1970-01-01'
        UNION ALL SELECT TIMESTAMP '1970-01-01'
        UNION ALL SELECT TIMESTAMP '1970-01-02'
        UNION ALL SELECT TIMESTAMP '1970-01-03'
        UNION ALL SELECT TIMESTAMP '1970-01-03'
        UNION ALL SELECT TIMESTAMP '1970-01-03'
        UNION ALL SELECT TIMESTAMP '1970-01-03'
        UNION ALL SELECT TIMESTAMP '1970-01-04'
        UNION ALL SELECT TIMESTAMP '1970-01-05'
        UNION ALL SELECT TIMESTAMP '1970-01-06'
        UNION ALL SELECT TIMESTAMP '1970-01-06'
        `,
        output: [
            { bucket: 0, low: 0, high: 72000, count: 2 },
            { bucket: 1, low: 72000, high: 144000, count: 1 },
            { bucket: 2, low: 144000, high: 216000, count: 4 },
            { bucket: 3, low: 216000, high: 288000, count: 1 },
            { bucket: 4, low: 288000, high: 360000, count: 1 },
            { bucket: 5, low: 360000, high: 432000, count: 2 }
        ]
    },
    {
        name: "second-level timestamps",
        input: `
    SELECT TIMESTAMP '1970-01-01 00:00:00'
    UNION ALL SELECT TIMESTAMP '1970-01-01 00:01:00'
    UNION ALL SELECT TIMESTAMP '1970-01-01 00:01:00'
        `,
        output: [
            { bucket: 0, low: 0, high: 30, count: 1 },
            { bucket: 1, low: 30, high: 60, count: 2 }
          ]
    }
];
// FIXME from Hamilton: I find the provider functionality a bit confusing and has
// too many layers of indirection to be comprehensible. I will re-approach this 
// at a later date. This should suffice as-is for now.
// export type HistogramDataProvider = DataProviderData<NumericHistogramTestCase>;
// export const numericHistogramTestData: HistogramDataProvider = {
//     title: "Numeric Histogram Test Data",
//     subData: numericHistograms.map(test => ({
//         title: test.name, args: test
//     }))
// }

// export const timestampHistogramTestData: HistogramDataProvider = {
//     title: "Timestamp Histogram Test Data",
//     subData: dateHistograms.map(test => ({
//         title: test.name, args: test
//     }))
// }
