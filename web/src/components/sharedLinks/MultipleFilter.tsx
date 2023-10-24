import type React from "react";
import { Drawer, Input, Popover, Space, Tag, theme } from "antd";
import { MultipleWrapper } from "./style";
import { ControlOutlined } from "@ant-design/icons";
import { useEffect, useRef, useState } from "react";
import AdvancedPane from "./AdvancedPane";
import {
  QueryCondition,
  SupportQueryKeys,
  supportQueryKeys,
} from "./FilterItem";
import { SupportOperatorValues } from "./filter.script";
import useStore from "@/store";
import dayjs from "dayjs";
import { useSearchParams } from "react-router-dom";

const { Search } = Input;

const transferFilterToDSLString = (
  { key, operator, value, unit }: QueryCondition,
  shim: boolean = false,
) => {
  // transform unit
  if (unit) {
    value = Array.isArray(value)
      ? value.map((c) => `"${c}${unit}"`)
      : `${value}${unit}`;
  }

  if (operator === "between" && Array.isArray(value)) {
    /* because server not support query by days_no_visit */
    // transfer days_not_visit to last_visited_at
    if (shim && key === "days_not_visit") {
      value = value.map(
        (v) =>
          `"${dayjs().subtract(Number(v), "day").format("YYYY-MM-DD HH:mm")}"`,
      );
      key = "last_visited_at";
      return `${key}<${value[0]} ${key}>${value[1]}`;
    }
    return `${key}>${value?.[0]} ${key}<${value?.[1]}`;
  }

  if (typeof value === "string" && !/^".+"$/.test(value)) {
    value = `"${value}"`;
  }

  /* because server not support query by days_no_visit */
  // transfer days_not_visit to last_visited_day
  if (shim && key === "days_not_visit") {
    value = `"${dayjs()
      .subtract(Number(value), "day")
      .format("YYYY-MM-DD HH:mm")}"`;
    key = "last_visited_at";
    if (operator === ">") {
      operator = "<";
    } else if (operator === "<") {
      operator = ">";
    }
  }
  return `${key}${operator}${value}`;
};

export const parseInputValue = (v: string): QueryCondition[] => {
  const keyRegString = `(${supportQueryKeys.join("|")})`;
  const operatorRegString = `\\s*(!?=|<=?|>=?|:)\\s*`;
  const valueRegString = `(?:"([^"]+)"|(\\d+))`;

  const filterRegString = `${keyRegString}${operatorRegString}${valueRegString}`;
  const filterReg = new RegExp(filterRegString, "ig");

  // 'title' is the default search key
  let haveValidTitleFilter = false;
  const newFilters = [...v.matchAll(filterReg)].map((v) => {
    const [, key, operator, value] = v.filter((x) => x);
    if (
      (key as SupportQueryKeys) === "title" &&
      (operator as SupportOperatorValues) === ":"
    ) {
      haveValidTitleFilter = true;
    }

    let unit = "";
    let conditionValue = /^\d+$/.test(value) ? Number(value) : value;
    // handle unit, current just ["GB"]
    if (/^(\d+)GB$/i.test(value)) {
      conditionValue = Number(value.match(/^(\d+)GB$/i)?.[1] || 0);
      unit = "GB";
    }
    if (typeof conditionValue === "string" && /^"(.*)"$/.test(conditionValue)) {
      conditionValue = conditionValue.match(/^"(.*)"$/)?.[1] || conditionValue;
    }
    return { key, operator, value: conditionValue, unit } as QueryCondition;
  });

  const remainingString = v.replace(filterReg, "").trim();
  if (remainingString && !haveValidTitleFilter) {
    newFilters.push({ key: "title", operator: ":", value: remainingString });
  }

  return newFilters;
};

interface ComponentProps {
  handleSearch: (search: string) => void;
}

