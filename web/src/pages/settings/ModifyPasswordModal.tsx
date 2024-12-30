import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Modal, Form, message } from "antd";

import { changeAccountPassword } from "@/api/account";
import { calcPasswordHash } from "@/util";
import { RoutePaths } from "@/router";
import useStore from "@/store";

import { StyledForm, PasswordInput } from "./style";

interface FieldType {
  password?: string;
  newPassword?: string;
  newPasswordRepeat?: string;
}

interface ModifyPasswordModalProps {
  open?: boolean;
  onClose?: () => void;
}

function ModifyPasswordModal(props: ModifyPasswordModalProps) {
  const { open } = props;
  const [confirmDisabled, setConfirmDisabled] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const signOut = useStore((state) => state.signOut);

  const handleClose = () => {
    props.onClose && props.onClose();
    form.resetFields();
  };

  const handleValuesChange = async () => {
    try {
      await form.validateFields({
        validateOnly: true,
      });
      setConfirmDisabled(false);
    } catch {
      setConfirmDisabled(true);
    }
  };

  const handleModifyPassword = async ({ password, newPassword }: FieldType) => {
    if (!password || !newPassword) {
      console.error("can not find form field.");
      return;
    }

    setIsLoading(true);
    const result = await changeAccountPassword({
      password_hash: calcPasswordHash(password.trim()),
      new_password_hash: calcPasswordHash(newPassword.trim()),
    });
    setIsLoading(false);

    if (result.error) {
      return message.error(result.error.message || "modify password fail");
    }

    message.success("modify password success");
    signOut();
    // navigate to login
    navigate(RoutePaths.Login);
  };

  return (
    <Modal
      title="Modify Password"
      okText="Confirm"
      open={open}
      onCancel={handleClose}
      onOk={() => form.submit()}
      okButtonProps={{
        disabled: confirmDisabled && !isLoading,
        loading: isLoading,
      }}
    >
      <StyledForm
        form={form}
        layout="vertical"
        onFinish={(values) =>
          handleModifyPassword(values as unknown as FieldType)
        }
        onValuesChange={handleValuesChange}
        validateTrigger="onBlur"
        autoComplete="off"
      >
        <Form.Item<FieldType>
          label="Enter Password"
          name="password"
          initialValue=""
          rules={[
            {
              required: true,
              type: "string",
              message: "password is required",
              transform: (v) => v.trim(),
            },
          ]}
        >
          <PasswordInput placeholder="password" />
        </Form.Item>
        <Form.Item<FieldType>
          label="New Password"
          name="newPassword"
          initialValue=""
          rules={[
            {
              required: true,
              type: "string",
              message: "password is required",
              transform: (v) => v.trim(),
            },
          ]}
        >
          <PasswordInput placeholder="password" />
        </Form.Item>
        <Form.Item<FieldType>
          label="New Password (Repeat)"
          name="newPasswordRepeat"
          initialValue=""
          rules={[
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (value !== getFieldValue("newPassword")) {
                  return Promise.reject(
                    new Error("new password is inconsistent"),
                  );
                }
                return Promise.resolve();
              },
            }),
          ]}
        >
          <PasswordInput placeholder="password" />
        </Form.Item>
      </StyledForm>
    </Modal>
  );
}

export default ModifyPasswordModal;
