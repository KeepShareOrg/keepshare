import styled from "styled-components";
import backgroundSvg from "@/assets/images/cloud-bg.svg";
import { Typography } from "antd";

const { Link } = Typography;

export const Background = styled.div`
  min-width: 100vw;
  min-height: 100vh;
  box-sizing: border-box;
  background: url(${backgroundSvg}) no-repeat center/cover;
`;

export const LogoPng = styled.img`
  margin: 12px 24px;
  height: 64px;
`;

export const ContentWrapper = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 340px;
  min-height: calc(100vh - 88px);
  padding-inline: 16px;
`;

export const BannerWrapper = styled.div<{ poster?: string }>`
  position: relative;
  width: 100%;
  max-width: 340px;
  height: 216px;
  border-radius: 16px;
  margin-inline: 6px;
  box-sizing: border-box;
  overflow: hidden;
  margin-top: auto;
  background: ${(props) =>
    props.poster ? `url(${props.poster}) no-repeat center/cover` : "#000"};

  @media screen and (max-width: 768px) {
    width: calc(100vw - 32px);
    height: 200px;
  }
`;

export const BannerImage = styled.img`
  position: absolute;
  width: 120px;
  height: 120px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
`;

export const SharedInfoBox = styled.div`
  position: absolute;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  max-width: 450px;
  left: 8px;
  bottom: 8px;
  color: #fff;
  line-height: 1.4em;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 12px;
  background: rgba(68, 68, 68, 0.60);
`;

export const ResourceLink = styled(Link)`
  font-size: 20px;
  word-break: break-all;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  color: rgba(0, 0, 0, 0.85)!important;

  &:hover {
    text-decoration: underline!important;
  }
`  