import { expect, test } from "vitest";
import { parseLinks } from "../src/util";

test("parseLinks keeps ed2k file names with spaces", () => {
  const content = `
    ed2k://|file|The Matrix.avi|735934464|8549FD92D11E47E3D9C3379E5EC8F842|/
    magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567
    https://example.invalid/file.zip
  `;
  expect(parseLinks(content)).toEqual([
    "ed2k://|file|The Matrix.avi|735934464|8549FD92D11E47E3D9C3379E5EC8F842|/",
    "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567",
    "https://example.invalid/file.zip",
  ]);
});

test("parseLinks keeps following links after an unterminated ed2k link", () => {
  const broken = "ed2k://|file|Broken name.bin|1024|8549fd92d11e47e3d9c3379e5ec8f842";
  const magnet = "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567";
  const http = "https://example.invalid/file.zip";
  for (const separator of ["\n", " "]) {
    expect(parseLinks(`${broken}${separator}${magnet}\n${http}`)).toEqual([broken, magnet, http]);
  }
});

test("parseLinks preserves spaces in ed2k names and separates same-line links", () => {
  const ed2k = "ed2k://|file|A  B.bin|1024|8549fd92d11e47e3d9c3379e5ec8f842|/";
  const http = "https://example.invalid/file.zip";
  expect(parseLinks(`${ed2k}  ${http}`)).toEqual([ed2k, http]);
});

test("parseLinks splits whitespace separated links", () => {
  expect(
    parseLinks(
      "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567 https://example.invalid/file.zip",
    ),
  ).toEqual([
    "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567",
    "https://example.invalid/file.zip",
  ]);
});

test("parseLinks returns no empty entries", () => {
  expect(parseLinks(" \n\r\t ")).toEqual([]);
});
