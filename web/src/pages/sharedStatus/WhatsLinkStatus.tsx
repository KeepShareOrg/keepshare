import { Typography, message, theme } from "antd";
import { Background, ContentWrapper, LogoPng } from "./style";
import LogoIcon from "@/assets/images/logo-with-text.png";
import { useEffect, useState } from "react";
import {
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

  const requestId = params.get("request_id") || "";
  const setThemeMode = useStore((state) => state.setThemeMode);
  // status page keep light mode
  useEffect(() => {
    setThemeMode("light");
  }, []);

  useEffect(() => {
    const autoId = params.get("id") || "";
    if (!/^\d+$/i.test(autoId)) {
      return;
    }

    let active = true;
    getSharedLinkInfo(autoId, requestId, true, true).then(
      ({ data: newFileInfo, error }) => {
        if (!active) {
          return;
        }
        if (newFileInfo) {
          setFileInfo(newFileInfo);
        }

        // Get data from whatsLink website
        if (!error && newFileInfo?.original_link) {
          getLinkInfoFromWhatsLink(newFileInfo.original_link)
            .then(({ data, error }) => {
              try {
                if (!active || error) {
                  return;
                }
                setFileInfo(
                  (current) => ({
                    ...current,
                    title: current.title || data?.name,
                    size: current.size || data?.size,
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
    return () => {
      active = false;
    };
  }, [params, requestId]);

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
