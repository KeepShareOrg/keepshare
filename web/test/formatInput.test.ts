import { expect, test } from "vitest";
import { type QueryCondition } from "../src/components/sharedLinks/FilterItem";
import { formatInputValue } from "../src/components/sharedLinks/MultipleFilter";

const sources: Array<{
  input: { text: string; filters: QueryCondition[] };
  output: [string, QueryCondition[]];
}> = [
  {
    input: {
      text: 'stored > "10GB" stored < "20GB" hello world',
      filters: [
        {
          key: "stored",
          operator: ">",
          value: "10GB",
        },
        {
          key: "stored",
          operator: "<",
          value: "20GB",
        },
      ],
    },
    output: [
      'title:"hello world" stored>"10GB" stored<"20GB"',
      [
        {
          key: "stored",
          operator: ">",
          value: "10GB",
        },
        {
          key: "stored",
          operator: "<",
          value: "20GB",
        },
        {
          key: "title",
          operator: ":",
          value: "hello world",
        },
      ],
    ],
  },
];

test("formatInput", () => {
  sources.forEach(({ input, output }) => {
    expect(formatInputValue(input.text, input.filters)).toEqual(output);
  });
});
