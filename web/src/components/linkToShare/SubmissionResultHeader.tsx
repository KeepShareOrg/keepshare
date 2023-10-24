import { LinkFormatType } from "@/constant";
import { DownloadOutlined } from "@ant-design/icons";
import { Typography, Space, theme, Button } from "antd";
import { useState } from "react";
import { CustomCheckableTag } from "../sharedLinks/style";

const { Title, Text } = Typography;

const tagsData: LinkFormatType[] = [
  LinkFormatType.TEXT,
  LinkFormatType.HTML,
  LinkFormatType.BB_CODE,
];

interface ComponentProps {
  handleCopy: VoidFunction;
  handleDownload: VoidFunction;
  handleFormatChange: (formatType: LinkFormatType) => void;
}
const SubmissionResultHeader = ({
  handleCopy,
  handleDownload,
  handleFormatChange,
}: ComponentProps) => {
  const { token } = theme.useToken();

  const [selectedTag, setSelectedTag] = useState<LinkFormatType>(
    LinkFormatType.TEXT,
  );

  const handleTagChange = (tag: LinkFormatType) => {
    setSelectedTag(tag);
    handleFormatChange(tag);
  };

  return (
    <>
      <Space
        align="baseline"
        style={{ width: "100%", justifyContent: "space-between" }}
      >
        <Title level={5}>Submission Results</Title>
        <Space wrap>
          <Button onClick={handleCopy}>Copy</Button>
          <Button
            type="primary"
            icon={<DownloadOutlined />}
            onClick={handleDownload}
          >
            Download CSV
          </Button>
        </Space>
      </Space>
      <Space>
        <Text strong>Link Format</Text>
        <Space size={[0, 8]} wrap style={{ marginLeft: token.marginLG }}>
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
    </>
  );
};

export default SubmissionResultHeader;
