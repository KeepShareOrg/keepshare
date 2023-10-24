import { Space, Typography, theme } from "antd";
import PikPakIcon from "@/assets/images/PikPak.png";
import { SelectionWrapper, IconImg, CustomCheckableTag } from "./style";
import { useState } from "react";
import { LinkFormatType } from "@/constant";
import useStore from "@/store";

const { Text } = Typography;
interface ComponentProps {
  handleFormat: (formatType: LinkFormatType) => void;
}
const SelectionFilter = ({ handleFormat }: ComponentProps) => {
  const { token } = theme.useToken();
  const tagsData: LinkFormatType[] = [
    LinkFormatType.TEXT,
    LinkFormatType.HTML,
    LinkFormatType.BB_CODE,
  ];
  const [selectedTag, setSelectedTag] = useState<LinkFormatType>(
    LinkFormatType.TEXT,
  );

  const isMobile = useStore((state) => state.isMobile);

  const handleTagChange = (tag: LinkFormatType) => {
    setSelectedTag(tag);
    handleFormat(tag);
  };

  return (
    <SelectionWrapper style={{ flexWrap: isMobile ? "wrap" : "nowrap" }}>
      <Text strong>Host</Text>
      <IconImg src={PikPakIcon} alt="icon" />
      <Text>PikPak</Text>
      <Space
        size={[0, 8]}
        wrap
        style={{
          marginTop: isMobile ? token.margin : "0",
        }}
      >
        <Text strong style={{ marginLeft: isMobile ? "0" : token.marginLG }}>
          Link Format
        </Text>
        <Space style={{ marginLeft: token.marginLG }}>
          {tagsData.map((tag) => (
            <CustomCheckableTag
              key={tag}
              checked={selectedTag === tag}
              onChange={() => handleTagChange(tag)}
            >
              {tag}
            </CustomCheckableTag>
          ))}
        </Space>
      </Space>
    </SelectionWrapper>
  );
};

export default SelectionFilter;