// store advanced filters, before user click search or clean
let tempFilters: QueryCondition[] = [];
const MultipleFilter = ({ handleSearch }: ComponentProps) => {
  const { token } = theme.useToken();
  const [advancePaneVisible, setAdvancePaneVisible] = useState(false);

  const handleVisibleChange = () => setAdvancePaneVisible(!advancePaneVisible);

  const [inputValue, setInputValue] = useState("");
  const [filters, setFilters] = useState<QueryCondition[]>([]);

  const handleFilterChange = (condition: QueryCondition) => {
    if (!condition.key) {
      return;
    }

    tempFilters = tempFilters.filter(({ key }) => key !== condition.key);

    if (condition.operator === "between") {
      if (!Array.isArray(condition.value) || condition.value.length !== 2) {
        return;
      }
      // operator "between", have two filter
      tempFilters.push({
        key: condition.key,
        operator: ">",
        value: condition.value[0],
        unit: condition.unit,
      });
      tempFilters.push({
        key: condition.key,
        operator: "<",
        value: condition.value[1],
        unit: condition.unit,
      });
    } else {
      if (
        condition.operator !== "*" &&
        (condition.value || condition.value === 0)
      ) {
        tempFilters.push({
          key: condition.key,
          operator: condition.operator,
          value: condition.value,
          unit: condition.unit,
        });
      }
    }
  };

  const [filterTags, setFilterTags] = useState<string[]>([]);
  const triggerSearch = (newFilters: QueryCondition[]) => {
    setFilters(newFilters);

    const dsl = newFilters.map((v) => transferFilterToDSLString(v)).join(" ");
    setInputValue(dsl);
    const shimDsl = newFilters
      .map((v) => transferFilterToDSLString(v, true))
      .join(" ");
    handleSearch(shimDsl);

    const tags = newFilters.map(
      ({ key, operator, value, unit }) =>
        `${key}${operator}${value}${unit || ""}`,
    );
    setFilterTags(tags);
  };

  const handleToggleAdvancedPane = (open: boolean) => {
    setAdvancePaneVisible(open);
  };

  const handleInputSearch = (searchValue: string) => {
    const newFilters = parseInputValue(searchValue);
    triggerSearch(newFilters);
    tempFilters = JSON.parse(JSON.stringify(newFilters));
  };

  const handleAdvancedPaneSearch = () => {
    triggerSearch(tempFilters);
    setAdvancePaneVisible(false);
  };

  const handleClearFilters = () => {
    triggerSearch([]);
    setAdvancePaneVisible(false);
  };

  const handleFilterTagClose = (e: React.MouseEvent, idx: number) => {
    e.preventDefault();
    const newFilters = filters.filter((_, i) => i !== idx);
    triggerSearch(newFilters);
  };

  const wrapperRef = useRef(null);
  // eslint-disable-next-line
  const getPopoverContainer = () => wrapperRef.current as any;

  const isMobile = useStore((state) => state.isMobile);

  const [searchParams, setSearchParams] = useSearchParams();
  useEffect(() => {
    const search = searchParams.get("search");
    if (search) {
      setSearchParams("");
      search && handleInputSearch(search);
    }
  }, []);

  return (
    <div ref={wrapperRef}>
      <Popover
        getPopupContainer={getPopoverContainer}
        autoAdjustOverflow={false}
        overlayStyle={{
          position: "relative",
          width: "fit-content",
          height: 0,
          left: 0,
        }}
        content={
          isMobile || (
            <AdvancedPane
              filters={filters}
              handleFilterChange={handleFilterChange}
              handleClearFilters={handleClearFilters}
              handleSearch={handleAdvancedPaneSearch}
            />
          )
        }
        placement="bottomLeft"
        arrow={false}
        trigger={[]}
        open={isMobile ? false : advancePaneVisible}
        onOpenChange={handleToggleAdvancedPane}
      >
        <MultipleWrapper>
          <Space.Compact block size="large">
            <Search
              allowClear
              placeholder="Search in table"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onSearch={handleInputSearch}
              suffix={
                <ControlOutlined
                  onClick={handleVisibleChange}
                  style={{ cursor: "pointer" }}
                />
              }
            />
          </Space.Compact>
        </MultipleWrapper>
      </Popover>
      <Space style={{ marginTop: token.margin }}>
        {filterTags.map((v, i) => (
          <Tag
            closable
            key={`${v}${i}`}
            color="blue"
            onClose={(e) => handleFilterTagClose(e, i)}
          >
            {v}
          </Tag>
        ))}
      </Space>
      {isMobile && (
        <Drawer
          title="Advanced Search"
          placement="top"
          onClose={() => setAdvancePaneVisible(false)}
          open={advancePaneVisible}
          height={"100vh"}
        >
          <AdvancedPane
            filters={filters}
            handleFilterChange={handleFilterChange}
            handleClearFilters={handleClearFilters}
            handleSearch={handleAdvancedPaneSearch}
          />
        </Drawer>
      )}
    </div>
  );
};

export default MultipleFilter;
