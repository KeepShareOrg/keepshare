import { useState, useRef } from "react";
import { LockOutlined, MailOutlined } from "@ant-design/icons";
import { Button, Divider, Space, Tabs, Typography, theme } from "antd";
import debounce from "lodash-es/debounce";

import { sendVerifyEmail } from "@/api/account";
import useStore from "@/store";

import ModifyEmailModal from "./ModifyEmailModal";
import ModifyPasswordModal from "./ModifyPasswordModal";
import { TextLink } from "./style";

const { Text } = Typography;

const Settings = () => {
  const [openModifyEmailModal, setOpenModifyEmailModal] = useState(false);
  const [openModifyPasswordModal, setOpenModifyPasswordModal] = useState(false);
  const [sentEmailMessage, setSentEmailMessage] = useState("");
  const [sendingEmail, setSendingEmail] = useState(false);
  const { token } = theme.useToken();
  const hasSentEmail = useRef(false);

  const userInfo = useStore((state) => state.userInfo);

  const handleOpenModifyEmailModal = () => {
    setOpenModifyEmailModal(true);
  };

  const handleOpenModifyPasswordModal = () => {
    setOpenModifyPasswordModal(true);
  };

  const handleSendVerifyEmail = debounce(async () => {
    if (hasSentEmail.current) return;
    hasSentEmail.current = true;

    setSendingEmail(true);
    const result = await sendVerifyEmail();
    setSendingEmail(false);

    if (result.error) {
      setSentEmailMessage("Email sending failed, please try again later.");
      return;
    }

    setSentEmailMessage("Email has been sent, please check it.");
  }, 300);

  const handleModifyEmailModalClose = () => {
    setOpenModifyEmailModal(false);
  };

  const handleModifyPasswordModalClose = () => {
    setOpenModifyPasswordModal(false);
  };

  return (
    <>
      <Tabs items={[{ key: "account", label: "Account" }]} />
      <Space align="start" wrap>
        <Space style={{ width: "200px" }}>
          <Text style={{ color: token.colorTextSecondary }}>Email</Text>
        </Space>
        <Space direction="vertical">
          <Text>{userInfo.email}</Text>
          <Button
            icon={<MailOutlined style={{ color: token.colorPrimary }} />}
            onClick={handleOpenModifyEmailModal}
          >
            Modify Email
          </Button>
          {userInfo.email_verified ? null : (
            <Space direction="vertical" align="start">
              <Text>
                You can reset your password if you forget it or get necessary
                notifications if you
                <TextLink
                  padding={token.paddingXS}
                  color={token.colorPrimaryHover}
                  loading={sendingEmail}
                  type="link"
                  disabled={Boolean(sentEmailMessage)}
                  onClick={handleSendVerifyEmail}
                >
                  {sendingEmail ? "Email Sending..." : "Verify Your Email."}
                </TextLink>
              </Text>
              {sentEmailMessage && (
                <Text style={{ fontWeight: token.fontWeightStrong }}>
                  {sentEmailMessage}
                </Text>
              )}
            </Space>
          )}
        </Space>
      </Space>
      <Divider />
      <Space align="start" wrap>
        <Space style={{ width: "200px" }}>
          <Text style={{ color: token.colorTextSecondary }}>Password</Text>
        </Space>
        <Space direction="vertical">
          <Text>···········</Text>
          <Button
            icon={<LockOutlined style={{ color: token.colorPrimary }} />}
            onClick={handleOpenModifyPasswordModal}
          >
            Modify Password
          </Button>
        </Space>
      </Space>
      <ModifyEmailModal
        open={openModifyEmailModal}
        onClose={handleModifyEmailModalClose}
      />
      <ModifyPasswordModal
        open={openModifyPasswordModal}
        onClose={handleModifyPasswordModalClose}
      />
    </>
  );
};

export default Settings;
