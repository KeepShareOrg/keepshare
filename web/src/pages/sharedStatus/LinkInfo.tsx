import type {
  GetLinkInfoFromWhatsLinkResponse,
  SharedLinkInfo,
} from "@/api/link";
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
import UnknownIcon from "@/assets/images/file-unknown.png";
import LinkPng from "@/assets/images/icon-link.png";
import { match } from "ts-pattern";

const { Text } = Typography;

const getShareIcon = async (
  fileType: GetLinkInfoFromWhatsLinkResponse["file_type"],
): Promise<string> => {
  try {
    const result = await match(fileType)
      .with("unknown", () => import("@/assets/images/file-unknown.png"))
      .with("folder", () => import("@/assets/images/file-folder.png"))
      .with("video", () => import("@/assets/images/file-video.png"))
      .with("text", () => import("@/assets/images/file-text.png"))
      .with("image", () => import("@/assets/images/file-image.png"))
      .with("audio", () => import("@/assets/images/file-audio.png"))
      .with("document", () => import("@/assets/images/file-document.png"))
      .with("archive", () => import("@/assets/images/file-archive.png"))
      .with("font", () => import("@/assets/images/file-font.png"))
      .exhaustive();
    return result.default || UnknownIcon;
  } catch {
    return UnknownIcon;
  }
};

export type LinkFileInfo = Partial<SharedLinkInfo> & {
  screenshot?: string;
  fileType?: GetLinkInfoFromWhatsLinkResponse["file_type"];
};
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

  const [shareIcon, setShareIcon] = useState(UnknownIcon);
  useEffect(() => {
    getShareIcon(fileInfo.fileType!).then(setShareIcon);
  }, [fileInfo]);

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
            <img src={shareIcon} style={{ width: "94px" }} alt="shareIcon" />
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
            style={{ marginTop: token.marginLG, maxWidth: "680px" }}
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
