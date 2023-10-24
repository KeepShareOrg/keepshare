/* eslint-disable @typescript-eslint/no-explicit-any */
import type { MenuProps, SelectProps } from "antd";
import { QueryCondition } from "./FilterItem";
import dayjs from "dayjs";

export type HandleClickType = (...args: any) => void;
const generateSelections = (
  list: string[],
  handleClick: any,
): MenuProps["items"] =>
  list.map((v, i) => ({
    key: `${i}-${v}`,
    label: v,
    onClick: () => handleClick(v, i),
  }));

export const getStoredSelections = (handleClick: HandleClickType) =>
  generateSelections(
    ["[Any]", "<10", "10 - 100", "100 - 1000 ", "1000 - 10000", ">10000"],
    handleClick,
  );

export const getVisitorsSelections = (handleClick: HandleClickType) =>
  generateSelections(
    ["[Any]", "<10", "10 - 100", "100 - 1000 ", "1000 - 10000", ">10000"],
    handleClick,
  );

export const getStorageSelections = (handleClick: HandleClickType) =>
  generateSelections(
    ["[Any]", "<1GB", "1GB - 5GB", "10GB - 50GB", "50GB - 100GB", ">100GB"],
    handleClick,
  );

export const DAY_MS = 24 * 60 * 60 * 1000;
export const YEAR_MS = 365 * DAY_MS;
export const transferSelectionToQueryCondition = (
  selectionValue: string,
): QueryCondition => {
  const content = selectionValue.trim();
  const { key, operator, value }: QueryCondition = {
    key: "",
    operator: "*",
    value: "",
  };

  if (/\[any\]/i.test(content)) {
    return { key, operator: "*", value };
  }

  if (/today/i.test(content)) {
    return {
      key: "",
      operator: "between",
      value: [
        dayjs(Date.now() - DAY_MS).format("YYYY-MM-DD hh:mm"),
        dayjs(Date.now()).format("YYYY-MM-DD hh:mm"),
      ],
    };
  }

  if (/yesterday/i.test(content)) {
    return {
      key: "",
      operator: "between",
      value: [
        dayjs(Date.now() - 2 * DAY_MS).format("YYYY-MM-DD hh:mm"),
        dayjs(Date.now() - DAY_MS).format("YYYY-MM-DD hh:mm"),
      ],
    };
  }

  if (/last(.*)day/i.test(content)) {
    const [, days] = content.match(/last(.*)day/i) || [];
    const num = parseInt(days, 10);
    return {
      key: "",
      operator: "between",
      value: [
        dayjs(Date.now() - num * DAY_MS).format("YYYY-MM-DD hh:mm"),
        dayjs(Date.now()).format("YYYY-MM-DD hh:mm"),
      ],
    };
  }

  if (/last(.*)year/i.test(content)) {
    const [, years] = content.match(/last(.*)year/i) || [];
    const num = parseInt(years, 10);
    return {
      key: "",
      operator: "between",
      value: [
        dayjs(Date.now() - num * YEAR_MS).format("YYYY-MM-DD hh:mm"),
        dayjs(Date.now()).format("YYYY-MM-DD hh:mm"),
      ],
    };
  }

  if (/^(<=|>=|<|>)(.*)/i.test(content)) {
    const [, o, v] = content.match(/^(<=|>=|<|>)(.*)/) || [];
    const num = parseInt(v, 10);
    return { key, operator: o?.trim() as SupportOperatorValues, value: num };
  }

  if (/.*-.*/i.test(content)) {
    const [, s, e] = content.match(/^(.*)-(.*)$/i) || [];
    const startValue = parseInt(s?.trim() || s, 10);
    const endValue = parseInt(e?.trim() || e, 10);
    return { key, operator: "between", value: [startValue, endValue] };
  }

  return { key, operator, value };
};

export type SupportOperatorValues =
  (typeof SupportOperatorValueMap)[SupportOperatorLabels];
export const enum SupportOperatorLabels {
  ANY = "[Any]",
  EQUALS = "Equals",
  NOT_EQUALS = "Not Equals",
  GREATER_THAN = "Greater Than",
  GREATER_THAN_OR_EQUAL_TO = "Greater Than or Equal To",
  LESS_THAN = "Less Than",
  LESS_THAN_OR_EQUAL_TO = "Less Than or Equal To",
  BETWEEN = "Between",
  MATCH = "Match",
}
export const SupportOperatorValueMap = {
  [SupportOperatorLabels.ANY]: "*",
  [SupportOperatorLabels.EQUALS]: "=",
  [SupportOperatorLabels.NOT_EQUALS]: "!=",
  [SupportOperatorLabels.GREATER_THAN]: ">",
  [SupportOperatorLabels.GREATER_THAN_OR_EQUAL_TO]: ">=",
  [SupportOperatorLabels.LESS_THAN]: "<",
  [SupportOperatorLabels.LESS_THAN_OR_EQUAL_TO]: "<=",
  [SupportOperatorLabels.BETWEEN]: "between",
  [SupportOperatorLabels.MATCH]: ":",
} as const;
export const ValueSupportOperatorMap: Record<
  SupportOperatorValues,
  SupportOperatorLabels
> = {
  "*": SupportOperatorLabels.ANY,
  "=": SupportOperatorLabels.EQUALS,
  "!=": SupportOperatorLabels.NOT_EQUALS,
  ">": SupportOperatorLabels.GREATER_THAN,
  ">=": SupportOperatorLabels.GREATER_THAN_OR_EQUAL_TO,
  "<": SupportOperatorLabels.LESS_THAN,
  "<=": SupportOperatorLabels.LESS_THAN_OR_EQUAL_TO,
  between: SupportOperatorLabels.BETWEEN,
  ":": SupportOperatorLabels.MATCH,
};
export const operatorOptions: SelectProps["options"] = [
  {
    key: "0",
    label: SupportOperatorLabels.ANY,
    value: SupportOperatorValueMap[SupportOperatorLabels.ANY],
  },
  {
    key: "1",
    label: SupportOperatorLabels.EQUALS,
    value: SupportOperatorValueMap[SupportOperatorLabels.EQUALS],
  },
  {
    key: "2",
    label: SupportOperatorLabels.NOT_EQUALS,
    value: SupportOperatorValueMap[SupportOperatorLabels.NOT_EQUALS],
  },
  {
    key: "3",
    label: SupportOperatorLabels.GREATER_THAN,
    value: SupportOperatorValueMap[SupportOperatorLabels.GREATER_THAN],
  },
  {
    key: "4",
    label: SupportOperatorLabels.GREATER_THAN_OR_EQUAL_TO,
    value:
      SupportOperatorValueMap[SupportOperatorLabels.GREATER_THAN_OR_EQUAL_TO],
  },
  {
    key: "5",
    label: SupportOperatorLabels.LESS_THAN,
    value: SupportOperatorValueMap[SupportOperatorLabels.LESS_THAN],
  },
  {
    key: "6",
    label: SupportOperatorLabels.LESS_THAN_OR_EQUAL_TO,
    value: SupportOperatorValueMap[SupportOperatorLabels.LESS_THAN_OR_EQUAL_TO],
  },
  {
    key: "7",
    label: SupportOperatorLabels.BETWEEN,
    value: SupportOperatorValueMap[SupportOperatorLabels.BETWEEN],
  },
];
