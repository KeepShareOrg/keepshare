import { Layout } from "antd";
import styled from "styled-components";

const { Content, Header } = Layout;

export const Background = styled(Layout)`
  background: var(--ks-bg);
  min-height: 100vh;
`;

export const ContentWrapper = styled(Content)`
  display: grid;
  place-items: center;
`;

export const MainLayoutHeader = styled(Header)`
  display: flex;
  align-items: center;
  justify-content: flex-end;
  height: 64px;
`;
