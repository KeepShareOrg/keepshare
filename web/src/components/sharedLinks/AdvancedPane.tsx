import { Button, Divider, Space, Typography, theme } from "antd";
import { MutableRefObject, useRef } from "react";
import { AdvancedPaneWrapper } from "./style";
import {
  SupportOperatorLabels,
  operatorOptions,
  getStorageSelections,
  getStoredSelections,
  getVisitorsSelections,
} from "./filter.script";
import FilterItem, { FilterItemType, QueryCondition } from "./FilterItem";
import useStore from "@/store";
import { SharedLinkTableKey } from "@/constant";

const { Title } = Typography;

const filterItemList: FilterItemType[] = [
  {
    type: "string-match",
    title: SharedLinkTableKey.TITLE,
    searchKey: "title",
  },
  {
    type: "string-match",
    title: SharedLinkTableKey.ORIGINAL_LINKS,
    searchKey: "original_link",
  },
  {
    type: "string-match",
    title: SharedLinkTableKey.HOST_SHARED_LINK,
    searchKey: "host_shared_link",
  },
  {
    type: "time-duration",
    title: SharedLinkTableKey.CREATED_AT,
    searchKey: "created_at",
  },
  {
    type: "enum",
    title: SharedLinkTableKey.CREATED_BY,
    searchKey: "created_by",
    enumList: ["[Any]", "Auto Share", "Link to Share"],
  },
  {
    type: "enum",
    title: SharedLinkTableKey.STATE,
    searchKey: "state",
    enumList: ["[Any]", "Valid", "In Blacklist"],
  },
  {
    type: "multiple",
    title: SharedLinkTableKey.STORED,
    searchKey: "stored",
    operators: operatorOptions,
    getSelections: getStoredSelections,
    defaultOperator: SupportOperatorLabels.ANY,
  },
  {
    type: "multiple",
    title: SharedLinkTableKey.DAYS_NOT_VISIT,
    searchKey: "days_not_visit",
    operators: operatorOptions,
    getSelections: getStoredSelections,
    defaultOperator: SupportOperatorLabels.ANY,
  },
  {
    type: "multiple",
    title: SharedLinkTableKey.VISITOR,
    searchKey: "visitor",
    operators: operatorOptions,
    getSelections: getVisitorsSelections,
    defaultOperator: SupportOperatorLabels.ANY,
  },
  {
    type: "multiple",
    title: SharedLinkTableKey.SIZE,
    searchKey: "size",
    operators: operatorOptions,
    getSelections: getStorageSelections,
    defaultOperator: SupportOperatorLabels.ANY,
    unit: "GB",
  },
];

interface ComponentProps {
  filters: QueryCondition[];
  handleFilterChange: (filter: QueryCondition) => void;
  handleClearFilters: () => void;
  handleSearch: () => void;
}
interface FilterItemRefs {
  resetStatus: () => void;
  updateStatus: (filters: QueryCondition[]) => void;
}
const AdvancedPane = ({
  filters,
  handleFilterChange,
  handleClearFilters,
  handleSearch,
}: ComponentProps) => {
  const { token } = theme.useToken();

  const filterItemRefs = useRef<MutableRefObject<FilterItemRefs | undefined>[]>(
    [],
  );
  filterItemRefs.current = [];

  const isMobile = useStore((state) => state.isMobile);

  const wrapperMobileStyle: React.CSSProperties = {
    width: "100%",
    boxSizing: "border-box",
  };

  const wrapperPcStyle: React.CSSProperties = {
    paddingInline: token.marginLG,
    paddingBlock: token.marginSM,
  };

  return (
    <AdvancedPaneWrapper style={isMobile ? wrapperMobileStyle : wrapperPcStyle}>
      {isMobile || (
        <>
          <Title level={5}>Advanced Search</Title>
          <Divider style={{ marginBlock: token.marginSM }} />
        </>
      )}
      <Space direction="vertical" size={token.sizeXL} style={{ width: "100%" }}>
        {filterItemList.map((f, i) => {
          // eslint-disable-next-line
          const ref = useRef<FilterItemRefs>();
          filterItemRefs.current[i] = ref;
          return (
            <FilterItem
              key={i}
              filters={filters}
              handleFilterChange={handleFilterChange}
              {...f}
            />
          );
        })}
      </Space>
      <Divider style={{ marginBlock: token.marginSM }} />
      <Space style={{ width: "100%", justifyContent: "flex-end" }}>
        <Button onClick={handleClearFilters}>Clear</Button>
        <Button type="primary" onClick={handleSearch}>
          Search
        </Button>
      </Space>
    </AdvancedPaneWrapper>
  );
};

export default AdvancedPane;
