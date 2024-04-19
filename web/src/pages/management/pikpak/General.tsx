import { Button, Divider, Space, Typography, message, theme } from "antd";
import {
  LockOutlined,
  InfoCircleOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
  ExclamationCircleFilled,
  CheckCircleFilled,
} from "@ant-design/icons";
import { RewardStatistic } from "./style";
import { useEffect, useState } from "react";
import {
  getPikPakHostInfo,
  getPikPakMasterAccountLoginStatus,
} from "@/api/pikpak";
import useStore from "@/store";
import ModifyMasterPasswordModal from "@/components/management/ModifyMasterPasswordModal";
import EnterPasswordModal from "@/components/management/EnterPasswordModal";

const { Text, Paragraph } = Typography;

const General = () => {
  const { token } = theme.useToken();
  const [info, setInfo] = useStore((state) => [
    state.pikPakInfo,
    state.setPikPakInfo,
  ]);
  const [loginValidStatus, setLoginValidStatus] = useState(false);

  const init = () => {
    getPikPakHostInfo().then(({ data, error }) => {
      if (error) {
        message.error(error.message);
        return;
      }
      data && setInfo(data);
    });

    getPikPakMasterAccountLoginStatus().then(({ data, error }) => {
      if (!error) {
        setLoginValidStatus(data?.status === "valid");
      }
    });
  };

  useEffect(() => init(), []);

  const [passwordVisible, setPasswordVisible] = useState(false);
  const [modifyPwdModalVisible, setModifyPwdModalVisible] = useState(false);
  const [enterPwdModalVisible, setEnterPwdModalVisible] = useState(false);
  return (
    <>
      <Space align="start" wrap>
        <Space style={{ width: "200px", marginRight: token.marginLG }}>
          <Text style={{ color: token.colorTextSecondary }}>
            Master Account
          </Text>
        </Space>
        <Space direction="vertical">
          <Space direction="vertical">
            <Text>Login E-mail</Text>
            <Space>
              <Text copyable strong>
                {info.master?.email || "-"}
              </Text>
            </Space>
          </Space>
          <Space direction="vertical" style={{ marginTop: token.marginXL }}>
            <Text>Password</Text>
            {info.master?.password ? (
              <Space>
                <Text copyable={passwordVisible}>
                  {passwordVisible ? info.master?.password : "........"}
                </Text>
                {passwordVisible ? (
                  <EyeOutlined onClick={() => setPasswordVisible(false)} />
                ) : (
                  <EyeInvisibleOutlined
                    onClick={() => setPasswordVisible(true)}
                  />
                )}
              </Space>
            ) : (
              <div>Not Saved</div>
            )}
            <Button
              icon={<LockOutlined style={{ color: token.colorPrimary }} />}
              onClick={() => setModifyPwdModalVisible(true)}
            >
              {info.master?.password ? "Modify Password" : "Reset Password"}
            </Button>
          </Space>
          <Space direction="vertical" style={{ marginTop: token.marginXL }}>
            <Text>Login State</Text>
            <Space direction="vertical">
              {loginValidStatus ? (
                <>
                  <Space align="end">
                    <CheckCircleFilled
                      style={{ color: "#389E0D", fontSize: "18px" }}
                    />
                    <Text strong style={{ fontSize: "16px" }}>
                      Valid
                    </Text>
                  </Space>
                  <Space>
                    <Paragraph style={{ maxWidth: "900px" }}>
                      <Text>
                        KeepShare requires the login of your master PikPak
                        account to calculate your sharing earning. If you change
                        your password within PikPak, your login state will be
                        disabled and your newly created shares will not be
                        counted as earnings.
                      </Text>
                    </Paragraph>
                  </Space>
                </>
              ) : (
                <>
                  <Space align="end">
                    <ExclamationCircleFilled
                      style={{ color: "#CF1322", fontSize: "18px" }}
                    />
                    <Text strong style={{ fontSize: "16px" }}>
                      Invalid
                    </Text>
                  </Space>
                  <Space>
                    <Paragraph style={{ maxWidth: "900px" }}>
                      <Text>
                        Your PikPak account login is invalid, which will cause
                        you to lose earnings on your newly created shares, so
                        please enter the password of your current PikPak master
                        account to have KeepShare refresh your login state.
                      </Text>
                    </Paragraph>
                  </Space>
                </>
              )}
              {!loginValidStatus && (
                <Button
                  type="primary"
                  onClick={() => setEnterPwdModalVisible(true)}
                >
                  Enter Password
                </Button>
              )}
            </Space>
          </Space>
        </Space>
      </Space>
      <Divider />
      <Space align="start" wrap>
        <Space style={{ width: "200px", marginRight: token.marginLG }}>
          <Text style={{ color: token.colorTextSecondary }}>Rewards</Text>
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
              Please go to the PikPak official referral program pro page, log in
              with the above master account and then withdraw your reward money.
            </Text>
          </Space>
        </Space>
      </Space>

      <EnterPasswordModal
        visible={enterPwdModalVisible}
        toggleVisible={setEnterPwdModalVisible}
        refreshInfo={init}
      />
      <ModifyMasterPasswordModal
        title={
          info.master?.password
            ? "Modify PikPak Master Account Password"
            : "Reset PikPak Master Account Password"
        }
        visible={modifyPwdModalVisible}
        toggleVisible={setModifyPwdModalVisible}
        refreshInfo={init}
      />
    </>
  );
};

export default General;
