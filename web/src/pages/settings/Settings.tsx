import { useState } from 'react';
import { LockOutlined, MailOutlined } from "@ant-design/icons";
import { Button, Divider, Space, Tabs, Typography, theme, message } from "antd";
import debounce from 'lodash-es/debounce';

import { sendVerifyEmail } from '@/api/account';
import useStore from "@/store";

import ModifyEmailModal from './ModifyEmailModal';
import ModifyPasswordModal from './ModifyPasswordModal';

const { Text, Link: TextLink } = Typography;

const Settings = () => {
  const [openModifyEmailModal, setOpenModifyEmailModal] = useState(false);
  const [openModifyPasswordModal, setOpenModifyPasswordModal] = useState(false);
  const { token } = theme.useToken();

  const userInfo = useStore((state) => state.userInfo);

  const handleOpenModifyEmailModal = () => {
    setOpenModifyEmailModal(true);
  };

  const handleOpenModifyPasswordModal = () => {
    setOpenModifyPasswordModal(true);
  };

  const handleSendVerifyEmail = debounce(async () => {
    const result = await sendVerifyEmail();

    if (result.error) {
      message.error(result.error.message || 'send fail');
      return;
    }

    message.success('send successfully, please check your email');
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
          <Button icon={<MailOutlined style={{ color: token.colorPrimary }} />} onClick={handleOpenModifyEmailModal}>
            Modify Email
          </Button>
          {userInfo.email_verified ? null : (
            <Space>
              <Text>
                You can reset your password if you forget it or get necessary notifications 
                if you
                <TextLink
                  style={{
                    color: token.colorPrimaryHover,
                    marginLeft: token.marginXS,
                  }}
                  onClick={handleSendVerifyEmail}
                >
                  Verify Your Email.
                </TextLink>
              </Text>
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
          <Button icon={<LockOutlined style={{ color: token.colorPrimary }} />} onClick={handleOpenModifyPasswordModal}>
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
