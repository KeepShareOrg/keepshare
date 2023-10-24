import { type ChangeEvent, useState, useEffect } from "react";
import type { BasicItemType } from "./FilterItem";
import { Col, Input, Row, Typography, theme } from "antd";
import useStore from "@/store";

const { Text } = Typography;

export interface StringMatchFilterItemType extends BasicItemType {
  type: "string-match";
}
export const StringMatchFilterItem = ({
  title,
  searchKey,
  filters,
  handleFilterChange,
}: StringMatchFilterItemType) => {
  const [inputValue, setInputValue] = useState("");

  useEffect(() => {
    setInputValue(
      `${filters?.find(({ key }) => key === searchKey)?.value || ""}`,
    );
  }, [filters]);

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value);
    handleFilterChange?.({
      key: searchKey,
      operator: ":",
      value: e.target.value,
    });
  };

  const { token } = theme.useToken();
  const isMobile = useStore((state) => state.isMobile);

  return (
    <Row align="middle" style={{ width: "100%" }}>
      <Col xs={24} md={5}>
        <Text strong>{title}</Text>
      </Col>
      <Col
        xs={24}
        md={10}
        style={{ marginTop: isMobile ? token.marginXS : "0" }}
      >
        <Input
          placeholder={`Enter ${title}`}
          value={inputValue}
          allowClear
          onChange={handleInputChange}
        />
      </Col>
    </Row>
  );
};
