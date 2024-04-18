import {
  confirmMasterAccountPassword,
  ConfirmPasswordRequest,
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
const { Text } = Typography;

interface ComponentProps {
  visible: boolean;
  toggleVisible: (visible: boolean) => void;
}
const EnterPasswordModal = ({ visible, toggleVisible }: ComponentProps) => {
  const { token } = theme.useToken();
  const [form] = Form.useForm();

  const handleConfirmPassword = async (data: ConfirmPasswordRequest) => {
    const { error } = await confirmMasterAccountPassword(data);
    if (error?.error) {
      message.error(error?.message || "error");
    } else {
      message.success("The login refresh was successful.");
      toggleVisible(false);
    }
  };

  return (
    <Modal
      title="Enter  PikPak Master Account Password"
      centered
      open={visible}
      maskClosable={false}
      keyboard={false}
      cancelText="Cancel"
      okText="Confirm"
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
        <Form.Item label="Enter password" name="password" initialValue={""}>
          <Input placeholder="Please enter password" size="large" />
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

export default EnterPasswordModal;
