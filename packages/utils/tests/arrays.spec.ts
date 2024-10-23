import { createBatches } from "../src/index";
import { describe, expect, it } from "vitest";

describe("createBatches", () => {
  const Input = ["a", "b", "c", "d", "e", "f", "g", "h"];
  const TestCases = [
    {
      batchSize: 3,
      expected: [
        ["a", "b", "c"],
        ["d", "e", "f"],
        ["g", "h"],
      ],
    },
    {
      batchSize: 4,
      expected: [
        ["a", "b", "c", "d"],
        ["e", "f", "g", "h"],
      ],
    },
    {
      batchSize: 10,
      expected: [["a", "b", "c", "d", "e", "f", "g", "h"]],
    },
  ];

  for (const { batchSize, expected } of TestCases) {
    it(`input=${Input.join(",")} batchSize=${batchSize}`, () => {
      expect(createBatches(Input, batchSize)).toEqual(expected);
    });
  }
});
