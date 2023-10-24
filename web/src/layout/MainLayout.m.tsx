import { Drawer, Layout, theme } from "antd";
import { Outlet } from "react-router-dom";
import { Background, MainLayoutHeader } from "./style";
import MainHeader from "@/components/common/mainHeader/MainHeader";
import MainAside from "@/components/common/mainAside/MainAside";
import useStore from "@/store";

const MainLayout = () => {
  const { Content } = Layout;
  const { token } = theme.useToken();

  const [isMobile, showMenuDrawer, setShowMenuDrawer] = useStore((state) => [
    state.isMobile,
    state.showMenuDrawer,
    state.setShowMenuDrawer,
  ]);

  return (
    <Background>
      <Layout>
        <MainLayoutHeader
          style={{
            background: token.colorBgContainer,
            paddingInline: isMobile ? token.margin : token.paddingXL,
          }}
        >
          <MainHeader />
        </MainLayoutHeader>
        <Content
          style={{
            height: "calc(100vh - 64px)",
            overflowY: "scroll",
            paddingInline: isMobile ? token.margin : token.paddingXL,
            background: token.colorBgContainer,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
      <Drawer
        placement="left"
        width={200}
        open={showMenuDrawer}
        onClose={() => setShowMenuDrawer(!showMenuDrawer)}
        headerStyle={{ display: "none" }}
        bodyStyle={{ padding: 0 }}
      >
        <MainAside />
      </Drawer>
    </Background>
  );
};

export default MainLayout;
