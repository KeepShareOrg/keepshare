import MainMenu from "./MainMenu";
import {
  LogoWrapper,
  LogoImg,
  LogoTitle,
  MainLayoutAside,
  MenuController,
  MenuControllerButton,
} from "./style";
import Logo from "@/assets/images/logo.png";
import { Space, theme } from "antd";
import { MenuUnfoldOutlined, MenuFoldOutlined } from "@ant-design/icons";
import useStore from "@/store";

interface ComponentProps {
  collapsed: boolean;
}
export const AsideLogo = ({ collapsed }: ComponentProps) => {
  const { token } = theme.useToken();

  return (
    <LogoWrapper
      style={{
        justifyContent: collapsed ? "center" : "flex-start",
      }}
    >
      <LogoImg src={Logo} />
      <LogoTitle
        style={{
          width: collapsed ? "0" : "fit-content",
          color: token.colorText,
        }}
      >
        KeepShare
      </LogoTitle>
    </LogoWrapper>
  );
};

const MainAside = () => {
  const [collapsed, setCollapsed, isMobile] = useStore((state) => [
    state.asideCollapsed,
    state.setAsideCollapsed,
    state.isMobile,
  ]);
  const { token } = theme.useToken();

  return (
    <MainLayoutAside
      collapsed={collapsed}
      style={{ background: "var(--ks-bg)", height: "100vh" }}
    >
      <AsideLogo collapsed={collapsed} />
      <Space
        direction="vertical"
        size="middle"
        style={{
          height: "calc(100% - 64px)",
          display: "flex",
          justifyContent: "space-between",
        }}
      >
        <MainMenu />
        {isMobile || (
          <MenuController
            style={{
              borderTopColor: token.colorBorder,
              borderColor: token.colorBorder,
              marginTop: "auto",
            }}
          >
            <MenuControllerButton
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
            />
          </MenuController>
        )}
      </Space>
    </MainLayoutAside>
  );
};

export default MainAside;
