import { Typography, message, theme } from "antd";
import { Background, ContentWrapper, LogoPng } from "./style";
import LogoIcon from "@/assets/images/logo-with-text.png";
import { useEffect, useState } from "react";
import {
  SharedLinkStatus,
  getLinkInfoFromWhatsLink,
  getSharedLinkInfo,
} from "@/api/link";
import { useSearchParams } from "react-router-dom";
import { getSupportLanguage } from "@/util";
import useStore from "@/store";
import { Trans, useTranslation } from "react-i18next";
import WhatsLinkInfo, { type WhatsLinkFileInfo } from "./WhatsLinkInfo";

const { Text, Link } = Typography;

const WhatsLinkStatus = () => {
  const { i18n } = useTranslation();

  const { token } = theme.useToken();

  const [fileInfo, setFileInfo] = useState<WhatsLinkFileInfo>({});

  const [params] = useSearchParams();

  const [requestId, setRequestId] = useState("");
  const setThemeMode = useStore((state) => state.setThemeMode);
  // status page keep light mode
  useEffect(() => {
    setThemeMode("light");
    setRequestId(params.get("request_id") || "");
  }, []);

  const [status, setStatus] = useState<SharedLinkStatus>("PENDING");

  useEffect(() => {
    const autoId = params.get("id") || "";
    if (!/^\d+$/i.test(autoId)) {
      return;
    }

    // Get data from keepshare server, but the data may be incomplete (returned from PikPak)
    getSharedLinkInfo(autoId, requestId, true).then(
      ({ data: newFileInfo, error }) => {
        if (newFileInfo) {
          const newStatus = newFileInfo.state;
          if (status === newStatus) {
            return;
          }
          setStatus(newStatus);
          setFileInfo(Object.assign({}, fileInfo, newFileInfo));
        }

        // Get data from whatsLink website
        if (!error) {
          getLinkInfoFromWhatsLink(newFileInfo?.original_link || "")
            .then(({ data, error }) => {
              try {
                if (error) {
                  return;
                }
                setFileInfo(
                  Object.assign({}, newFileInfo, {
                    title: fileInfo?.title || data?.name,
                    size: fileInfo?.size || data?.size,
                    screenshot: data?.screenshots?.[0]?.screenshot,
                    screenshots: data?.screenshots,
                    fileType: data?.file_type || "unknown",
                  }),
                );
              } catch (e) {
                console.error("get link info error: ", e);
              }
            })
            .catch((err) => {
              console.warn("get link info from whatslink error: ", err);
            });
        }

        error && message.error(error.message);
      },
    );
  }, []);

  useEffect(() => {
    i18n.changeLanguage(getSupportLanguage());
  }, []);

  return (
    <Background>
      <Link href="/">
        <LogoPng src={LogoIcon} />
      </Link>
      <ContentWrapper>
        <WhatsLinkInfo fileInfo={fileInfo} />
        <Text
          style={{
            display: "block",
            width: '100%',
            marginTop: "auto",
            color: token.colorTextTertiary,
            marginBottom: token.marginLG,
            textAlign: "center",
          }}
        >
        <Trans i18nKey="5cHad1gLjCn3u0sZhKNaU"
          components={[
            <Link
              underline
              href="https://whatslink.info"
              style={{ color: token.colorPrimary }}
            >
              whatslink.info
            </Link>,
            <Link
              underline
              href="https://github.com/KeepShareOrg/keepshare"
              style={{ color: token.colorPrimary }}
            >
              <Trans i18nKey="elUJuA04qeJah2ebplzP0"></Trans>
            </Link>,
          ]}
        />
      </Text>
      </ContentWrapper>
    </Background>
  );
};

export default WhatsLinkStatus;
