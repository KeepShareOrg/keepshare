import { Space, Typography, message, theme } from "antd";
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
import Loading from "./Loading";
import LinkInfo, { type LinkFileInfo } from "./LinkInfo";

const { Title, Text, Link } = Typography;

const useStatusDescribeText = (
  status: SharedLinkStatus,
  holder: JSX.Element,
) => {
  const { t } = useTranslation();

  if (["UNKNOWN", "DELETED", "NOT_FOUND", "BLOCKED"].includes(status)) {
    return {
      title: t("mEjDyHbG9xiu_6NlaegOn"),
      subtitle: (
        <Text>
          <Trans i18nKey="dW_5y60qwkDKvThMDiFl" components={[holder]}></Trans>
        </Text>
      ),
    };
  }

  return {
    title: t("2D0jDl0qeMSqvV0Ly6Iyd"),
    subtitle: (
      <Text>
        <Trans i18nKey="jgAl1qdprLfErnmioSgOu" components={[holder]}></Trans>
      </Text>
    ),
  };
};

const SharedStatus = () => {
  const { t, i18n } = useTranslation();

  const { token } = theme.useToken();

  const [fileInfo, setFileInfo] = useState<LinkFileInfo>({});

  const [params] = useSearchParams();

  const [requestId, setRequestId] = useState("");
  const setThemeMode = useStore((state) => state.setThemeMode);
  // status page keep light mode
  useEffect(() => {
    setThemeMode("light");
    setRequestId(params.get("request_id") || "");
  }, []);

  const { original_link: link } = fileInfo;

  const isMobile = useStore((state) => state.isMobile);

  const remoteDownload = (
    <Link
      href={`https://mypikpak.com/drive/url-checker?url=${window.encodeURIComponent(
        link || "",
      )}`}
      style={{ color: token.colorPrimary }}
    >
      {t("hDvGl13AlFfsLIi2jQ3xP")}
    </Link>
  );
  const [status, setStatus] = useState<SharedLinkStatus>("PENDING");
  const { title, subtitle } = useStatusDescribeText(status, remoteDownload);

  const MAX_LOOP_TIMES = 20;
  const [loopTimes, setLoopTimes] = useState(0);
  useEffect(() => {
    const autoId = params.get("id") || "";
    if (!/^\d+$/i.test(autoId)) {
      return;
    }

    const isLoopEnd = loopTimes > MAX_LOOP_TIMES;
    // Get data from keepshare server, but the data may be incomplete (returned from PikPak)
    getSharedLinkInfo(autoId, requestId, isLoopEnd).then(
      ({ data: fileInfo, error }) => {
        if (fileInfo) {
          if (loopTimes <= MAX_LOOP_TIMES) {
            const timer = setTimeout(() => {
              setLoopTimes(loopTimes + 1);
              timer && clearTimeout(timer);
            }, 2000);
          }

          const newStatus = fileInfo.state;
          if (status === newStatus) {
            return;
          }

          const hostSharedLink = fileInfo.host_shared_link;
          if (newStatus === "OK" && hostSharedLink) {
            location.href = hostSharedLink;
          } else {
            setStatus(newStatus);
            setFileInfo(fileInfo);
          }

          if (loopTimes === MAX_LOOP_TIMES) {
            setStatus("CREATED");
          }
        }

        // Get data from whatsLink website
        if (!error) {
          getLinkInfoFromWhatsLink(fileInfo?.original_link || "")
            .then(({ data, error }) => {
              if (error) {
                return;
              }
              setFileInfo(
                Object.assign({}, fileInfo, {
                  title: fileInfo?.title || data?.name,
                  size: fileInfo?.size || data?.size,
                  screenshot: data?.screenshots[0]?.screenshot,
                }),
              );
            })
            .catch(() => {});
        }

        error && message.error(error.message);
      },
    );
  }, [loopTimes]);

  useEffect(() => {
    i18n.changeLanguage(getSupportLanguage());
  }, []);

  return (
    <Background>
      <Link href="/">
        <LogoPng src={LogoIcon} />
      </Link>
      {loopTimes < 5 && params.get("st") !== "1" ? (
        <Loading>
          <ContentWrapper style={{ minHeight: "auto" }}>
            <LinkInfo
              fileInfo={fileInfo}
              visibleBlocks={["banner", "filename"]}
            />
          </ContentWrapper>
        </Loading>
      ) : (
        <ContentWrapper>
          <LinkInfo fileInfo={fileInfo} />
          <Space
            align={isMobile ? "center" : "start"}
            style={{ marginTop: "56px", maxWidth: "660px" }}
            direction={isMobile ? "vertical" : "horizontal"}
            size={20}
          >
            <Title level={3} style={{ lineHeight: "1em" }}>
              {title}
            </Title>
            <Text style={{ display: "inline-block", maxWidth: "440px" }}>
              {subtitle}
            </Text>
          </Space>
          <Text
            style={{
              color: token.colorTextTertiary,
              marginTop: "auto",
              marginBottom: token.marginLG,
              textAlign: "center",
            }}
          >
            <Trans
              i18nKey="3BuCl0v1UfHeRbDqnuh0N"
              components={[
                <Link
                  underline
                  href="https://whatslink.info"
                  style={{ color: token.colorTextTertiary }}
                >
                  whatslink.info
                </Link>,
              ]}
            ></Trans>
          </Text>
        </ContentWrapper>
      )}
    </Background>
  );
};

export default SharedStatus;
