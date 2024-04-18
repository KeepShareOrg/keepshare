import { getPikPakMasterAccountLoginStatus } from "@/api/pikpak";
import { ExclamationCircleFilled } from "@ant-design/icons";
import { Typography, Space, Modal } from "antd";
import { useEffect, useState } from "react";
import EnterPasswordModal from "../management/EnterPasswordModal";

const { Text, Title } = Typography;

const LoginWarningModal = () => {
  const [visible, setVisible] = useState(false);
  const [enterPwdModalVisible, setEnterPwdModalVisible] = useState(false);

  useEffect(() => {
    getPikPakMasterAccountLoginStatus().then(({ data }) => {
      setVisible(data?.status === "invalid");
    });
  }, []);

  const handleEnterPassword = () => {
    setVisible(false);
    setEnterPwdModalVisible(true);
  };

  return (
    <>
      <Modal
        centered
        open={visible}
        onCancel={() => setVisible(false)}
        maskClosable={false}
        keyboard={false}
        closable={false}
        cancelText="Not now"
        okText="Enter Password"
        okButtonProps={{ onClick: handleEnterPassword }}
      >
        <Space align="start">
          <Space style={{ marginRight: "8px" }}>
            <ExclamationCircleFilled
              style={{
                fontSize: "22px",
                marginTop: "4px",
                color: "#CF1322",
                alignSelf: "flex-start",
              }}
            />
          </Space>
          <Space direction="vertical">
            <Title style={{ fontSize: "16px" }}>
              The PikPak master account login is invalid, which prevents you
              from getting earning for newly created share link.
            </Title>
            <Text>
              Please enter the password of your current PikPak master account to
              refresh KeepShare's login state.
            </Text>
          </Space>
        </Space>
      </Modal>

      <EnterPasswordModal
        visible={enterPwdModalVisible}
        toggleVisible={setEnterPwdModalVisible}
      />
    </>
  );
};

export default LoginWarningModal;
