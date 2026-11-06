import { Space, Typography, message, theme } from "antd";
import { Background, ContentWrapper, LogoPng } from "./style";
import LogoIcon from "@/assets/images/logo-with-text.png";
import { useEffect, useState } from "react";
import {
  SharedLinkStatus,
  getLinkInfoFromWhatsLink,
  getSharedLinkInfo,
} from "@/api/link";
import { useNavigate, useSearchParams } from "react-router-dom";
import { RoutePaths } from "@/router";
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

  const requestId = params.get("request_id") || "";
  const navigate = useNavigate();
  const setThemeMode = useStore((state) => state.setThemeMode);
  // status page keep light mode
  useEffect(() => {
    setThemeMode("light");
  }, []);

  const { original_link: link } = fileInfo;

  const isMobile = useStore((state) => state.isMobile);

  const remoteDownload = (
    <Link
      href={`https://mypikpak.com/drive/url-checker?url=${window.encodeURIComponent(
        link || "",
      )}`}
      target="_blank"
      style={{ color: token.colorPrimary }}
    >
      {t("hDvGl13AlFfsLIi2jQ3xP")}
    </Link>
  );
  const [status, setStatus] = useState<SharedLinkStatus>("PENDING");
  const { title, subtitle } = useStatusDescribeText(status, remoteDownload);

  useEffect(() => {
    if (!link) {
      return;
    }
    let active = true;
    getLinkInfoFromWhatsLink(link)
      .then(({ data, error }) => {
        if (!active || error || !data) {
          return;
        }
        setFileInfo((current) => ({
          ...current,
          title: current.title || data.name,
          size: current.size || data.size,
          screenshot: data.screenshots?.[0]?.screenshot,
          fileType: data.file_type || "unknown",
        }));
      })
      .catch((err) => {
        console.warn("get link info from whatslink error: ", err);
      });
    return () => {
      active = false;
    };
  }, [link]);

  const MAX_LOOP_TIMES = 20;
  const [loopTimes, setLoopTimes] = useState(0);
  useEffect(() => {
    const autoId = params.get("id") || "";
    if (!/^\d+$/i.test(autoId)) {
      return;
    }

    const isLoopEnd = loopTimes > MAX_LOOP_TIMES;
    let timer: ReturnType<typeof setTimeout> | null = null;
    let active = true;
    // Get data from keepshare server, but the data may be incomplete (returned from PikPak)
    getSharedLinkInfo(autoId, requestId, isLoopEnd).then(
      ({ data: newFileInfo, error }) => {
        if (!active) {
          return;
        }
        if (newFileInfo) {
          if (newFileInfo.resource_only) {
            active = false;
            timer && clearTimeout(timer);
            const search = new URLSearchParams({
              id: autoId,
              request_id: requestId,
            });
            navigate(
              { pathname: RoutePaths.WhatsLinkStatus, search: `?${search}` },
              { replace: true },
            );
            return;
          }
          if (loopTimes <= MAX_LOOP_TIMES) {
            timer = setTimeout(() => {
              timer = null;
              if (active) {
                setLoopTimes((times) => times + 1);
              }
            }, 2000);
          }

          const newStatus = newFileInfo.state;

          const hostSharedLink = newFileInfo.host_shared_link;
          if (newStatus === "OK" && hostSharedLink) {
            active = false;
            location.href = hostSharedLink;
            timer && clearTimeout(timer);
            return;
          } else {
            setStatus(newStatus);
            setFileInfo((current) => ({
              ...current,
              ...newFileInfo,
              title: newFileInfo.title || current.title,
              size: newFileInfo.size || current.size,
            }));
          }

          if (loopTimes === MAX_LOOP_TIMES) {
            setStatus("CREATED");
          }
        }

        error && message.error(error.message);
      },
    );
    return () => {
      active = false;
      timer && clearTimeout(timer);
    };
  }, [loopTimes, params, navigate, requestId]);

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
