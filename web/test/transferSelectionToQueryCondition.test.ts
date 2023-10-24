import { expect, test } from "vitest";
import {
  DAY_MS,
  YEAR_MS,
  transferSelectionToQueryCondition,
} from "../src/components/sharedLinks/filter.script";
import dayjs from "dayjs";

const testInputAndOutput = [
  {
    input: "[Any]",
    output: { key: "", operator: "*", value: "" },
  },
  {
    input: " <10",
    output: { key: "", operator: "<", value: 10 },
  },
  {
    input: "10 - 100",
    output: { key: "", operator: "between", value: [10, 100] },
  },
  {
    input: "50GB - 100GB",
    output: { key: "", operator: "between", value: [50, 100] },
  },
  {
    input: ">= 100GB",
    output: { key: "", operator: ">=", value: 100 },
  },
  {
    input: "Today",
    output: {
      key: "",
      operator: "between",
      value: [
        dayjs(Date.now() - DAY_MS).format("YYYY-MM-DD hh:mm"),
        dayjs(Date.now()).format("YYYY-MM-DD hh:mm"),
      ],
    },
  },
  {
    input: "Yesterday",
    output: {
      key: "",
      operator: "between",
      value: [
        dayjs(Date.now() - 2 * DAY_MS).format("YYYY-MM-DD hh:mm"),
        dayjs(Date.now() - DAY_MS).format("YYYY-MM-DD hh:mm"),
      ],
    },
  },
  {
    input: "Last 90 days",
    output: {
      key: "",
      operator: "between",
      value: [
        dayjs(Date.now() - 90 * DAY_MS).format("YYYY-MM-DD hh:mm"),
        dayjs(Date.now()).format("YYYY-MM-DD hh:mm"),
      ],
    },
  },
  {
    input: "Last 1 years",
    output: {
      key: "",
      operator: "between",
      value: [
        dayjs(Date.now() - YEAR_MS).format("YYYY-MM-DD hh:mm"),
        dayjs(Date.now()).format("YYYY-MM-DD hh:mm"),
      ],
    },
  },
];

testInputAndOutput.forEach((v) => {
  test("test transfer selection to queryCondition", () => {
    expect(transferSelectionToQueryCondition(v.input)).toEqual(v.output);
  });
});
