import { Drawer, Layout, theme } from "antd";
import { Outlet } from "react-router-dom";
import { Background, MainLayoutHeader } from "./style";
import MainHeader from "@/components/common/mainHeader/MainHeader";
import MainAside, {
  ASIDE_WIDTH,
} from "@/components/common/mainAside/MainAside";
import useStore from "@/store";
import LoginWarningModal from "@/components/common/LoginWarningModal";

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
        width={ASIDE_WIDTH}
        open={showMenuDrawer}
        onClose={() => setShowMenuDrawer(!showMenuDrawer)}
        headerStyle={{ display: "none" }}
        bodyStyle={{ padding: 0 }}
      >
        <MainAside />
      </Drawer>

      <LoginWarningModal />
    </Background>
  );
};

export default MainLayout;
