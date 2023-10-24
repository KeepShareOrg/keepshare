import { Button, Divider, Space, Typography, message, theme } from "antd";
import {
  LockOutlined,
  InfoCircleOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
} from "@ant-design/icons";
import { RewardStatistic } from "./style";
import { useEffect, useState } from "react";
import { getPikPakHostInfo } from "@/api/pikpak";
import useStore from "@/store";

const { Text } = Typography;

const General = () => {
  const { token } = theme.useToken();
  const [info, setInfo] = useStore((state) => [
    state.pikPakInfo,
    state.setPikPakInfo,
  ]);

  useEffect(() => {
    getPikPakHostInfo().then(({ data, error }) => {
      if (error) {
        message.error(error.message);
        return;
      }
      data && setInfo(data);
    });
  }, []);

  const [passwordVisible, setPasswordVisible] = useState(false);
  return (
    <>
      <Space align="start" wrap>
        <Space style={{ width: "200px", marginRight: token.marginLG }}>
          <Text style={{ color: token.colorTextSecondary }}>
            Master Account
          </Text>
        </Space>
        <Space
          direction="vertical"
          style={{ minWidth: "300px", marginRight: token.marginLG }}
        >
          <Text>Login E-mail</Text>
          <Space>
            <Text copyable strong>
              {info.master?.email || "-"}
            </Text>
          </Space>
        </Space>
        <Space direction="vertical">
          <Text>Password</Text>
          <Space>
            <Text copyable={passwordVisible}>
              {passwordVisible ? info.master?.password : "........"}
            </Text>
            {passwordVisible ? (
              <EyeOutlined onClick={() => setPasswordVisible(false)} />
            ) : (
              <EyeInvisibleOutlined onClick={() => setPasswordVisible(true)} />
            )}
          </Space>
          <Button icon={<LockOutlined style={{ color: token.colorPrimary }} />}>
            Modify Password
          </Button>
        </Space>
      </Space>
      <Divider />
      <Space align="start" wrap>
        <Space style={{ width: "200px", marginRight: token.marginLG }}>
          <Text style={{ color: token.colorTextSecondary }}>
            Master Account
          </Text>
        </Space>
        <Space direction="vertical">
          <RewardStatistic
            title="Account Total Rewards (USD)"
            value={info.revenue}
            color={token.colorText}
            style={{ color: token.colorText }}
            valueStyle={{ fontSize: token.fontSizeHeading2 }}
          />
          <Button
            href="https://mypikpak.com/referral/"
            target="_blank"
            style={{ marginTop: token.margin }}
            type="primary"
          >
            Go to PikPak Referral Program Pro
          </Button>
          <Space>
            <InfoCircleOutlined style={{ color: token.colorTextSecondary }} />
            <Text style={{ color: token.colorTextSecondary }}>
              KeepShare currently does not support direct cash withdrawal,
              please go to the official PikPak referral program page to withdraw
              income
            </Text>
          </Space>
        </Space>
      </Space>
    </>
  );
};

export default General;
