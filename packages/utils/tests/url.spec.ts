import { getFullUrlForPath, getUrlForPath } from "../src/index";
import { describe, it, expect } from "vitest";

describe("url-utils", () => {
  describe("getFullUrlForPath", () => {
    const base = "http://localhost";
    const Variations: Array<[string, string, string]> = [
      ["/path/to/dashboard", "/new/path/dashboard", "/new/path/dashboard"],
      ["/path/to/dashboard", "new/path/dashboard", "/new/path/dashboard"],
      [
        "/path/to/dashboard/?features=all",
        "new/path/dashboard",
        "/new/path/dashboard?features=all",
      ],
      [
        "/path/to/dashboard?features=all&state=qwerty",
        "new/path/dashboard",
        "/new/path/dashboard?features=all",
      ],
      [
        "/path/to/dashboard?state=qwerty",
        "new/path/dashboard",
        "/new/path/dashboard",
      ],
    ];
    for (const [currentPath, newPath, expectedPath] of Variations) {
      it(`${currentPath} => ${expectedPath}`, () => {
        expect(getFullUrlForPath(new URL(base + currentPath), newPath)).toBe(
          expectedPath,
        );
      });
    }
  });

  it("getFullUrlForPath with explicit retainParam", () => {
    expect(
      getFullUrlForPath(
        new URL(
          "https://ui.rilldata.com/path/to/dashboard?features=all&state=qwerty&partner=asdfgh",
        ),
        "/new/path/to/dashboard",
        ["state", "partner"],
      ),
    ).toBe("/new/path/to/dashboard?state=qwerty&partner=asdfgh");
  });

  it("getFullUrl with https link", () => {
    expect(
      getUrlForPath(
        new URL("https://ui.rilldata.com/path/to/dashboard"),
        "/new/path/to/dashboard",
      ).toString(),
    ).toBe("https://ui.rilldata.com/new/path/to/dashboard");
  });
});
