import { Row, Space, Typography } from "antd";
// import NotFoundPageBanner from "@/assets/images/404.png";

const { Text } = Typography;

const ErrorPage = () => {
  return (
    <Row justify="center" align="middle" style={{ height: "100vh" }}>
      <Space direction="vertical" style={{ alignItems: "center" }}>
        {/* <img src={NotFoundPageBanner} alt="error" style={{ width: "220px" }} /> */}
        <Text>The page is lost, take a break and try again~</Text>
      </Space>
    </Row>
  );
};

export default ErrorPage;
