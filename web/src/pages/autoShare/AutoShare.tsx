import { Space, Typography, theme, Divider, Steps } from "antd";
import {
  CopyOutlined,
  InfoCircleOutlined,
  CheckOutlined,
} from "@ant-design/icons";
import {
  AutoShareTemplate,
  EduBannerImg,
  MobileEduBannerImg,
  ShareLinkBox,
  TemplateDescribe,
  URLEncodedBox,
} from "./style";
import EduBanner from "@/assets/images/edu.png";
import useStore from "@/store";
import { useState } from "react";
import { Link } from "react-router-dom";
import { RoutePaths } from "@/router";
import { copyToClipboard } from "@/util";

const { Title, Text, Paragraph, Link: TextLink } = Typography;

/**
 * Renders the AutoShare component.
 *
 * @return {JSX.Element} The rendered AutoShare component.
 */
const AutoShare = () => {
  const { token } = theme.useToken();

  const [userInfo, isMobile] = useStore((state) => [
    state.userInfo,
    state.isMobile,
  ]);

  const keepShareLink = `${location.origin}/${userInfo.channel_id}/`;

  const [exampleLink, setExampleLink] = useState(
    "magnet:?xt=urn:btih:XSPQGEYHIR6CD4GTCYM6P4UDRMTMEOEJ",
  );

  const handleCopyKeepShareLink = () => {
    copyToClipboard(keepShareLink);
  };

  return (
    <>
      <Space style={{ flexWrap: isMobile ? "wrap" : "nowrap" }}>
        <Space direction="vertical">
          <Paragraph>
            <Title level={4}>
              Easily and automatically convert magnet or other download links
              into file hosting and sharing links based on your template.
            </Title>
            <Text>
              Combine the download link in the format of the following template
              into a keep sharing link, and post it on your website or social
              media. When someone opens this link for the first time, KeepShare
              will upload it remotely at the file hosting provider, create a
              share and then jump to this share. All completely automatic and at
              no cost, just wait for the rewards from your file hosting provider
              to arrive.
            </Text>
          </Paragraph>
          <Text>Auto Keep Sharing Link Template:</Text>
          <AutoShareTemplate>
            <ShareLinkBox style={{ padding: isMobile ? "10px" : "10px 32px" }}>
              <Text
                style={{ color: token.colorWhite, whiteSpace: "nowrap" }}
                onClick={handleCopyKeepShareLink}
                copyable={{
                  icon: [
                    <CopyOutlined
                      key="copy-icon"
                      style={{ color: token.colorWhite }}
                    />,
                    <CheckOutlined
                      key="copied-icon"
                      style={{ color: token.colorWhite }}
                    />,
                  ],
                  tooltips: ["copy", "copied!"],
                }}
              >
                {userInfo.user_id ? keepShareLink : ""}
              </Text>
            </ShareLinkBox>
            <URLEncodedBox
              style={{
                background: token.colorPrimary,
                whiteSpace: "nowrap",
                padding: isMobile ? "10px" : "10px 32px",
              }}
            >
              <Text style={{ color: token.colorWhite }}>URL-Encode</Text>
              {isMobile || (
                <Text style={{ color: token.colorWhite }}>(Download-Link)</Text>
              )}
            </URLEncodedBox>
          </AutoShareTemplate>
          <TemplateDescribe token={token}>
            Your Keep Sharing Link Prefix
          </TemplateDescribe>
          <TemplateDescribe token={token} color={token.colorPrimary}>
            URL-encoded download link
          </TemplateDescribe>
          <Paragraph>
            <Text>Javascript Code: </Text>
            <Text copyable={true}>
              {`'https://keepshare.org/${userInfo.channel_id}/' + encodeURIComponent(downloadURL)`}
            </Text>
          </Paragraph>
        </Space>
        {isMobile ? (
          <MobileEduBannerImg src={EduBanner} />
        ) : (
          <EduBannerImg src={EduBanner} />
        )}
      </Space>
      <Divider />
      <Paragraph>
        <Title level={4} style={{ marginBottom: token.marginMD }}>
          An example to understand how the Auto-Share Link Template is generated
        </Title>
        <Steps
          current={-1}
          direction="vertical"
          items={[
            {
              title: "For a magnet download link (feel free to modify)",
              description: (
                <Text
                  style={{ width: "600px", marginLeft: "12px" }}
                  editable={{ onChange: setExampleLink }}
                >
                  {exampleLink}
                </Text>
              ),
            },
            {
              title: "URL-Encode the link URL to",
              description: (
                <Text style={{ marginLeft: "12px" }}>
                  {window.encodeURIComponent(exampleLink)}
                </Text>
              ),
            },
            {
              title:
                "Combine your link prefix and the encoded URL according to the template, the keep sharing link is",
              description: (
                <TextLink
                  style={{ color: token.colorPrimaryHover, marginLeft: "12px" }}
                  href={`${keepShareLink}${window.encodeURIComponent(
                    exampleLink,
                  )}`}
                  target="_blank"
                >
                  {`${keepShareLink}${window.encodeURIComponent(exampleLink)}`}
                </TextLink>
              ),
            },
          ]}
        />
        <Paragraph>
          <Space
            style={{ alignItems: "flex-start", marginBottom: token.margin }}
          >
            <InfoCircleOutlined style={{ color: token.colorTextSecondary }} />
            <Text style={{ color: token.colorTextSecondary }}>
              No need to manually generate a keep sharing link here. You can
              directly combine the link string according to the template and
              publish it to your website or social media immediately.
            </Text>
          </Space>
          <Space>
            <Text>
              You can prevent certain links from generating Auto-Share links by
              <Link
                style={{
                  color: token.colorPrimaryHover,
                  marginLeft: token.marginXXS,
                }}
                to={`${RoutePaths.SharedLinks}?tab=link-blacklist`}
              >
                Adding Links to the Blacklist
              </Link>
            </Text>
          </Space>
          <Divider />
          <Paragraph>
            <Link
              style={{ color: token.colorPrimaryHover }}
              to={RoutePaths.SharedLinks}
            >
              View in Shared Links
            </Link>
          </Paragraph>
        </Paragraph>
      </Paragraph>
    </>
  );
};

export default AutoShare;
