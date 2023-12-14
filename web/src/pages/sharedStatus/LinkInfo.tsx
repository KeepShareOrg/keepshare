import { type SharedLinkInfo } from "@/api/link";
import { copyToClipboard, formatBytes } from "@/util";
import { Button, Space, message, theme, Typography } from "antd";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  BannerImage,
  BannerWrapper,
  ResourceLink,
  SharedInfoBox,
} from "./style";
import { CopyOutlined } from "@ant-design/icons";
import ShareIcon from "@/assets/images/prepare-status-banner.png";
import LinkPng from "@/assets/images/icon-link.png";

const { Text } = Typography;

export type LinkFileInfo = Partial<SharedLinkInfo> & { screenshot?: string };
export type LinkInfoBlock = "banner" | "filename" | "link";

interface LinkInfoInterface {
  fileInfo: LinkFileInfo;
  visibleBlocks?: LinkInfoBlock[];
}
const LinkInfo = ({ fileInfo, visibleBlocks }: LinkInfoInterface) => {
  const { t } = useTranslation();

  const { title: filename, size: storage } = fileInfo;

  const size = formatBytes((storage as number) || 0);
  const { token } = theme.useToken();
  const link = fileInfo.original_link;

  const handleCopyLink = () => {
    try {
      link && copyToClipboard(link);
      message.success(t("xKhHo2JwfdzWgJXiJ0GeI"));
    } catch {
      message.error(t("aiCd4EgbrLDu4cdLlBy"));
    }
  };

  const [blocks, setBlocks] = useState<LinkInfoBlock[]>([
    "banner",
    "filename",
    "link",
  ]);
  useEffect(() => {
    Array.isArray(visibleBlocks) && setBlocks(visibleBlocks);
  }, [blocks]);

  if (JSON.stringify(fileInfo) === "{}") {
    return <></>;
  }

  return (
    <>
      {blocks.includes("banner") &&
        (fileInfo.screenshot ? (
          <BannerWrapper style={{ marginInline: token.margin }}>
            {fileInfo.screenshot && (
              <BannerImage
                src={fileInfo.screenshot}
                alt="banner"
                style={{
                  width: "100%",
                  height: "100%",
                }}
              />
            )}
            <SharedInfoBox>
              {t("whMzAm8sGpQfOTqadiXu")} {size}
            </SharedInfoBox>
          </BannerWrapper>
        ) : (
          <Space
            direction="vertical"
            align="center"
            style={{ marginTop: "auto" }}
          >
            <img src={ShareIcon} style={{ width: "94px" }} alt="shareIcon" />
            {storage ? (
              <Text
                style={{
                  color: token.colorTextTertiary,
                  fontSize: token.fontSizeSM,
                }}
              >
                {t("whMzAm8sGpQfOTqadiXu")} {size}
              </Text>
            ) : null}
          </Space>
        ))}

      {blocks.includes("filename") && (
        <Text
          style={{
            maxWidth: "min(600px, 100vw)",
            marginTop: "12px",
            textAlign: "center",
            lineHeight: "1.4em",
          }}
        >
          {filename}
        </Text>
      )}

      {blocks.includes("link") && (
        <>
          <Space
            align="start"
            style={{ marginTop: token.marginLG, maxWidth: "660px" }}
          >
            <img
              src={LinkPng}
              alt="link"
              width="24"
              style={{ marginTop: "5px" }}
            />
            <ResourceLink href={link}>{link}</ResourceLink>
          </Space>
          <Space style={{ marginTop: token.margin }}>
            <Button
              type="primary"
              icon={<CopyOutlined />}
              onClick={handleCopyLink}
            >
              {t("fbWqi7mJuMCxEw3SwCf_0")}
            </Button>
          </Space>
        </>
      )}
    </>
  );
};

export default LinkInfo;
