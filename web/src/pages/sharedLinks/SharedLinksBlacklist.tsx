import { addToBlacklist } from "@/api/link";
import BlackListTable from "@/components/sharedLinks/BlacklistTable";
import useStore from "@/store";
import { parseLinks } from "@/util";
import { Space, Typography, Input, theme, Button, message } from "antd";
import { useCallback, useState } from "react";

const { TextArea } = Input;
const { Paragraph, Title } = Typography;

const SharedLinksBlacklist = () => {
  const { token } = theme.useToken();
  const [linkContent, setLinkContent] = useState("");
  const [links, setLinks] = useState<string[]>([]);

  const handleAddToBlacklist = useCallback(async () => {
    try {
      const tempLinks = parseLinks(linkContent);
      tempLinks.length > 0 && setLinks(tempLinks);

      const { error } = await addToBlacklist(links);
      if (error) {
        message.error(error.message);
        return;
      }
      message.success("add to blacklist success!");
    } catch (err) {
      console.error("add to blacklist error: ", err);
    }
  }, [linkContent]);

  const isMobile = useStore((state) => state.isMobile);

  return (
    <>
      <Paragraph>
        <Space direction="vertical">
          <Title level={4}>
            KeepShare allows you to prevent certain links from Auto-Share, you
            can add it to the blacklist.
          </Title>
          <TextArea
            value={linkContent}
            onChange={(e) => setLinkContent(e.target.value)}
            rows={8}
            placeholder="Enter the original links or Auto-Share Links"
            style={{
              marginTop: token.margin,
              width: isMobile ? "100%" : "640px",
            }}
          />
        </Space>
      </Paragraph>
      <Button
        type="primary"
        onClick={handleAddToBlacklist}
        disabled={links.length > 0}
      >
        Add to Blacklist
      </Button>
      <BlackListTable />
    </>
  );
};

export default SharedLinksBlacklist;
