import type {
  GetLinkInfoFromWhatsLinkResponse,
  SharedLinkInfo,
} from "@/api/link";
import UnknownIcon from "@/assets/images/file-unknown.png";
import { copyToClipboard, formatBytes } from "@/util";
import { Button, Space, message, theme, Typography } from "antd";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { ResourceLink, WslBannerImage, WslBannerWrapper, WslInfoWrapper } from "./style";
import { CopyOutlined, ArrowLeftOutlined } from "@ant-design/icons";
import linkPngUrl from "@/assets/images/icon-link.png";
import { getShareIcon } from "./LinkInfo";

const { Text } = Typography;

export type WhatsLinkFileInfo = Partial<SharedLinkInfo> & {
  screenshot?: string;
  fileType?: GetLinkInfoFromWhatsLinkResponse["file_type"];
};
export type LinkInfoBlock = "banner" | "filename" | "link";

interface LinkInfoInterface {
  fileInfo: WhatsLinkFileInfo;
  visibleBlocks?: LinkInfoBlock[];
}
const WhatsLinkInfo = ({ fileInfo, visibleBlocks }: LinkInfoInterface) => {
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
    <WslInfoWrapper>
      {
        blocks.includes("banner") &&
        Array.isArray(fileInfo.screenshots) && fileInfo.screenshots?.length > 0 ?
        (
          <WslBannerWrapper>
            {fileInfo.screenshots?.map(({ screenshot }, index) => (
              <WslBannerImage
                src={screenshot}
                key={index}
                alt="banner"
                style={{
                  height: "132px",
                }}
              />
            ))}
          </WslBannerWrapper>
        ) :  (
          <Space
            direction="vertical"
            align="center"
            style={{ marginTop: "auto" }}
          >
            <img src={shareIcon} style={{ width: "94px" }} alt="shareIcon" />
          </Space>
        )
      }

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

      <Space>
        <Text style={{ color: token.colorTextTertiary, fontSize: token.fontSizeSM }}>{size}</Text>
      </Space>

      {blocks.includes("link") && (
        <>
          <Space
            align="start"
            style={{ marginTop: token.marginLG, maxWidth: "680px" }}
          >
            <img
              src={linkPngUrl}
              alt="link"
              width="24"
              style={{ marginTop: "5px" }}
            />
            <ResourceLink href={link}>{link}</ResourceLink>
          </Space>
          <Space style={{ marginTop: token.margin }}>
            {
              document.referrer && (
                <Button
                  icon={<ArrowLeftOutlined />}
                  onClick={() => window.history.back()}
                >
                  {t('6Krkunr0j0BHryv1p8p7e')}
                </Button>
              )
            }
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
    </WslInfoWrapper>
  );
};

export default WhatsLinkInfo;
