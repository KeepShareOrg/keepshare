import { createShareLink } from "@/api/link";
import SubmissionResultTable from "@/components/linkToShare/SubmissionResultTable";
import useStore from "@/store";
import { parseLinks } from "@/util";
import { InfoCircleOutlined } from "@ant-design/icons";
import { Space, Typography, theme, Input, Button, message } from "antd";
import { useCallback, useState } from "react";

const { TextArea } = Input;
const { Paragraph, Title, Text } = Typography;
const LinkToShare = () => {
  const { token } = theme.useToken();
  const isMobile = useStore((state) => state.isMobile);

  const [linkContent, setLinkContent] = useState("");
  const [links, setLinks] = useState<string[]>([]);
  const [showSubmissionTable, setSubmissionTable] = useState(false);

  const handleCreateShareLink = useCallback(async () => {
    try {
      const links = parseLinks(linkContent);
      links.length > 0 && setLinks(links);

      const { error } = await createShareLink(links);

      if (error !== null) {
        message.error("create share links failed!");
        return;
      }
      message.success("create share links success!");
      setLinkContent("");
      setSubmissionTable(true);
    } catch (err) {
      console.log("create share link error: ", err);
    }
  }, [linkContent]);

  return (
    <Space direction="vertical">
      <Paragraph>
        <Space direction="vertical">
          <Title level={4}>
            Generate shared links from Magnet and other links in batches
          </Title>
          <Text>
            Remote uploading and sharing preparations will begin immediately
            after submission. You can post these created shared links, or you
            can also pre-create Auto-Sharing links so that peoples can get the
            shared files as soon as possible when accessing the keep sharing
            link.
          </Text>
          <TextArea
            value={linkContent}
            onChange={(e) => setLinkContent(e.target.value)}
            rows={8}
            placeholder="input link"
            style={{
              marginTop: token.margin,
              width: isMobile ? "100%" : "640px",
            }}
          />
          <Space style={{ alignItems: "flex-start", marginTop: token.margin }}>
            <InfoCircleOutlined style={{ color: token.colorTextSecondary }} />
            <Text style={{ color: token.colorTextSecondary }}>
              The shared link will not be recreated if the corresponding input
              link has been shared.
            </Text>
          </Space>
        </Space>
      </Paragraph>
      <Button type="primary" onClick={handleCreateShareLink}>
        Submit
      </Button>
      {showSubmissionTable && <SubmissionResultTable links={links} />}
    </Space>
  );
};

export default LinkToShare;
