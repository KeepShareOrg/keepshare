import { Row, Col, Space, Typography, theme } from "antd";
import type { BasicItemType } from "./FilterItem";
import { useEffect, useState } from "react";
import useStore from "@/store";
import { CustomCheckableTag } from "./style";

const { Text } = Typography;
export interface EnumFilterItemType extends BasicItemType {
  type: "enum";
  enumList: string[];
}
export const EnumFilterItem = ({
  title,
  searchKey,
  filters,
  enumList,
  handleFilterChange,
}: EnumFilterItemType) => {
  const [selectedStateTag, setSelectedStateTag] = useState<string>(enumList[0]);

  useEffect(() => {
    const filter = filters?.find(({ key }) => key === searchKey);
    const value = filter?.value as string;

    filter || setSelectedStateTag(enumList[0]);
    enumList.includes(value) && setSelectedStateTag(value);
  }, [filters]);

  const handleSelectChange = (tag: string) => {
    setSelectedStateTag(tag);
    handleFilterChange?.({ key: searchKey, operator: "=", value: tag });
  };

  const { token } = theme.useToken();
  const isMobile = useStore((state) => state.isMobile);

  return (
    <Row align="middle">
      <Col xs={24} md={5}>
        <Text strong>{title}</Text>
      </Col>
      <Col xs={24} md={10} style={{ marginTop: isMobile ? token.marginXS : 0 }}>
        <Space size={[0, 8]} wrap>
          {enumList.map((tag) => (
            <CustomCheckableTag
              key={tag}
              checked={selectedStateTag === tag}
              onChange={() => handleSelectChange(tag)}
            >
              {tag}
            </CustomCheckableTag>
          ))}
        </Space>
      </Col>
    </Row>
  );
};
