import { Layout, theme } from "antd";
import { Outlet } from "react-router-dom";
import { Background, MainLayoutHeader } from "./style";
import MainHeader from "@/components/common/mainHeader/MainHeader";
import MainAside from "@/components/common/mainAside/MainAside";

const MainLayout = () => {
  const { Content } = Layout;
  const { token } = theme.useToken();

  return (
    <Background>
      <MainAside />
      <Layout>
        <MainLayoutHeader style={{ background: token.colorBgContainer }}>
          <MainHeader />
        </MainLayoutHeader>
        <Content
          style={{
            height: "calc(100vh - 64px)",
            overflowY: "scroll",
            paddingInline: token.paddingXL,
            background: token.colorBgContainer,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Background>
  );
};

export default MainLayout;
