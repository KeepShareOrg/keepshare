import { match } from "ts-pattern";
import { SupportOperatorValues } from "./filter.script";
import {
  TimeDurationFilterItem,
  type TimeDurationFilterItemType,
} from "./TimeDurationFilterItem";
import {
  StringMatchFilterItem,
  type StringMatchFilterItemType,
} from "./StringMatchFIlterItem";
import { EnumFilterItem, type EnumFilterItemType } from "./EnumFilterItem";
import {
  MultipleFilterItem,
  type MultipleFilterItemType,
} from "./MultipleFilterItem";

export const supportQueryKeys = [
  "title",
  "original_link",
  "host_shared_link",
  "created_at",
  "created_by",
  "state",
  "stored",
  "days_not_visit",
  "visitor",
  "size",
] as const;
export type SupportQueryKeys = (typeof supportQueryKeys)[number];

export interface QueryCondition {
  key: string;
  operator: SupportOperatorValues;
  value: string | number | string[] | number[];
  unit?: string;
}

export interface BasicItemType {
  title: string;
  searchKey: SupportQueryKeys;
  filters?: QueryCondition[];
  handleFilterChange?: (condition: QueryCondition) => void;
  unit?: string;
}

export type FilterItemType =
  | StringMatchFilterItemType
  | TimeDurationFilterItemType
  | EnumFilterItemType
  | MultipleFilterItemType;

const FilterItem = (params: FilterItemType) => {
  return match(params)
    .with({ type: "enum" }, (p) => <EnumFilterItem {...p} />)
    .with({ type: "string-match" }, (p) => <StringMatchFilterItem {...p} />)
    .with({ type: "multiple" }, (p) => <MultipleFilterItem {...p} />)
    .with({ type: "time-duration" }, (p) => <TimeDurationFilterItem {...p} />)
    .exhaustive();
};

export default FilterItem;
