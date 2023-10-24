import type { GlobalToken } from "antd";
import styled from "styled-components";

export const AutoShareTemplate = styled.div`
  display: flex;
`;

export const ShareLinkBox = styled.div`
  display: flex;
  align-items: center;
  padding: 10px 32px;
  height: 32px;
  border-radius: 4px 0 0 4px;
  background: var(--ks-cyan);
  color: #fff;
`;

export const URLEncodedBox = styled.div`
  display: flex;
  align-items: center;
  height: 32px;
  border-radius: 0 4px 4px 0;
  color: #fff;
`;

export const EduBannerImg = styled.img`
  width: 480px;
  height: 200px;
  object-fit: fill;
`;

export const MobileEduBannerImg = styled.img`
  margin-top: 12px;
  max-width: 100%;
  object-fit: fill;
`;

export const TemplateDescribe = styled.div<{
  token?: GlobalToken;
  color?: React.CSSProperties["backgroundColor"];
}>`
  position: relative;
  padding-left: 16px;
  font-size: 12px;
  color: ${({ token }) => token?.colorTextSecondary};

  &::before {
    content: "";
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    display: block;
    width: 6px;
    height: 6px;
    border-radius: 1px;
    background-color: ${({ color }) => color || "var(--ks-cyan)"};
  }
`;
