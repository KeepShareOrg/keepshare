import {
  queryResetPasswordTaskStatus,
  resetPassword,
  ResetPasswordRequest,
} from "@/api/pikpak";
import {
  Modal,
  Space,
  Form,
  Input,
  theme,
  Checkbox,
  Typography,
  message,
} from "antd";
import { useState } from "react";
const { Text } = Typography;

interface ComponentProps {
  visible: boolean;
  toggleVisible: (visible: boolean) => void;
  refreshInfo: VoidFunction;
}
const ModifyPasswordModal = ({
  visible,
  toggleVisible,
  refreshInfo,
}: ComponentProps) => {
  const { token } = theme.useToken();
  const [form] = Form.useForm();

  const [loading, setLoading] = useState(false);
  const handleConfirmPassword = async (
    data: ResetPasswordRequest & { repeat_password: string },
  ) => {
    if (data.new_password === "" || data.repeat_password === "") {
      message.error("Please enter the new password twice");
      return;
    }
    if (data.new_password !== data.repeat_password) {
      message.error("The passwords entered twice are inconsistent");
      return;
    }
    setLoading(true);
    const { data: resetResp, error: resetErr } = await resetPassword(data);
    if (resetErr?.error || !resetResp?.task_id) {
      message.error(resetErr?.message || "Modify password error");
      setLoading(false);
      return;
    }

    let pullCount = 0;
    let timer = setInterval(async () => {
      const done = () => {
        clearInterval(timer);
        setLoading(false);
        form.resetFields();
      };

      const { data: queryResp, error: queryErr } =
        await queryResetPasswordTaskStatus(resetResp?.task_id);
      console.log(queryResp);
      if (queryErr?.error || queryResp?.status == "ERROR") {
        message.error(resetErr?.message || "Modify password error.");
        done();
        return;
      }

      if (queryResp?.status === "DONE") {
        message.success("The password was modified successfully.");
        done();
        refreshInfo();
        toggleVisible(false);
        return;
      }

      pullCount++;
      if (pullCount > 30) {
        done();
      }
    }, 1000);
  };

  return (
    <Modal
      title="Modify  PikPak Master Account Password"
      centered
      open={visible}
      maskClosable={false}
      keyboard={false}
      cancelText="Cancel"
      okText="Confirm"
      confirmLoading={loading}
      onCancel={() => toggleVisible(false)}
      okButtonProps={{ onClick: form.submit }}
      cancelButtonProps={{ onClick: () => toggleVisible(false) }}
    >
      <Form
        form={form}
        layout="vertical"
        style={{ marginTop: token.marginXL }}
        onFinish={handleConfirmPassword}
      >
        <Form.Item label="New Password" name="new_password" initialValue={""}>
          <Input placeholder="Please enter password" size="large" />
        </Form.Item>
        <Form.Item
          label="New Password(Repeat)"
          name="repeat_password"
          initialValue={""}
        >
          <Input placeholder="Please repeat password" size="large" />
        </Form.Item>
        <Form.Item
          name="save_password"
          valuePropName="checked"
          initialValue={true}
          style={{ marginBottom: 0 }}
        >
          <Checkbox>Save your new password in KeepShare</Checkbox>
        </Form.Item>
        <Space style={{ marginLeft: token.marginLG }}>
          <Text style={{ color: token.colorTextSecondary }}>
            If checked, KeepShare can automatically refresh the PikPak account
            login state with your password when it expires. You can also prevent
            KeepShare from saving your password, and KeepShare will only
            maintain the current login state as long as possible, but will not
            record your password in any form. You can review this in our open
            source code.
          </Text>
        </Space>
      </Form>
    </Modal>
  );
};

export default ModifyPasswordModal;
