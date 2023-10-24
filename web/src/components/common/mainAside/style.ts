import styled from "styled-components";
import { Button, Layout } from "antd";

const { Sider } = Layout;

export const LogoWrapper = styled.div`
  display: flex;
  align-items: center;
  padding: 12px 8px;
`;

export const LogoImg = styled.img`
  width: 40px;
  height: 40px;
  object-fit: fill;
`;

export const LogoTitle = styled.p`
  font-size: 16px;
  font-weight: 700;
  margin-left: 4px;
  overflow: hidden;
`;

export const MainLayoutAside = styled(Sider)`
  width: 208px;
  height: 100%;
  background: var(--ks-bg);
`;

export const MenuController = styled.div`
  display: flex;
  align-items: center;
  padding: 0 16px;
  height: 40px;
  box-sizing: border-box;
  border-top: 0.5px solid transparent;
`;

export const MenuControllerButton = styled(Button)`
  font-size: 16px;
`;
