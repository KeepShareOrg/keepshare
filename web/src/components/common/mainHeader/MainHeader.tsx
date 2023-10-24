import { Space, Switch, theme } from "antd";
import Avatar from "./Avatar";
import LogoPng from "@/assets/images/logo.png";
import useStore from "@/store";
import { MenuController, MenuControllerButton } from "../mainAside/style";
import { MenuFoldOutlined, MenuUnfoldOutlined } from "@ant-design/icons";

const MainHeader = () => {
  const { token } = theme.useToken();
  const [themeMode, setThemeMode, isMobile, showMenuDrawer, setShowMenuDrawer] =
    useStore((state) => [
      state.themeMode,
      state.setThemeMode,
      state.isMobile,
      state.showMenuDrawer,
      state.setShowMenuDrawer,
    ]);

  const handleChangeTheme = () => {
    setThemeMode(themeMode === "dark" ? "light" : "dark");
  };

  return (
    <>
      {isMobile && (
        <Space style={{ marginRight: "auto" }} align="center">
          <img
            src={LogoPng}
            alt="logo"
            style={{
              display: "block",
              objectFit: "fill",
              width: token.sizeXL,
            }}
          />
          <MenuController
            style={{
              borderTopColor: token.colorBorder,
              borderColor: token.colorBorder,
              padding: 0,
            }}
          >
            <MenuControllerButton
              type="text"
              icon={
                showMenuDrawer ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />
              }
              onClick={() => setShowMenuDrawer(!showMenuDrawer)}
            />
          </MenuController>
        </Space>
      )}
      <p style={{ marginRight: token.marginXS }}>
        {themeMode === "dark" ? "Dark Mode" : "Light Mode"}
      </p>
      <Switch
        checked={themeMode === "dark"}
        onChange={handleChangeTheme}
        style={{ marginRight: token.marginMD }}
      />
      <Avatar />
    </>
  );
};

export default MainHeader;
