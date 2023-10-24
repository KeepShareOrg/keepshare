import useStore from "@/store";
import { LockOutlined, MailOutlined } from "@ant-design/icons";
import { Button, Divider, Space, Tabs, Typography, theme } from "antd";

const { Text, Link: TextLink } = Typography;
const Settings = () => {
  const { token } = theme.useToken();

  const userInfo = useStore((state) => state.userInfo);

  return (
    <>
      <Tabs items={[{ key: "account", label: "Account" }]} />
      <Space align="start" wrap>
        <Space style={{ width: "200px" }}>
          <Text style={{ color: token.colorTextSecondary }}>Email</Text>
        </Space>
        <Space direction="vertical">
          <Text>{userInfo.email}</Text>
          <Button icon={<MailOutlined style={{ color: token.colorPrimary }} />}>
            Modify Email
          </Button>
          <Space>
            <Text>
              In order to avoid being unable to log in after you forget your
              password,
              <TextLink
                style={{
                  color: token.colorPrimaryHover,
                  marginLeft: token.marginXS,
                }}
              >
                Please Verify Your Email.
              </TextLink>
            </Text>
          </Space>
        </Space>
      </Space>
      <Divider />
      <Space align="start" wrap>
        <Space style={{ width: "200px" }}>
          <Text style={{ color: token.colorTextSecondary }}>Password</Text>
        </Space>
        <Space direction="vertical">
          <Text>···········</Text>
          <Button icon={<LockOutlined style={{ color: token.colorPrimary }} />}>
            Modify Password
          </Button>
        </Space>
      </Space>
    </>
  );
};

export default Settings;
